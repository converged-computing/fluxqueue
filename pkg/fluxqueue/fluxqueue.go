package fluxqueue

import (
	"strings"
)

var (
	CancelledState = "cancelled"
	CleanupQueue   = "cleanup"
	Unsatisfiable  = "unsatisfiable"
)

// JobResult serializes a result from Fluxnetes in the scheduler back to metadata
type JobResult struct {
	JobID     int32  `json:"jobid"`
	Nodes     string `json:"nodes"`
	PodID     string `json:"podid"`
	PodSpec   string `json:"podspec"`
	Names     string `json:"names"`
	Namespace string `json:"namespace"`
	GroupName string `json:"groupName"`
}

func (j JobResult) GetNodes() []string {
	return strings.Split(j.Nodes, ",")
}

func (j JobResult) GetPodNames() []string {
	return strings.Split(j.Names, ",")
}
