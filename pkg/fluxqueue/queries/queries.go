package queries

const (

	// Pending Queue queries
	// We always check if a job is in pending before Enqueue, because if so, we aren't allowed to modify / add again
	// In practice with Kubernetes this should not happen, as the API server would detect the object already exists
	IsPendingQuery = "select * from pending_queue where name = $1 and namespace = $2;"

	// Insert into pending queue (assumes after above query, we've checked it does not exist)
	InsertIntoPending = "insert into pending_queue (jobspec, flux_job_name, namespace, name, type, reservation, duration, size, cores) values ($1, $2, $3, $4, $5, $6, $7, $8, $9);"
	// TODO add back created_at

	// We remove from pending to allow another group submission of the same name on cleanup
	DeleteFromPendingQuery = "delete from pending_queue where name=$1 and namespace=$2;"

	// Easy Queries to get jobs
	// Select jobs based on creation timestamp
	SelectPendingByCreation = "select jobspec, name, flux_job_name, namespace, type, reservation, duration, size, cores from pending_queue order by created_at desc;"

	// Reservations
	AddReservationQuery           = "insert into reservations (name, flux_id) values ($1, $2);"
	DeleteReservationsQuery       = "truncate reservations; delete from reservations;"
	DeleteSingleReservationsQuery = "delete from reservations where flux_id=$1;"
	GetReservationsQuery          = "select (name, flux_id) from reservations;"
)
