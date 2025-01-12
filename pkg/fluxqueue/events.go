package fluxqueue

import (
	"context"
	"os"

	"github.com/converged-computing/fluxqueue/pkg/fluxqueue/queries"
	"github.com/jackc/pgx/v5/pgxpool"
)

// deleteFromPending ensures that the job is removed from pending
// This happens on a job that is impossible to schedule, or one that is successfully running
func deleteFromPending(namespace, name string) error {

	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		qlog.Info("Issue creating new pool", "Error", err.Error())
		return err
	}
	defer pool.Close()

	_, err = pool.Exec(context.Background(), queries.DeleteFromPendingQuery, name, namespace)
	if err != nil {
		qlog.Info("Error deleting job from pending queue", "Namespace", namespace, "Name", name)
		return err
	}
	qlog.Info("Removed job from pending", "Namespace", namespace, "Name", name)
	return nil
}
