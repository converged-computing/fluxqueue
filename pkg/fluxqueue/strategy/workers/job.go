package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	corev1 "k8s.io/api/core/v1"
	klog "k8s.io/klog/v2"

	pb "github.com/converged-computing/fluxion/pkg/fluxion-grpc"

	"github.com/converged-computing/fluxqueue/pkg/fluxqueue/queries"
	"github.com/riverqueue/river"
)

// The job worker submits jobs to fluxion with match allocate
// or match allocate else reserve, depending on the reservation depth
func (args JobArgs) Kind() string { return "job" }

type JobWorker struct {
	river.WorkerDefaults[JobArgs]
}

type JobArgs struct {

	// Submit Args
	Jobspec   string `json:"jobspec"`
	Podspec   string `json:"podspec"`
	GroupName string `json:"groupName"`
	GroupSize int32  `json:"groupSize"`
	Duration  int32  `json:"duration"`
	Namespace string `json:"namespace"`

	// If true, we are allowed to ask Fluxion for a reservation
	Reservation bool `json:"reservation"`

	// Nodes return to Kubernetes to bind
	Nodes string `json:"nodes"`

	// Comma separated list of names
	Names string `json:"names"`
}

// Work performs the AskFlux action. Cases include:
// Allocated: the job was successful and does not need to be re-queued. We return nil (completed)
// NotAllocated: the job cannot be allocated and needs to be requeued
// Not possible for some reason, likely needs a cancel
// Are there cases of scheduling out into the future further?
// See https://riverqueue.com/docs/snoozing-jobs
func (w JobWorker) Work(ctx context.Context, job *river.Job[JobArgs]) error {
	klog.Infof("[JOB-WORKER-START] JobStatus Running for group %s", job.Args.GroupName)

	// Convert jobspec back to json, and then pod
	var pod corev1.Pod
	err := json.Unmarshal([]byte(job.Args.Podspec), &pod)
	if err != nil {
		return err
	}

	// IMPORTANT: this is a JobSpec for *one* pod, assuming they are all the same.
	// This obviously may not be true if we have a hetereogenous PodGroup.
	// We name it based on the group, since it will represent the group
	// TODO(vsoch): generate this from a group of podspecs instead
	jobspec := "" //resources.PreparePodJobSpec(&pod, job.Args.GroupName)
	klog.Infof("Prepared pod jobspec %s", jobspec)

	// Connect to the Fluxion service. Returning an error means we retry
	// see: https://riverqueue.com/docs/job-retries
	conn, err := grpc.Dial("127.0.0.1:4242", grpc.WithInsecure())
	if err != nil {
		klog.Error(err, "[Fluxnetes] AskFlux error connecting to server")
		return err
	}
	defer conn.Close()

	//	Let's ask Flux if we can allocate the job!
	fluxion := pb.NewFluxionServiceClient(conn)
	fluxionCtx, cancel := context.WithTimeout(context.Background(), 200*time.Second)
	defer cancel()

	// Prepare the request to allocate.
	// Note that reserve will just give an ETA for the future.
	// We don't want to actually run this job again then, because newer
	// jobs could come in and take precedence. It's more an FYI for the
	// user when we expose some kubectl tool.
	request := &pb.MatchRequest{}
	//		Podspec: jobspec,
	//		Reserve: job.Args.Reservation,
	//		Count:   job.Args.GroupSize,
	//		JobName: job.Args.GroupName,
	//	}

	// An error here is an error with making the request, nothing about
	// the match/allocation itself.
	response, err := fluxion.Match(fluxionCtx, request)
	if err != nil {
		klog.Error("[Fluxnetes] AskFlux did not receive any match response", err)
		return err
	}

	// Convert the response into an error code that indicates if we should run again.
	// We must update the database with nodes from here with a query
	// This will be sent back to the Kubernetes scheduler
	// TODO(vsoch): should this be error (which will retry) or cancel (not)?
	pool, err := pgxpool.New(fluxionCtx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}

	// Not reserved AND not allocated indicates not possible
	if !response.Reserved && response.GetAllocation() == "" {
		errorMessage := fmt.Sprintf("Fluxion could not allocate nodes for %s, likely Unsatisfiable", job.Args.GroupName)
		klog.Info(errorMessage)
		return river.JobCancel(fmt.Errorf(errorMessage))
	}

	// Flux job identifier (known to fluxion)
	fluxID := response.GetJobid()

	// If it's reserved, we need to add the id to our reservation table
	// TODO need to clean up this table...
	if response.Reserved {
		rRows, err := pool.Query(fluxionCtx, queries.AddReservationQuery, job.Args.GroupName, fluxID)
		if err != nil {
			return err
		}
		defer rRows.Close()
	}

	// This means we didn't get an allocation - we might have a reservation (to do
	// something with later) but for now we just print it.
	// TODO we need an IsAllocated function
	if response.GetAllocation() == "" {
		errorMessage := fmt.Errorf("fluxion could not allocate nodes for %s", job.Args.GroupName)
		klog.Info(errorMessage)

		// This will have the job be retried in the queue, still based on sorted schedule time and priority
		return errorMessage
	}
	klog.Infof("Fluxion response with allocation is %s", response)

	// Get the nodelist and serialize into list of strings for job args
	// TODO need function to get nodelist
	//nodelist := response.Allocation

	// We assume that each node gets N tasks
	nodes := []string{}
	//for _, node := range nodelist {
	//	for i := 0; i < int(node.Tasks); i++ {
	//		nodes = append(nodes, node.NodeID)
	//	}
	//}
	nodeStr := strings.Join(nodes, ",")

	// Update nodes for the job
	rows, err := pool.Query(fluxionCtx, queries.UpdateNodesQuery, nodeStr, job.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Add the job id to pending (for later cleanup)
	_, err = pool.Exec(fluxionCtx, queries.UpdatingPendingWithFluxID, fluxID, job.Args.GroupName, job.Args.Namespace)
	if err != nil {
		return err
	}

	// Kick off a cleaning job for when everything should be cancelled, but only if
	// there is a deadline set. We can't set a deadline for services, etc.
	// This is here instead of responding to deletion / termination since a job might
	// run longer than the duration it is allowed.
	if job.Args.Duration > 0 {
		err = SubmitCleanup(ctx, pool, pod.Spec.ActiveDeadlineSeconds, job.Args.Podspec, int64(fluxID), true, []string{})
		if err != nil {
			return err
		}
	}
	klog.Infof("[JOB-WORKER-COMPLETE] nodes allocated %s for group %s (flux job id %d)\n",
		nodeStr, job.Args.GroupName, fluxID)
	return nil
}
