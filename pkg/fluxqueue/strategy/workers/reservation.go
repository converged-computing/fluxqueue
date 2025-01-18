package workers

import (
	"context"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"k8s.io/client-go/rest"

	"github.com/converged-computing/fluxion/pkg/client"
	pb "github.com/converged-computing/fluxion/pkg/fluxion-grpc"
	"github.com/converged-computing/fluxqueue/pkg/fluxqueue/queries"
	"github.com/converged-computing/fluxqueue/pkg/fluxqueue/types"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

// The Reservation worker clears reservations
type ReservationArgs struct{}

// The cleanup workers cleans up a reservation (issuing cancel)
func (args ReservationArgs) Kind() string { return "reservation" }

type ReservationWorker struct {
	river.WorkerDefaults[ReservationArgs]
}

// NewJobWorker returns a new job worker with a Fluxion client
func NewReservationWorker(cfg rest.Config) (*ReservationWorker, error) {
	worker := ReservationWorker{}
	return &worker, nil
}

// Work performs the clearing of reservations
func (w ReservationWorker) Work(ctx context.Context, job *river.Job[ReservationArgs]) error {
	wlog.Info("Clearing reservations")
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer pool.Close()

	rRows, err := pool.Query(ctx, queries.GetReservationsQuery)
	if err != nil {
		return err
	}
	defer rRows.Close()

	// Connect to the Fluxion service.
	// TODO what happens on error? We don't clear reservations, retried later?
	fluxion, err := client.NewClient("127.0.0.1:4242")
	if err != nil {
		wlog.Error(err, "Fluxion error connecting to server")
		return err
	}
	defer fluxion.Close()

	//	Tell flux to cancel the job id
	fluxionCtx, cancel := context.WithTimeout(context.Background(), 200*time.Second)
	defer cancel()

	rows, err := pgx.CollectRows(rRows, pgx.RowToStructByName[types.ReservationModel])
	if err != nil {
		wlog.Error(err, "collecting rows for pending jobs")
		return err
	}
	for _, item := range rows {
		request := &pb.CancelRequest{JobID: item.JobId}
		_, err = fluxion.Cancel(fluxionCtx, request)
		if err != nil {
			wlog.Info("Issue cancelling reservation %d for %s: %s", item.JobId, item.Name, err)
			continue
		}
		// Only delete reservation if we cancelled in fluxion
		_, err = pool.Exec(ctx, queries.DeleteSingleReservationsQuery, item.JobId)
		if err != nil {
			wlog.Info("Error deleting reservation from database", "JobID", item.JobId, "Name", item.Name)
			return err
		}
	}
	return nil
}
