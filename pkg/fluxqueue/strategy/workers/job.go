package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/converged-computing/fluxion/pkg/client"
	pb "github.com/converged-computing/fluxion/pkg/fluxion-grpc"
	api "github.com/converged-computing/fluxqueue/api/v1alpha1"
	"github.com/converged-computing/fluxqueue/pkg/defaults"
	"github.com/converged-computing/fluxqueue/pkg/fluxqueue/queries"
	"github.com/converged-computing/fluxqueue/pkg/types"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	patchTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	klog "k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	// ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
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
	Object      []byte `json:"object"`
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

// Work performs the AskFlux action. Cases include:
// Allocated: the job was successful and does not need to be re-queued. We return nil (completed)
// NotAllocated: the job cannot be allocated and needs to be requeued
// Not possible for some reason, likely needs a cancel
// Are there cases of scheduling out into the future further?
// See https://riverqueue.com/docs/snoozing-jobs
func (w JobWorker) Work(ctx context.Context, job *river.Job[JobArgs]) error {
	wlog.Info("[WORK] Asking Fluxion running for job", "Namespace", job.Args.Namespace, "Name", job.Args.Name, "Args", job.Args)

	//	Let's ask Flux if we can allocate the job!
	fluxionCtx, cancel := context.WithTimeout(context.Background(), 200*time.Second)
	defer cancel()

	// Prepare the request to allocate - convert string to bytes
	fmt.Println(job.Args.Jobspec)
	request := &pb.MatchRequest{Jobspec: job.Args.Jobspec, Reservation: job.Args.Reservation == 1}

	// This is the host where fluxion is running, will be localhost 4242 for sidecar
	// Note that this design isn't ideal - we should be calling this just once
	fluxion, err := client.NewClient("127.0.0.1:4242")
	if err != nil {
		klog.Error(err, "[WORK] Fluxion error connecting to server")
		return err
	}

	// An error here is an error with making the request, nothing about
	// the match/allocation itself.
	response, err := fluxion.Match(fluxionCtx, request)
	fmt.Println(response)
	if err != nil {
		wlog.Info("[WORK] Fluxion did not receive any match response", "Error", err)
		return err
	}

	// Convert the response into an error code that indicates if we should run again.
	// If we have an allocation, the job/etc must be un-suspended or released
	// The database also needs to be updated (not sure how to do that yet)
	pool, err := pgxpool.New(fluxionCtx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}

	// Not reserved AND not allocated indicates not possible
	if !response.Reserved && response.GetAllocation() == "" {
		return river.JobCancel(fmt.Errorf("fluxion could not allocate nodes for %s/%s, likely Unsatisfiable", job.Args.Namespace, job.Args.Name))
	}

	// Flux job identifier (known to fluxion)
	fluxID := response.GetJobid()

	// If it's reserved, we need to add the id to our reservation table
	// TODO need to clean up this table...
	if response.Reserved {
		rRows, err := pool.Query(fluxionCtx, queries.AddReservationQuery, job.Args.Name, fluxID)
		if err != nil {
			return err
		}
		defer rRows.Close()
	}

	// This means we didn't get an allocation - we might have a reservation (to do
	// something with later) but for now we just print it.
	// TODO we need an IsAllocated function
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
	wlog.Info("Allocation response", "Nodes", nodes)

	// Parse binary data back into respective object and update in API
	if job.Args.Type == api.JobWrappedJob.String() {
		err = unsuspendJob(job.Args.Object, nodes)
		if err != nil {
			wlog.Info("Error unsuspending job", "Namespace", job.Args.Namespace, "Name", job.Args.Name, "Error", err)
			return err
		}
	} else if job.Args.Type == api.JobWrappedPod.String() {
		err = w.ungatePod(job.Args.Namespace, job.Args.Name, nodes)
		if err != nil {
			wlog.Info("Error ungating pod", "Namespace", job.Args.Namespace, "Name", job.Args.Name, "Error", err)
			return err
		}
		wlog.Info("Success ungating pod", "Namespace", job.Args.Namespace, "Name", job.Args.Name)
	} else {
		wlog.Info("Error understanding job type", "Type", job.Args.Type, "Name", job.Args.Namespace, "Name", job.Args.Name)
		return fmt.Errorf("unknown job type %s passed to fluxion schedule for job %s/%s", job.Args.Type, job.Args.Namespace, job.Args.Name)
	}

	// TODO Kick off a cleaning job for when everything should be cancelled, but only if
	// there is a deadline set. We can't set a deadline for services, etc.
	// This is here instead of responding to deletion / termination since a job might
	// run longer than the duration it is allowed.
	//if job.Args.Duration > 0 {
	// err = SubmitCleanup(ctx, pool, pod.Spec.ActiveDeadlineSeconds, job.Args.Podspec, int64(fluxID), true, []string{})
	//if err != nil {
	//	return err
	//}
	//}
	wlog.Info("[WORK] nodes allocated for job", "JobId", fluxID, "Nodes", nodes, "Namespace", job.Args.Namespace, "Name", job.Args.Name)
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

	wlog.Info("Parsing fluxion nodes", "Nodes", graph.Graph.Nodes)

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
func unsuspendJob(object []byte, nodes []string) error {
	return nil
}

// Ungate the pod, adding an annotation for nodes along with the fluxion scheduler
func (w JobWorker) ungatePod(namespace, name string, nodes []string) error {
	ctx := context.Background()

	// Get the pod to update
	client, err := kubernetes.NewForConfig(&w.RESTConfig)
	if err != nil {
		return err
	}
	pod, err := client.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
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

	// Patch the pod to add the nodes
	nodesStr := strings.Join(nodes, ",")
	payload := `{"metadata": {"labels": {"` + defaults.NodesLabel + `": "` + nodesStr + `"}}}`
	fmt.Println(payload)
	_, err = client.CoreV1().Pods(namespace).Patch(ctx, name, patchTypes.MergePatchType, []byte(payload), metav1.PatchOptions{})
	if err != nil {
		return err
	}

	// Patch the pod to remove the scheduling gate at the correct index
	patch := []byte(`[{"op": "remove", "path": "/spec/schedulingGates/` + fmt.Sprintf("%d", gateIndex) + `"}]`)
	_, err = client.CoreV1().Pods(namespace).Patch(ctx, name, patchTypes.JSONPatchType, patch, metav1.PatchOptions{})
	return err
}
