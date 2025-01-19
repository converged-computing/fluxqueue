package strategy

import (
	"context"

	"k8s.io/client-go/rest"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"

	"github.com/converged-computing/fluxqueue/pkg/fluxqueue/defaults"
	"github.com/converged-computing/fluxqueue/pkg/fluxqueue/queries"
	"github.com/converged-computing/fluxqueue/pkg/fluxqueue/strategy/workers"
	work "github.com/converged-computing/fluxqueue/pkg/fluxqueue/strategy/workers"
	"github.com/converged-computing/fluxqueue/pkg/fluxqueue/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	elog = ctrl.Log.WithName("easy")
)

// Easy with Backfill
// Schedule jobs that come in first, but allow smaller jobs to fill in
type EasyBackfill struct{}

// Name returns shortened "first come first serve"
func (EasyBackfill) Name() string {
	return "easy"
}

// GetReservationDepth returns the depth of 1
// easy allows for one reservation.
func (EasyBackfill) GetReservationDepth() int32 {
	return int32(1)
}

// ReservationModel makes it easy to convert a rows response
// into the data structures here
type ReservationModel struct {
	GroupName string `db:"group_name"`
	FluxID    int64  `db:"flux_id"`
}

// AddtWorkers adds the worker for the queue strategy
// job worker: a queue to submit jobs to fluxion
// cleanup worker: a queue to cleanup
func (EasyBackfill) AddWorkers(workers *river.Workers, cfg rest.Config) error {

	// These workers are in the default (fluxion) queue with one worker
	jobWorker, err := work.NewJobWorker(cfg)
	if err != nil {
		return err
	}
	cleanupWorker, err := work.NewCleanupWorker(cfg)
	if err != nil {
		return err
	}
	reservationWorker, err := work.NewReservationWorker(cfg)
	if err != nil {
		return err
	}

	// These workers can be run concurrently (>1 worker)
	ungateWorker, err := work.NewUngateWorker(cfg)
	if err != nil {
		return err
	}
	river.AddWorker(workers, jobWorker)
	river.AddWorker(workers, cleanupWorker)
	river.AddWorker(workers, reservationWorker)
	river.AddWorker(workers, ungateWorker)
	return nil
}

// Cleanup submit a job to call fluxion
func (s EasyBackfill) Cleanup(
	ctx context.Context,
	pool *pgxpool.Pool,
	riverClient *river.Client[pgx.Tx],
	fluxids []int64) error {

	// Shared insertOpts.
	// Tags can eventually be specific to job attributes, queues, etc.
	// This also sets the queue to the cleanup queue
	insertOpts := river.InsertOpts{
		MaxAttempts: defaults.MaxCancelAttempts,
		Tags:        []string{s.Name()},
		Queue:       "cleanup_queue",
	}

	// A cleanup worker issues a cancel request to fluxion
	cancelJobs := []work.CleanupArgs{}
	for _, fluxid := range fluxids {
		cancelJobs = append(cancelJobs, work.CleanupArgs{FluxID: fluxid})
	}

	// Prepare batch job for cleanup workers
	batch := []river.InsertManyParams{}
	for _, cleanupArgs := range cancelJobs {
		args := river.InsertManyParams{Args: cleanupArgs, InsertOpts: &insertOpts}
		batch = append(batch, args)
	}

	// Insert the cleanup jobs
	if len(batch) > 0 {
		_, err := riverClient.InsertMany(ctx, batch)
		if err != nil {
			return err
		}
	}
	return nil
}

// Schedule moves pending jobs into being scheduled, which means doing
// some kind of sort, asking Fluxion, and then alerting the operator when a job
// is scheduled. When this happens, it us unsuspended / ungated to move to the
// fluxion scheduler plugin, where it also is given the exact node assignment.
// We return a listing of river.JobArgs (JobArgs here) to be submit with batch.

// Easy strategy:
// In this case it is first come first serve - we just sort based on the timestamp
// and add them to the worker queue. The job here is a request to AskFlux, and the
// result of that determines if the job is actually scheduled for Kubernetes.
// We don't limit the number of jobs from pending - we go through them all.
// Other strategies can handle AskFlux and submitting batch differently.
func (s EasyBackfill) Schedule(
	ctx context.Context,
	pool *pgxpool.Pool,
	reservationDepth int32,
) ([]river.InsertManyParams, error) {

	// Serialize pending queue into jobs for river
	// Each will run AskFlux (to fluxion) to attempt schedule
	jobs, err := s.ReadyJobs(ctx, pool)
	if err != nil {
		elog.Error(err, "Issue FCFS with backfill querying for ready groups")
		return nil, err
	}

	// Shared insertOpts.
	// Tags can eventually be specific to job attributes, queues, etc.
	// The default queue is used for allocation requests
	insertOpts := river.InsertOpts{
		MaxAttempts: defaults.MaxAttempts,
		Tags:        []string{s.Name()},
		Queue:       river.QueueDefault,
	}

	// The easy strategy reserves up to a reservation depth
	// https://riverqueue.com/docs/batch-job-insertion
	// Note: this is how to eventually add Priority (1-4, 4 is lowest)
	// And we can customize other InsertOpts. Of interest is Pending:
	// https://github.com/riverqueue/river/blob/master/insert_opts.go#L35-L40
	// Note also that ScheduledAt can be used to ask fluxion in the future
	batch := []river.InsertManyParams{}
	for i, jobArgs := range jobs {
		args := river.InsertManyParams{Args: jobArgs, InsertOpts: &insertOpts}

		// Reservation depth of -1 is disabling reservations
		if reservationDepth != -1 && int32(i) < reservationDepth {
			jobArgs.Reservation = 1
		}
		batch = append(batch, args)
	}
	return batch, nil
}

// getReadyGroups gets groups that are ready for moving from provisional to pending
// We also save the pod names so we can assign (bind) to nodes later
func (s EasyBackfill) ReadyJobs(ctx context.Context, pool *pgxpool.Pool) ([]workers.JobArgs, error) {

	// First retrieve the group names that are the right size
	rows, err := pool.Query(ctx, queries.SelectPendingByCreation)
	if err != nil {
		elog.Error(err, "selecting pending jobs")
		return nil, err
	}
	defer rows.Close()

	models, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.JobModel])
	if err != nil {
		elog.Error(err, "collecting rows for pending jobs")
		return nil, err
	}

	// Collect rows into map, and then slice of jobs
	jobs := []workers.JobArgs{}

	// Collect rows into single result
	for _, model := range models {
		jobArgs := workers.JobArgs{
			Jobspec:     model.JobSpec,
			Name:        model.Name,
			Namespace:   model.Namespace,
			FluxJobName: model.FluxJobName,
			Type:        model.Type,
			Reservation: model.Reservation,
			Size:        model.Size,
			Duration:    model.Duration,
		}
		jobs = append(jobs, jobArgs)
	}
	return jobs, nil
}

// PostSubmit does clearing of reservations
// unlike schedule, we provide the client here to do the work, the reason being
// we might want to check success and do an additional operation (delete).
// This happens after we enqueue a new set, assuming the cycle of workers has also
// moved through the queue. In practice (with async) this might not be true. We
// need to draw a state diagram to figure this out.
func (s EasyBackfill) PostSubmit(
	ctx context.Context,
	pool *pgxpool.Pool,
	riverClient *river.Client[pgx.Tx],
) error {
	return nil
}
