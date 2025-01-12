package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/converged-computing/fluxion/pkg/client"
	"k8s.io/client-go/rest"

	pb "github.com/converged-computing/fluxion/pkg/fluxion-grpc"

	"github.com/riverqueue/river"
)

type CleanupArgs struct {
	// Cleanup issues a cancel request to fluxion
	// This should be triggered by Kubernetes deletion events
	FluxID  int64  `json:"fluxid"`
	Podspec string `json:"podspec"`
}

// The cleanup workers cleans up a reservation (issuing cancel)
func (args CleanupArgs) Kind() string { return "cleanup" }

type CleanupWorker struct {
	river.WorkerDefaults[CleanupArgs]
}

// NewJobWorker returns a new job worker with a Fluxion client
func NewCleanupWorker(cfg rest.Config) (*CleanupWorker, error) {
	worker := CleanupWorker{}
	return &worker, nil
}

// Work performs the Cancel action, first cancelling in Kubernetes (if needed)
// and then cancelling in fluxion.
func (w CleanupWorker) Work(ctx context.Context, job *river.Job[CleanupArgs]) error {

	// Wrapper to actual cleanup function that can be called from elsewhere
	return Cleanup(ctx, job.Args.FluxID)
}

// Cleanup handles a call to fluxion to cancel (if appropriate) along with Kubernetes object deletion,
// and finally, deletion from Pending queue (table) to allow new jobs in
func Cleanup(ctx context.Context, fluxid int64) error {
	wlog.Info("Cleanup (cancel) running", "JobID", fluxid)

	// We only delete from fluxion if there is a flux id
	// A valid fluxID is 0 or greater
	var err error
	if fluxid > -1 {
		err = deleteFluxion(fluxid)
		if err != nil {
			wlog.Info("Error issuing cancel to fluxion", "fluxID", fluxid)
		}
		return err
	}
	return nil
}

// deleteFluxion issues a cancel to Fluxion, our scheduler
func deleteFluxion(fluxID int64) error {

	// Connect to the Fluxion service. Returning an error means we retry
	// see: https://riverqueue.com/docs/job-retries
	fluxion, err := client.NewClient("127.0.0.1:4242")
	if err != nil {
		wlog.Error(err, "Fluxion error connecting to server")
		return err
	}
	defer fluxion.Close()

	//	Tell flux to cancel the job id
	fluxionCtx, cancel := context.WithTimeout(context.Background(), 200*time.Second)
	defer cancel()

	// Prepare the request to cancel
	// https://github.com/flux-framework/flux-sched/blob/master/resource/reapi/bindings/go/src/fluxcli/reapi_cli.go#L226
	request := &pb.CancelRequest{JobID: fluxID}

	// Assume if there is an error we should try again
	// TODO:(vsoch) How to distinguish between cancel error
	// and possible already cancelled?
	_, err = fluxion.Cancel(fluxionCtx, request)
	if err != nil {
		return fmt.Errorf("[Fluxion] Issue with cancel %s", err)
	}
	wlog.Info("[Fluxion] Successful cancel", "JobID", fluxID)
	return err
}
