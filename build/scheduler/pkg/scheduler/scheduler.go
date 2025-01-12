package scheduler

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	corev1helpers "k8s.io/component-helpers/scheduling/corev1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

var (
	NodesLabel         = "fluxqueue/fluxion-nodes"
	UnschedulableLabel = "fluxqueue/unschedulable"
)

// This is a simple scheduler plugin that receives pods that already have a node assignment.
// The assignments are done by fluxion, which is running as a service.
// To ensure consistency, any pod that does not have the fluxion annotation is denied.
// This ensures that the state of the cluster is owned and controlled by fluxion.

type FluxionScheduler struct{}

var (
	_ framework.PreFilterPlugin = &FluxionScheduler{}
	_ framework.FilterPlugin    = &FluxionScheduler{}
	_ framework.QueueSortPlugin = &FluxionScheduler{}
)

const (
	Name = "FluxionScheduler"
)

func (fs *FluxionScheduler) Name() string {
	return Name
}

// Less is required for QueueSortPlugin.
// We sort based on priority, init timestamps, and then keys
func (fs *FluxionScheduler) Less(podInfo1, podInfo2 *framework.QueuedPodInfo) bool {
	prio1 := corev1helpers.PodPriority(podInfo1.Pod)
	prio2 := corev1helpers.PodPriority(podInfo2.Pod)
	if prio1 != prio2 {
		return prio1 > prio2
	}
	return getNamespacedName(podInfo1.Pod) < getNamespacedName(podInfo2.Pod)
}

func (fs *FluxionScheduler) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

// GetNamespacedName returns the namespaced name.
func getNamespacedName(obj metav1.Object) string {
	return fmt.Sprintf("%v/%v", obj.GetNamespace(), obj.GetName())
}

// Prefilter returns the node from the label, otherwise we don't schedule
func (fs *FluxionScheduler) PreFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod) (*framework.PreFilterResult, *framework.Status) {
	klog.InfoS("PreFilter received contender pod", "pod", klog.KObj(pod))

	// Case 1: unscheduleable and unresolvable
	_, isImpossible := pod.ObjectMeta.Labels[UnschedulableLabel]
	if isImpossible {
		return nil, framework.NewStatus(framework.UnschedulableAndUnresolvable, "Fluxion cannot find resources")
	}

	nodesLabel, ok := pod.ObjectMeta.Labels[NodesLabel]
	if !ok {
		klog.InfoS("Pod is missing nodes label", "Label", NodesLabel, "pod", klog.KObj(pod))
		return nil, framework.NewStatus(framework.UnschedulableAndUnresolvable, "Pod missing nodes label")
	}

	// Create a map to store the JSON data
	nodes := strings.Split(nodesLabel, "__")

	// Does this pod have an index? (e.g., stateful set)
	podIndex, ok := pod.ObjectMeta.Labels["apps.kubernetes.io/pod-index"]
	if !ok {

		// Next try for batch completion index (job pod)
		podIndex, ok = pod.ObjectMeta.Labels["batch.kubernetes.io/job-completion-index"]
		if !ok {
			podIndex = "0"
		}
	}
	index, err := strconv.Atoi(podIndex)
	if err != nil {
		return nil, framework.NewStatus(framework.UnschedulableAndUnresolvable, err.Error())
	}

	// Get the correct name based on the podIndex
	// This shouldn't happen, but don't let it index an array too short
	if index >= len(nodes) {
		return nil, framework.NewStatus(framework.Error, "Not enough nodes for pod")
	}
	nodeNames := sets.New(nodes[index])
	result := framework.PreFilterResult{NodeNames: nodeNames}
	klog.InfoS("PreFilter node assignment", "pod", klog.KObj(pod), "node", nodes[index])
	return &result, framework.NewStatus(framework.Success, "Fluxion scheduler assigned node")
}

func (fs *FluxionScheduler) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	klog.InfoS("Filter received pod assignment", "pod", klog.KObj(pod), "node", nodeInfo.Node().Name)
	return framework.NewStatus(framework.Success)
}

// New returns an empty FluxionScheduler plugin, which only provides a queue sort!
func New(_ context.Context, _ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
	return &FluxionScheduler{}, nil
}
