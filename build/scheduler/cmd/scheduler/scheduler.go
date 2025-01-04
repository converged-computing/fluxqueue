package main

import (
	"os"

	"github.com/converged-computing/fluxqueue/pkg/scheduler"
	"k8s.io/component-base/cli"
	_ "k8s.io/component-base/logs/json/register" // for JSON log format registration
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	_ "k8s.io/component-base/metrics/prometheus/version" // for version metric registration
	"k8s.io/kubernetes/cmd/kube-scheduler/app"
)

func main() {
	// Fluxnetes is actually added as an in-tree plugin here
	command := app.NewSchedulerCommand(
		app.WithPlugin(scheduler.Name, scheduler.New),
	)
	code := cli.Run(command)
	os.Exit(code)
}
