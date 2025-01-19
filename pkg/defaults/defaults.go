package defaults

var (
	KubernetesJsonGraphFormat = "/kubecluster.json"
	SchedulingGateName        = "fluxqueue"
	SchedulerName             = "FluxionScheduler"

	FluxJobIdLabel     = "fluxqueue/jobid"
	NodesLabel         = "fluxqueue/fluxion-nodes"
	SeenLabel          = "fluxqueue.seen"
	SelectorLabel      = "fluxqueue.selector"
	UnschedulableLabel = "fluxqueue/unschedulable"
)
