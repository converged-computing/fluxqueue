package fluxqueue

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivershared/util/slogutil"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"

	api "github.com/converged-computing/fluxqueue/api/v1alpha1"
	"github.com/converged-computing/fluxqueue/pkg/fluxqueue/queries"
	strategies "github.com/converged-computing/fluxqueue/pkg/fluxqueue/strategy"
	"github.com/converged-computing/fluxqueue/pkg/fluxqueue/strategy/workers"
	"github.com/converged-computing/fluxqueue/pkg/fluxqueue/types"
)

const (
	queueMaxWorkers = 10
	mutexLocked     = 1
)

var (
	qlog = ctrl.Log.WithName("queue")
)

// Queue holds handles to queue database and event handles
// The database Pool also allows interacting with the pods table (database.go)
type Queue struct {
	Pool         *pgxpool.Pool
	riverClient  *river.Client[pgx.Tx]
	EventChannel *QueueEvent
	Strategy     strategies.QueueStrategy

	// IMPORTANT: subscriptions need to use same context
	// that client submit them uses
	Context context.Context

	// Lock the queue during a scheduling cycle
	lock       sync.Mutex
	RESTClient rest.Interface

	// Reservation depth:
	// Less than -1 is invalid (and throws error)
	// -1 means no reservations are done
	// 0 means reservations are done, but no depth set
	// Anything greater than 0 is a reservation value
	ReservationDepth int32
}

type ChannelFunction func()

// QueueEvent holds the channel and defer function
type QueueEvent struct {
	Channel  <-chan *river.Event
	Function ChannelFunction
}

// IsInScheduleLoop looks at the state of the mutex to determine if it is locked
// If it's locked, we are in a loop and return true. Otherwise, false.
func (q *Queue) IsInScheduleLoop() bool {

	// the mutex object has a private variable, state,
	// that will be 1 when locked
	state := reflect.ValueOf(&q.lock).Elem().FieldByName("state")
	return state.Int()&mutexLocked == mutexLocked
}

// NewQueue starts a new queue with a river client
func NewQueue(ctx context.Context, cfg rest.Config) (*Queue, error) {
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}

	// The default strategy now mirrors what fluence with Kubernetes does
	// This can eventually be customizable. We provide the pool to the
	// strategy because it also manages the provisional queue.
	strategy := strategies.EasyBackfill{}
	workers := river.NewWorkers()

	// Each strategy has its own worker type
	err = strategy.AddWorkers(workers, cfg)
	if err != nil {
		return nil, err
	}

	riverClient, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		// Change the verbosity of the logger here
		Logger: slog.New(&slogutil.SlogMessageOnlyHandler{Level: slog.LevelWarn}),
		Queues: map[string]river.QueueConfig{

			// Default queue handles job allocation
			river.QueueDefault: {MaxWorkers: queueMaxWorkers},

			// TODO do we have control of deletions?
			// Cleanup queue is typically for cancel
			"cleanup_queue": {MaxWorkers: queueMaxWorkers},
		},
		Workers: workers,
	})
	if err != nil {
		return nil, err
	}

	// Create the queue and setup events for it
	err = riverClient.Start(ctx)
	if err != nil {
		return nil, err
	}

	// Validates reservation depth, if a strategy supports it
	// A value of -1 means disabled
	depth := strategy.GetReservationDepth()
	if depth < 0 && depth != -1 {
		return nil, fmt.Errorf("reservation depth of a strategy must be >= -1")
	}
	return &Queue{
		riverClient:      riverClient,
		Pool:             pool,
		Strategy:         strategy,
		Context:          ctx,
		ReservationDepth: depth,
	}, nil
}

// StopQueue creates a client (without calling start) only intended to
// issue stop, so we can leave out workers and queue from Config
func (q *Queue) Stop(ctx context.Context) error {
	if q.riverClient != nil {
		return q.riverClient.Stop(ctx)
	}
	return nil
}

// Enqueue a new job to the pending queue, which is just a database table
// This functionality is shared across strategies. We add the jobs
// as they are submit to Kubernetes
func (q *Queue) Enqueue(spec *api.FluxJob) (types.EnqueueStatus, error) {

	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		qlog.Error(err, "creating new pool")
		return types.Unknown, err
	}
	defer pool.Close()

	// First check - a job that is already in pending (unique by flux job name and namespace)
	// is not allowed to be submit again. The job is either waiting or running.
	result, err := pool.Exec(context.Background(), queries.IsPendingQuery, spec.Name, spec.Namespace)
	if err != nil {
		qlog.Info("Error checking if job is in pending queue", "Namespace", spec.Namespace, "Name", spec.Name)
		return types.Unknown, err
	}
	if strings.Contains(result.String(), "INSERT 1") {
		return types.JobAlreadyInPending, nil
	}

	// We use the CRD creation timestamp
	// ts := &pgtype.Timestamptz{Time: spec.ObjectMeta.CreationTimestamp.Time}

	// Reservation needs to be integer
	reservation := 0
	if spec.Spec.Reservation {
		reservation = 1
	}

	// Insert into pending queue, only if name and namespace don't exist.
	// TODO figure out how to do timestamp. By default it will use inserted.
	_, err = pool.Exec(context.Background(), queries.InsertIntoPending,
		spec.Spec.JobSpec,
		spec.Name, // flux job name
		spec.Namespace,
		spec.Spec.Name, // original job name
		spec.Spec.Type,
		reservation,
		spec.Spec.Duration,
		spec.Spec.Nodes,
	)

	// If unknown, we won't give status submit, and it should requeue to try again
	if err != nil {
		return types.Unknown, err
	}
	return types.JobEnqueueSuccess, nil
}

// Cleanup deletes (cancels) job with fluxion
func (q *Queue) Cleanup(fluxids []int64) error {
	return q.Strategy.Cleanup(q.Context, q.Pool, q.riverClient, fluxids)
}

// Schedule assesses jobs in pending to be sent to Fluxion.
// The strategy determines which get chosen (e.g., scheduled at time or priority)
func (q *Queue) Schedule() error {

	// Don't kick off another loop if we are already in one
	if q.IsInScheduleLoop() {
		return nil
	}
	// Acquire the lock on the queue
	q.lock.Lock()
	defer q.lock.Unlock()

	// This generates a batch of jobs to send to ask Fluxion for nodes
	batch, err := q.Strategy.Schedule(q.Context, q.Pool, q.ReservationDepth)
	if err != nil {
		return err
	}

	// Run each job task to schedule nodes for it (or ask again later)
	if len(batch) > 0 {
		_, err := q.riverClient.InsertMany(q.Context, batch)
		if err != nil {
			return err
		}

		// Once we know they are inserted, they need to be deleted from pending
		// or else they might be scheduled again. We can assume all strategies
		// use the common worker args for submit to Fluxion.
		for _, job := range batch {
			args := job.Args.(workers.JobArgs)
			deleteFromPending(args.Namespace, args.Name)
		}
	}

	// Post submit functions for a queue strategy
	return q.Strategy.PostSubmit(q.Context, q.Pool, q.riverClient)
}
