package workers

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/converged-computing/fluxion/pkg/client"
	pb "github.com/converged-computing/fluxion/pkg/fluxion-grpc"
	"github.com/converged-computing/fluxqueue/pkg/fluxqueue/queries"
	"github.com/jackc/pgx/v5/pgxpool"
	klog "k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/riverqueue/river"
)

var (
	wlog = ctrl.Log.WithName("worker")
)

// The job worker submits jobs to fluxion with match allocate
// or match allocate else reserve, depending on the reservation depth
func (args JobArgs) Kind() string { return "job" }

type JobWorker struct {
	river.WorkerDefaults[JobArgs]
	fluxion client.Client
}

// NewJobWorker returns a new job worker with a Fluxion client
func NewJobWorker() (*JobWorker, error) {
	worker := JobWorker{}

	// This is the host where fluxion is running, will be localhost 4242 for sidecar
	// Note that this design isn't ideal - we should be calling this just once
	c, err := client.NewClient("127.0.0.1:4242")
	if err != nil {
		klog.Error(err, "[WORK] Fluxion error connecting to server")
		return nil, err
	}
	worker.fluxion = c
	//	defer worker.fluxion.Close()
	return &worker, err
}

// JobArgs serializes a postgres row back into fields for the FluxJob
// We add extra fields to anticipate getting node assignments
type JobArgs struct {
	Jobspec   string `json:"jobspec"`
	Object    []byte `json:"object"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"`

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
	klog.Infof("[WORK] Asking Fluxion running for job %s/%s", job.Args.Namespace, job.Args.Name)

	//	Let's ask Flux if we can allocate the job!
	fluxionCtx, cancel := context.WithTimeout(context.Background(), 200*time.Second)
	defer cancel()

	// Prepare the request to allocate - convert string to bytes
	fmt.Println(job.Args.Jobspec)
	request := &pb.MatchRequest{Jobspec: job.Args.Jobspec, Reservation: job.Args.Reservation == 1}

	// An error here is an error with making the request, nothing about
	// the match/allocation itself.
	response, err := w.fluxion.Match(fluxionCtx, request)
	fmt.Println(response)
	if err != nil {
		wlog.Info("[WORK] Fluxion did not receive any match response", err)
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
		errorMessage := fmt.Errorf("fluxion could not allocate nodes for %s", job.Args.Name)
		klog.Info(errorMessage)

		// This will have the job be retried in the queue, still based on sorted schedule time and priority
		return errorMessage
	}
	klog.Infof("Fluxion response with allocation is %s", response)

	// Get the nodelist and serialize into list of strings for job args
	// TODO need function to get nodelist
	nodes := strings.Split(response.Allocation, ",")
	nodeStr := strings.Join(nodes, ",")
	fmt.Println(nodeStr)
	// Update nodes for the job
	//	rows, err := pool.Query(fluxionCtx, queries.UpdateNodesQuery, nodeStr, job.ID)
	//	if err != nil {
	//		return err
	//	}
	//	defer rows.Close()

	// Add the job id to pending (for later cleanup)
	//	_, err = pool.Exec(fluxionCtx, queries.UpdatingPendingWithFluxID, fluxID, job.Args.GroupName, job.Args.Namespace)
	//	if err != nil {
	//		return err
	//	}

	// Kick off a cleaning job for when everything should be cancelled, but only if
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
