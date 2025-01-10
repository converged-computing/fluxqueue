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
	klog "k8s.io/klog/v2"

	api "github.com/converged-computing/fluxqueue/api/v1alpha1"
	"github.com/converged-computing/fluxqueue/pkg/fluxqueue/queries"
	strategies "github.com/converged-computing/fluxqueue/pkg/fluxqueue/strategy"
	"github.com/converged-computing/fluxqueue/pkg/fluxqueue/types"
)

const (
	queueMaxWorkers = 10
	mutexLocked     = 1
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

			// Cleanup queue is only for cancel
			"cancel_queue": {MaxWorkers: queueMaxWorkers},
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

	// Validates reservation depth
	// TODO(vsoch) allow -1 to disable
	depth := strategy.GetReservationDepth()
	if depth < 0 {
		return nil, fmt.Errorf("reservation depth of a strategy must be >= -1")
	}

	queue := Queue{
		riverClient:      riverClient,
		Pool:             pool,
		Strategy:         strategy,
		Context:          ctx,
		ReservationDepth: depth,
	}
	queue.setupEvents()
	return &queue, nil
}

// StopQueue creates a client (without calling start) only intended to
// issue stop, so we can leave out workers and queue from Config
func (q *Queue) Stop(ctx context.Context) error {
	if q.riverClient != nil {
		return q.riverClient.Stop(ctx)
	}
	return nil
}

// We can tell how a job runs via events
// setupEvents create subscription channels for each event type
func (q *Queue) setupEvents() {

	// Subscribers tell the River client the kinds of events they'd like to receive.
	// We add them to a listing to be used by Kubernetes. These can be subscribed
	// to from elsewhere too (anywhere). Note that we are not subscribing to failed
	// or snoozed, because they right now mean "allocation not possible" and that
	// is too much noise.
	c, trigger := q.riverClient.Subscribe(
		river.EventKindJobCompleted,
		river.EventKindJobCancelled,
		// Be careful about re-enabling failed, that means you'll get a notification
		// for every job that isn't allocated.
		//		river.EventKindJobFailed, (retryable)
		//		river.EventKindJobSnoozed, (scheduled later, not used yet)
	)
	q.EventChannel = &QueueEvent{Function: trigger, Channel: c}
}

// Common queue / database functions across strategies!
// GetFluxID returns the flux ID, and -1 if not found (deleted)
/*func (q *Queue) GetFluxID(namespace, groupName string) (int64, error) {
	var fluxID int32 = -1
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		klog.Errorf("Issue creating new pool %s", err)
		return int64(fluxID), err
	}
	defer pool.Close()
	result := pool.QueryRow(context.Background(), queries.GetFluxID, groupName, namespace)
	err = result.Scan(&fluxID)

	// This can simply mean it was already deleted from pending
	if err != nil {
		klog.Infof("Error retrieving FluxID for %s/%s: %s", groupName, namespace, err)
		return int64(-1), err
	}
	return int64(fluxID), err
}*/

// GetInformer returns the pod informer to run as a go routine
// TODO this should be done through the operator, and then triggered here
func (q *Queue) GetInformer() error { //cache.SharedIndexInformer {
	return nil
	//	return cache.SharedIndexInformer{}

	// Performance improvement when retrieving list of objects by namespace or we'll log 'index not exist' warning.
	//	podsInformer := q.Handle.SharedInformerFactory().Core().V1().Pods().Informer()
	//	podsInformer.AddIndexers(cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	// Event handlers to call on update/delete for cleanup
	//	podsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
	//		UpdateFunc: q.UpdatePodEvent,
	//		DeleteFunc: q.DeletePodEvent,
	//	})
	//	return podsInformer
}

// Enqueue a new job to the pending queue, which is just a database table
// This functionality is shared across strategies. We add the jobs
// as they are submit to Kubernetes
func (q *Queue) Enqueue(spec *api.FluxJob) (types.EnqueueStatus, error) {

	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		klog.Errorf("Issue creating new pool %s", err)
		return types.Unknown, err
	}
	defer pool.Close()

	// First check - a job that is already in pending (unique by flux job name and namespace)
	// is not allowed to be submit again. The job is either waiting or running.
	result, err := pool.Exec(context.Background(), queries.IsPendingQuery, spec.Name, spec.Namespace)
	if err != nil {
		klog.Infof("Error checking if job %s/%s is in pending queue", spec.Namespace, spec.Name)
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

// Schedule assesses jobs in pending to be sent to Fluxion.
// The strategy determines which get chosen (e.g., scheduled at time or priority)
// TODO: we need another mechanism to kick off the queue
func (q *Queue) Schedule() error {

	// Don't kick off another loop if we are already in one
	if q.IsInScheduleLoop() {
		return nil
	}
	// Acquire the lock
	q.lock.Lock()         // Acquire the lock
	defer q.lock.Unlock() // Release the lock when the function exits

	// This generates a batch of jobs to send to ask Fluxion for nodes
	batch, err := q.Strategy.Schedule(q.Context, q.Pool, q.ReservationDepth)
	if err != nil {
		return err
	}

	if len(batch) > 0 {
		count, err := q.riverClient.InsertMany(q.Context, batch)
		if err != nil {
			return err
		}
		klog.Info(count)
	}

	// Post submit functions
	return nil // q.Strategy.PostSubmit(q.Context, q.Pool, q.riverClient)
}

// GetCreationTimestamp returns the creation time of a podGroup or a pod in seconds (time.MicroTime)
// We either get this from the pod itself (if size 1) or from the database
/*func (q *Queue) GetCreationTimestamp(pod *corev1.Pod, groupName string) (metav1.MicroTime, error) {

	// First see if we've seen the group before, the creation times are shared across a group
	ts := metav1.MicroTime{}

	// This query will fail if there are no rows (the podGroup is not known in the namespace)
	row := q.Pool.QueryRow(context.Background(), queries.GetTimestampQuery, groupName, pod.Namespace)
	err := row.Scan(&ts)
	if err == nil {
		klog.Info("Creation timestamp is", ts)
		return ts, err
	}
	return groups.GetPodCreationTimestamp(pod), nil
}*/
