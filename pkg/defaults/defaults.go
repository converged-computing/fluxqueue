package defaults

var (
	KubernetesJsonGraphFormat = "/kubecluster.json"
	SchedulingGateName        = "fluxqueue"
	SchedulerName             = "FluxionScheduler"

	SeenLabel      = "fluxqueue.seen"
	NodesLabel     = "fluxqueue/fluxion-nodes"
	FluxJobIdLabel = "fluxqueue/jobid"
)
