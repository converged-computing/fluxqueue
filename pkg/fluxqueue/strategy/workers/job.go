package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/converged-computing/fluxion/pkg/client"
	pb "github.com/converged-computing/fluxion/pkg/fluxion-grpc"
	api "github.com/converged-computing/fluxqueue/api/v1alpha1"
	"github.com/converged-computing/fluxqueue/pkg/defaults"
	"github.com/converged-computing/fluxqueue/pkg/fluxqueue/queries"
	"github.com/converged-computing/fluxqueue/pkg/types"
	"github.com/riverqueue/river"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	patchTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	wlog = ctrl.Log.WithName("worker")
)

// The job worker submits jobs to fluxion with match allocate
// or match allocate else reserve, depending on the reservation depth
func (args JobArgs) Kind() string { return "job" }

type JobWorker struct {
	river.WorkerDefaults[JobArgs]
	RESTConfig rest.Config
}

// NewJobWorker returns a new job worker with a Fluxion client
func NewJobWorker(cfg rest.Config) (*JobWorker, error) {
	worker := JobWorker{RESTConfig: cfg}
	//	defer worker.fluxion.Close()
	return &worker, nil
}

// JobArgs serializes a postgres row back into fields for the FluxJob
// We add extra fields to anticipate getting node assignments
type JobArgs struct {
	Jobspec     string `json:"jobspec"`
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	FluxJobName string `json:"flux_job_name"`
	Type        string `json:"type"`

	// If true, we are allowed to ask Fluxion for a reservation
	Reservation int32 `json:"reservation"`
	Duration    int32 `json:"duration"`
	Size        int32 `json:"size"`

	// Nodes to return to Kubernetes custom scheduler plugin to bind
	Nodes string `json:"nodes"`
}

// Work performs the AskFlux action. Any error returned that is due to not having resources means
// the job will remain in the worker queue to AskFluxion again.
func (w JobWorker) Work(ctx context.Context, job *river.Job[JobArgs]) error {
	wlog.Info("Asking Fluxion to schedule job", "Namespace", job.Args.Namespace, "Name", job.Args.Name, "Nodes", job.Args.Size)

	// Let's ask Flux if we can allocate nodes for the job!
	fluxionCtx, cancel := context.WithTimeout(context.Background(), 200*time.Second)
	defer cancel()

	// Prepare the request to allocate - convert string to bytes
	request := &pb.MatchRequest{Jobspec: job.Args.Jobspec, Reservation: job.Args.Reservation == 1}

	// This is the host where fluxion is running, will be localhost 4242 for sidecar
	fluxion, err := client.NewClient("127.0.0.1:4242")
	if err != nil {
		wlog.Error(err, "Fluxion error connecting to server")
		return err
	}
	defer fluxion.Close()

	// An error here is an error with making the request, nothing about
	// the match/allocation itself.
	response, err := fluxion.Match(fluxionCtx, request)
	if err != nil {
		wlog.Info("[WORK] Fluxion did not receive any match response", "Error", err)
		return err
	}

	// If we asked for a reservation, and it wasn't reserved AND not allocated, this means it's not possible
	// We currently don't have grow/shrink added so this means it will never be possible.
	// We will unsuspend the job but add a label that indicates it is not schedulable.
	// The cancel here will complete the task (and we won't ask again)
	if job.Args.Reservation == 1 && !response.Reserved && response.GetAllocation() == "" {
		w.markUnschedulable(job.Args)
		return river.JobCancel(fmt.Errorf("fluxion could not allocate nodes for %s/%s, likely Unsatisfiable", job.Args.Namespace, job.Args.Name))
	}

	// Flux job identifier (known to fluxion)
	fluxID := response.GetJobid()

	// If it's reserved, add the id to our reservation table
	// TODO need to clean up this table... but these tasks run async...
	if response.Reserved {
		w.reserveJob(fluxionCtx, job.Args, fluxID)
	}

	// This means we didn't get an allocation - we might have a reservation
	if response.GetAllocation() == "" {
		// This will have the job be retried in the queue, still based on sorted schedule time and priority
		return fmt.Errorf("fluxion could not allocate nodes for job %s/%s", job.Args.Namespace, job.Args.Name)
	}

	// Now get the nodes. These are actually cores assigned to nodes, so we need to keep count
	nodes, err := parseNodes(response.Allocation)
	if err != nil {
		wlog.Info("Error parsing nodes from fluxion response", "Namespace", job.Args.Namespace, "Name", job.Args.Name, "Error", err)
		return err
	}
	wlog.Info("Fluxion allocation response", "Nodes", nodes)

	// Unsuspend the job or ungate the pods, adding the node assignments as labels for the scheduler
	err = w.releaseJob(ctx, job.Args, fluxID, nodes)
	if err != nil {
		return err
	}
	wlog.Info("Fluxion finished allocating nodes for job", "JobId", fluxID, "Nodes", nodes, "Namespace", job.Args.Namespace, "Name", job.Args.Name)
	return nil
}

// Release job will unsuspend a job or ungate pods to allow for scheduling
func (w JobWorker) releaseJob(ctx context.Context, args JobArgs, fluxID int64, nodes []string) error {
	var err error

	if args.Type == api.JobWrappedJob.String() {

		// Kubernetes Job Type
		err = w.unsuspendJob(args.Namespace, args.Name, nodes, fluxID)
		if err != nil {
			wlog.Info("Error unsuspending job", "Namespace", args.Namespace, "Name", args.Name, "Error", err)
			return err
		}
		wlog.Info("Success unsuspending job", "Namespace", args.Namespace, "Name", args.Name)

	} else if args.Type == api.JobWrappedDeployment.String() ||
		args.Type == api.JobWrappedPod.String() ||
		args.Type == api.JobWrappedReplicaSet.String() ||
		args.Type == api.JobWrappedStatefulSet.String() {
		w.ungatePod(ctx, args.Namespace, args.Name, args.Type, nodes, fluxID)

	} else {

		// Unknown type (gets marked as unschedulable)
		wlog.Info("Error understanding job type", "Type", args.Type, "Name", args.Namespace, "Name", args.Name)
		return fmt.Errorf("unknown job type %s passed to fluxion schedule for job %s/%s", args.Type, args.Namespace, args.Name)
	}
	return err
}

// Reject job adds labels to the pods to indicate not schedulable
func (w JobWorker) markUnschedulable(args JobArgs) error {
	if args.Type == api.JobWrappedJob.String() {
		err := w.rejectJob(args.Namespace, args.Name)
		if err != nil {
			wlog.Info("Error marking job unschedulable", "Namespace", args.Namespace, "Name", args.Name, "Error", err)
			return err
		}
	} else if args.Type == api.JobWrappedPod.String() {
		err := w.rejectPod(args.Namespace, args.Name)
		if err != nil {
			wlog.Info("Error marking pod unschedulable", "Namespace", args.Namespace, "Name", args.Name, "Error", err)
			return err
		}
	}
	return nil
}

// reserveJob adds the flux job id to the reservation table to cleanup later
func (w JobWorker) reserveJob(ctx context.Context, args JobArgs, fluxID int64) error {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("issue creating new pool: %s", err.Error())
	}
	defer pool.Close()

	rRows, err := pool.Query(ctx, queries.AddReservationQuery, args.Name, fluxID)
	if err != nil {
		return err
	}
	defer rRows.Close()
	return nil
}

// parseNodes parses the allocation nodes into a lookup with core counts
// We will add these as labels onto each pod for the scheduler, or as one
func parseNodes(allocation string) ([]string, error) {

	// We can eventually send over more metadata, for now just a list of nodes
	nodesWithCores := map[string]int{}
	nodes := []string{}

	// The response is the graph with assignments. Here we parse the graph into a struct to get nodes.
	var graph types.AllocationResponse
	err := json.Unmarshal([]byte(allocation), &graph)
	if err != nil {
		return nodes, err
	}

	// To start, just parse nodes and not cores (since we can't bind on level of core)
	for _, node := range graph.Graph.Nodes {
		if node.Metadata.Type == "node" {
			nodeId := node.Metadata.Basename
			_, ok := nodesWithCores[nodeId]
			if !ok {
				nodesWithCores[nodeId] = 0
				nodes = append(nodes, nodeId)
			}
			// Keep a record of cores assigned per node
			nodesWithCores[nodeId] += 1
		}
	}
	return nodes, nil
}

// Unsuspend the job, adding an annotation for nodes along with the fluxion scheduler
func (w JobWorker) unsuspendJob(namespace, name string, nodes []string, fluxId int64) error {
	ctx := context.Background()

	// Get the pod to update
	client, err := kubernetes.NewForConfig(&w.RESTConfig)
	if err != nil {
		return err
	}

	// Convert jobid to string
	jobid := fmt.Sprintf("%d", fluxId)

	// Add the nodes and flux id as an annotation to the pods that will be generated
	nodesStr := strings.Join(nodes, "__")
	payload := `{"spec": {"suspend": false, "template": {"metadata": {"labels": {"` + defaults.NodesLabel + `": "` + nodesStr + `", "` + defaults.FluxJobIdLabel + `": "` + jobid + `"}}}}}`
	_, err = client.BatchV1().Jobs(namespace).Patch(ctx, name, patchTypes.StrategicMergePatchType, []byte(payload), metav1.PatchOptions{})
	return err
}

// A stateful set has an ordinal index and can be given to scheduler to assign
// we only need to tweak this if for some reason the index is no longer valid
// (e.g., scaling up and down)
func (w JobWorker) ungateSet(namespace, name string, nodes []string, fluxId int64) error {
	ctx := context.Background()

	client, err := kubernetes.NewForConfig(&w.RESTConfig)
	if err != nil {
		return err
	}
	jobid := fmt.Sprintf("%d", fluxId)
	nodesStr := strings.Join(nodes, "__")
	payload := `{"spec": {"template": {"metadata": {"labels": {"` + defaults.NodesLabel + `": "` + nodesStr + `", "` + defaults.FluxJobIdLabel + `": "` + jobid + `"}}}}}`
	_, err = client.AppsV1().StatefulSets(namespace).Patch(ctx, name, patchTypes.StrategicMergePatchType, []byte(payload), metav1.PatchOptions{})
	return err
}

// patchUnsuspend patches a pod to unsuspend it.
func patchUnsuspend(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
	patch := []byte(`[{"op": "replace", "path": "/spec/suspend", "value": null}]`)
	_, err := client.BatchV1().Jobs(namespace).Patch(ctx, name, patchTypes.JSONPatchType, patch, metav1.PatchOptions{})
	return err
}

// rejectJob adds a label to indicate unschedulable and unresolvable
func (w JobWorker) rejectJob(namespace, name string) error {
	ctx := context.Background()

	// Get the pod to update
	client, err := kubernetes.NewForConfig(&w.RESTConfig)
	if err != nil {
		return err
	}
	payload := `{"spec": {"suspend": false, "template": {"metadata": {"labels": {"` + defaults.UnschedulableLabel + `": "yes"}}}}}`
	_, err = client.BatchV1().Jobs(namespace).Patch(ctx, name, patchTypes.StrategicMergePatchType, []byte(payload), metav1.PatchOptions{})
	if err != nil {
		return err
	}
	// And unsuspend the job so it is sent to the scheduler
	return patchUnsuspend(ctx, client, name, namespace)
}

// ungatePod submits jobs to ungate. We do this because Kubernetes isn't always reliable
// to get pods that we need via the API, or operations to patch, etc.
func (w JobWorker) ungatePod(
	ctx context.Context,
	namespace, name, jobType string,
	nodes []string,
	fluxId int64,
) error {

	// Create a job to ungate the deployment pods
	riverClient := river.ClientFromContext[pgx.Tx](ctx)
	insertOpts := river.InsertOpts{
		Tags:  []string{"ungate"},
		Queue: "task_queue",
	}
	ungateArgs := UngateArgs{
		Name:      name,
		Namespace: namespace,
		Nodes:     nodes,
		JobID:     fluxId,
		Type:      jobType,
	}
	_, err := riverClient.Insert(ctx, ungateArgs, &insertOpts)
	if err != nil {
		wlog.Info("Error inserting ungate job", "Namespace", namespace, "Name", name, "Error", err)
	}
	return err
}

// removeGate removes the scheduling gate from the pod
func removeGate(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
	pod, err := client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// Create a JSON patch to remove the scheduling gates
	gateIndex := 0
	for i, gate := range pod.Spec.SchedulingGates {
		if gate.Name == defaults.SchedulingGateName {
			gateIndex = i
			break
		}
	}

	// Patch the pod to remove the scheduling gate at the correct index
	patch := []byte(`[{"op": "remove", "path": "/spec/schedulingGates/` + fmt.Sprintf("%d", gateIndex) + `"}]`)
	_, err = client.CoreV1().Pods(namespace).Patch(ctx, name, patchTypes.JSONPatchType, patch, metav1.PatchOptions{})
	return err
}

// Reject a pod, mark as unschedulable and unresolvable
func (w JobWorker) rejectPod(namespace, name string) error {
	ctx := context.Background()

	// Get the pod to update
	client, err := kubernetes.NewForConfig(&w.RESTConfig)
	if err != nil {
		return err
	}
	payload := `{"metadata": {"labels": {"` + defaults.UnschedulableLabel + `": "yes"}}}`
	_, err = client.CoreV1().Pods(namespace).Patch(ctx, name, patchTypes.MergePatchType, []byte(payload), metav1.PatchOptions{})
	if err != nil {
		return err
	}
	return removeGate(ctx, client, namespace, name)
}
