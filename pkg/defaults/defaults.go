package defaults

var (
	KubernetesJsonGraphFormat = "/kubecluster.json"
	SchedulingGateName        = "fluxqueue"
	SchedulerName             = "FluxionScheduler"

	FluxJobIdLabel     = "fluxqueue/jobid"
	NodesLabel         = "fluxqueue/fluxion-nodes"
	SeenLabel          = "fluxqueue.seen"
	UnschedulableLabel = "fluxqueue/unschedulable"
)
