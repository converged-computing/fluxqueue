package controller

import (
	"context"
	"fmt"
	"os"

	pb "github.com/converged-computing/fluxion/pkg/fluxion-grpc"
	"github.com/converged-computing/fluxqueue/pkg/defaults"
	"github.com/converged-computing/fluxqueue/pkg/jgf"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	defaultClusterName = "cluster"
	controlPlaneLabel  = "node-role.kubernetes.io/control-plane"
)

// InitFluxion creates the in cluster resource graph
func (r *FluxJobReconciler) InitFluxion(ctx context.Context, policy string) error {

	// Last argument is a label for nodes, if we want
	err := r.createInClusterJGF(defaults.KubernetesJsonGraphFormat, "")
	if err != nil {
		return err
	}

	// This file needs to be written for GetResources to read later
	jgf, err := os.ReadFile(defaults.KubernetesJsonGraphFormat)
	if err != nil {
		return err
	}
	graph := string(jgf)
	// fmt.Println(graph)
	rlog.Info("match policy", "Policy", policy)
	request := &pb.InitRequest{Policy: policy, Jgf: graph}
	response, err := r.Fluxion.Init(ctx, request)
	if err != nil {
		return err
	}
	rlog.Info("⭐️ Init cluster status", "Status", response.Status)
	return nil
}

// CreateInClusterJGF creates the Json Graph Format from the Kubernetes API
func (r *FluxJobReconciler) createInClusterJGF(filename, skipLabel string) error {
	ctx := context.Background()
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println("Error getting InClusterConfig")
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error getting ClientSet: %s", err)
		return err
	}
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing nodes: %s", err)
		return err
	}

	// Create a Flux Json Graph Format (JGF) with all cluster nodes
	fluxgraph := jgf.NewFluxJGF()

	// Initialize the cluster. The top level of the graph is the cluster
	// This assumes fluxion is only serving one cluster.
	// previous comments indicate that we choose between the level
	// of a rack and a subnet. A rack doesn't make sense (the nodes could
	// be on multiple racks) so subnet is likely the right abstraction
	clusterNode, err := fluxgraph.InitCluster(defaultClusterName)
	if err != nil {
		return err
	}
	fmt.Println("Number nodes ", len(nodes.Items))

	// TODO for follow up / next PR:
	// Metrics / summary should be an attribute of the JGF outer flux graph
	// Resources should come in from entire group (and not repres. pod)
	var totalAllocCpu int64 = 0

	// Keep a lookup of subnet nodes in case we see one twice
	// We don't want to create a new entity for it in the graph
	subnetLookup := map[string]jgf.Node{}
	var subnetCounter int64 = 0

	for nodeCount, node := range nodes.Items {

		// We should not be scheduling to the control plane
		_, ok := node.Labels[controlPlaneLabel]
		if ok {
			fmt.Println("Skipping control plane node ", node.GetName())
			continue
		}

		// Anything labeled with "skipLabel" meaning it is present,
		// should be skipped
		if skipLabel != "" {
			_, ok := node.Labels[skipLabel]
			if ok {
				fmt.Printf("Skipping node %s\n", node.GetName())
				continue
			}
		}

		if node.Spec.Unschedulable {
			fmt.Printf("Skipping node %s, unschedulable\n", node.GetName())
			continue
		}
		selection := fmt.Sprintf(
			"spec.nodeName=%s,status.phase!=%s,status.phase!=%s",
			node.GetName(), string(corev1.PodSucceeded), string(corev1.PodFailed),
		)
		fieldselector, err := fields.ParseSelector(selection)
		if err != nil {
			return err
		}
		pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
			FieldSelector: fieldselector.String(),
		})
		if err != nil {
			return err
		}

		// Have we seen this subnet node before?
		subnetName := node.Labels["topology.kubernetes.io/zone"]
		subnetNode, exists := subnetLookup[subnetName]
		if !exists {
			// Build the subnet according to topology.kubernetes.io/zone label
			subnetNode = fluxgraph.MakeSubnet(subnetName, subnetCounter)
			subnetCounter += 1

			// This is one example of bidirectional, I won't document in
			// all following occurrences but this is what the function does
			// [cluster] -> contains -> [subnet]
			// [subnet]  ->       in -> [cluster]
			fluxgraph.MakeBidirectionalEdge(clusterNode.Id, subnetNode.Id)
		}

		// These are requests for existing pods, for cpu and memory
		reqs := computeTotalRequests(pods)
		cpuReqs := reqs[corev1.ResourceCPU]
		memReqs := reqs[corev1.ResourceMemory]

		// Actual values that we have available (minus requests)
		totalCpu := node.Status.Allocatable.Cpu().MilliValue()
		totalMem := node.Status.Allocatable.Memory().Value()

		// Values accounting for requests
		availCpu := int64((totalCpu - cpuReqs.MilliValue()) / 1000)
		availMem := totalMem - memReqs.Value()

		// Show existing to compare to
		fmt.Printf("\n📦️ %s\n", node.GetName())
		fmt.Printf("      allocated cpu: %d\n", cpuReqs.Value())
		fmt.Printf("      available cpu: %d\n", availCpu)
		fmt.Printf("      allocated mem: %d\n", memReqs.Value())
		fmt.Printf("      available mem: %d\n", availMem)
		fmt.Printf("       running pods: %d\n\n", len(pods.Items))

		// keep track of overall total
		totalAllocCpu += availCpu
		gpuAllocatable, hasGpuAllocatable := node.Status.Allocatable["nvidia.com/gpu"]

		// TODO possibly look at pod resources vs. node.Status.Allocatable
		// Make the compute node, which is a child of the subnet
		// The parameters here are the node name, and the parent path
		computeNode := fluxgraph.MakeNode(node.Name, subnetNode.Metadata.Name, int64(nodeCount))

		// [subnet] -> contains -> [compute node]
		fluxgraph.MakeBidirectionalEdge(subnetNode.Id, computeNode.Id)

		// Here we are adding GPU resources under nodes
		if hasGpuAllocatable {
			fmt.Println("GPU Resource quantity ", gpuAllocatable.Value())
			for index := 0; index < int(gpuAllocatable.Value()); index++ {

				// The subpath (from and not including root) is the subnet -> node
				subpath := fmt.Sprintf("%s/%s", subnetNode.Metadata.Name, computeNode.Metadata.Name)

				// TODO: can this size be greater than 1?
				gpuNode := fluxgraph.MakeGPU(jgf.NvidiaGPU, subpath, 1, int64(index))

				// [compute] -> contains -> [gpu]
				fluxgraph.MakeBidirectionalEdge(computeNode.Id, gpuNode.Id)
			}

		}

		// Here is where we are adding cores
		for index := 0; index < int(availCpu); index++ {
			subpath := fmt.Sprintf("%s/%s", subnetNode.Metadata.Name, computeNode.Metadata.Name)
			coreNode := fluxgraph.MakeCore(jgf.CoreType, subpath, int64(index))
			fluxgraph.MakeBidirectionalEdge(computeNode.Id, coreNode.Id)
		}

		// Here is where we are adding memory
		fractionMem := availMem >> 30
		for i := 0; i < int(fractionMem); i++ {
			subpath := fmt.Sprintf("%s/%s", subnetNode.Metadata.Name, computeNode.Metadata.Name)
			memoryNode := fluxgraph.MakeMemory(jgf.MemoryType, subpath, 1<<10, int64(i))
			fluxgraph.MakeBidirectionalEdge(computeNode.Id, memoryNode.Id)
		}
	}

	// Get the jgf back as bytes, and we will return string
	err = fluxgraph.WriteJGF(filename)
	if err != nil {
		return err
	}
	return nil
}

// computeTotalRequests sums up the pod requests for the list. We do not consider limits.
func computeTotalRequests(podList *corev1.PodList) map[corev1.ResourceName]resource.Quantity {
	total := map[corev1.ResourceName]resource.Quantity{}
	for _, pod := range podList.Items {
		podReqs, _ := PodRequestsAndLimits(&pod)
		for podReqName, podReqValue := range podReqs {
			if v, ok := total[podReqName]; !ok {
				total[podReqName] = podReqValue
			} else {
				v.Add(podReqValue)
				total[podReqName] = v
			}
		}
	}
	return total
}
