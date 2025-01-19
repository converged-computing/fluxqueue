package workers

import (
	"context"
	"fmt"
	"strings"

	api "github.com/converged-computing/fluxqueue/api/v1alpha1"
	"github.com/converged-computing/fluxqueue/pkg/defaults"
	"github.com/riverqueue/river"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	patchTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Ungate workers explicitly ungate pods, and add node labels
func (args UngateArgs) Kind() string { return "ungate" }

type UngateWorker struct {
	river.WorkerDefaults[UngateArgs]
	RESTConfig rest.Config
}

// NewJobWorker returns a new job worker with a Fluxion client
func NewUngateWorker(cfg rest.Config) (*UngateWorker, error) {
	worker := UngateWorker{RESTConfig: cfg}
	return &worker, nil
}

// JobArgs serializes a postgres row back into fields for the FluxJob
// We add extra fields to anticipate getting node assignments
type UngateArgs struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	Namespace string   `json:"namespace"`
	Nodes     []string `json:"nodes"`
	JobID     int64    `json:"jobid"`
}

// Ungate a specific pod for a group (e.g., deployment)
// Right now we aren't using this for single pods (but can/will)
func (w UngateWorker) Work(ctx context.Context, job *river.Job[UngateArgs]) error {

	var err error
	wlog.Info("Running ungate worker", "Namespace", job.Args.Namespace, "Name", job.Args.Name)
	jobid := fmt.Sprintf("%d", job.Args.JobID)

	client, err := kubernetes.NewForConfig(&w.RESTConfig)
	if err != nil {
		wlog.Info("Error getting Kubernetes client", "Namespace", job.Args.Namespace, "Name", job.Args.Name, "Error", err)
		return err
	}

	// Ungate single pod (should only be one)
	if job.Args.Type == api.JobWrappedPod.String() {
		nodesStr := strings.Join(job.Args.Nodes, "__")
		payload := `{"metadata": {"labels": {"` + defaults.NodesLabel + `": "` + nodesStr + `", "` + defaults.FluxJobIdLabel + `": "` + jobid + `"}}}`
		_, err = client.CoreV1().Pods(job.Args.Namespace).Patch(ctx, job.Args.Name, patchTypes.MergePatchType, []byte(payload), metav1.PatchOptions{})
		if err != nil {
			return err
		}
		err = removeGate(ctx, client, job.Args.Namespace, job.Args.Name)
		if err != nil {
			wlog.Info("Error in removing single pod", "Error", err)
			return err
		}
		return err
	}

	// If we get here, we have deployment, statefulset, replicaset
	var selector string
	if job.Args.Type == api.JobWrappedDeployment.String() {
		selector = fmt.Sprintf("%s=deployment-%s-%s", defaults.SelectorLabel, job.Args.Name, job.Args.Namespace)
	} else if job.Args.Type == api.JobWrappedReplicaSet.String() {
		selector = fmt.Sprintf("%s=replicaset-%s-%s", defaults.SelectorLabel, job.Args.Name, job.Args.Namespace)
	} else if job.Args.Type == api.JobWrappedStatefulSet.String() {
		selector = fmt.Sprintf("%s=statefulset-%s-%s", defaults.SelectorLabel, job.Args.Name, job.Args.Namespace)
	}

	// 4. Get pods in the default namespace
	pods, err := client.CoreV1().Pods(job.Args.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	wlog.Info("Selector returned pods for nodes", "Pods", len(pods.Items), "Nodes", len(job.Args.Nodes))
	if err != nil {
		wlog.Info("Error listing pods in ungate worker", "Namespace", job.Args.Namespace, "Name", job.Args.Name, "Error", err)
		return err
	}
	// Ungate as many as we are able
	for i, pod := range pods.Items {

		// This shouldn't happen
		if i >= len(pods.Items) {
			wlog.Info("Warning - we have more pods than nodes")
			break
		}

		// We should not try to ungate (and assign a node) to a pod that
		// already has been ungated
		ungated := true
		if pod.Spec.SchedulingGates != nil {
			for _, gate := range pod.Spec.SchedulingGates {
				if gate.Name == defaults.SchedulingGateName {
					ungated = false
					break
				}
			}
		}
		if ungated {
			continue
		}
		payload := `{"metadata": {"labels": {"` + defaults.NodesLabel + `": "` + job.Args.Nodes[i] + `", "` + defaults.FluxJobIdLabel + `": "` + jobid + `"}}}`
		_, err = client.CoreV1().Pods(job.Args.Namespace).Patch(ctx, pod.ObjectMeta.Name, patchTypes.MergePatchType, []byte(payload), metav1.PatchOptions{})
		if err != nil {
			wlog.Info("Error in patching deployment pod", "Error", err)
			return err
		}
		err = removeGate(ctx, client, pod.ObjectMeta.Namespace, pod.ObjectMeta.Name)
		if err != nil {
			wlog.Info("Error in removing deployment pod gate", "Error", err)
			return err
		}
	}

	// Kubernetes has not created the pod objects yet
	// Returning an error will have it run again, with a delay
	// https://riverqueue.com/docs/job-retries
	if len(pods.Items) < len(job.Args.Nodes) {
		return fmt.Errorf("ungate pods job did not have all pods")
	}
	return err
}
