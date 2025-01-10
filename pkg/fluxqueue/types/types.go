package types

// EnqueueStatus is returned by the queue to determine what action to take
type EnqueueStatus int

const (
	// If a pod is already in provisional, group provisional, and pending
	JobEnqueueSuccess EnqueueStatus = iota + 1

	// The pod has already been moved into pending (and submit, but not complete)
	// and we do not accept new pods for the group
	JobAlreadyInPending

	// The pod is invalid (podspec cannot serialize, etc) and should be discarded
	JobInvalid

	// Unknown means some other error happened (usually not related to pod)
	Unknown
)

// Job Database Model we are retrieving for jobs
// This should map back into the FluxJob
type JobModel struct {
	JobSpec     string `db:"jobspec"`
	FluxJobName string `db:"flux_job_name"`
	Name        string `db:"name"`
	Namespace   string `db:"namespace"`
	Type        string `db:"type"`
	Reservation int32  `db:"reservation"`
	Duration    int32  `db:"duration"`
	Size        int32  `db:"size"`
}
