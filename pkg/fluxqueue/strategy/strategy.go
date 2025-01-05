package strategy

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

// Interface for a queue strategy
// A queue strategy both controls a work function (what the worker does, and arguments)
// Along with how to orchestrate the last part of the schedule loop, with post submit.

type QueueStrategy interface {
	Name() string

	AddWorkers(*river.Workers) error

	// Schedule takes pending pods and submits to fluxion
	Schedule(context.Context, *pgxpool.Pool, int32) ([]river.InsertManyParams, error)
	PostSubmit(context.Context, *pgxpool.Pool, *river.Client[pgx.Tx]) error

	// Return metadata about the strategy for the Queue to know
	GetReservationDepth() int32
}
