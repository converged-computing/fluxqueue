package scheduler

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	corev1helpers "k8s.io/component-helpers/scheduling/corev1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

var (
	CancelledState = "cancelled"
	CleanupQueue   = "cleanup"
	Unsatisfiable  = "unsatisfiable"
)

// This is a simple scheduler plugin that receives pods that already have a node assignment.
// The assignments are done by fluxion, which is running as a service.
// To ensure consistency, any pod that does not have the fluxion annotation is denied.
// This ensures that the state of the cluster is owned and controlled by fluxion.

type FluxionScheduler struct{}

var (
	_ framework.FilterPlugin    = &FluxionScheduler{}
	_ framework.QueueSortPlugin = &FluxionScheduler{}
)

const (
	Name = "FluxionScheduler"
)

func (fs *FluxionScheduler) Name() string {
	return Name
}

// Less is required for QueueSortPlugin.
// We sort based on priority, init timestamps, and then keys
func (fs *FluxionScheduler) Less(podInfo1, podInfo2 *framework.QueuedPodInfo) bool {
	prio1 := corev1helpers.PodPriority(podInfo1.Pod)
	prio2 := corev1helpers.PodPriority(podInfo2.Pod)
	if prio1 != prio2 {
		return prio1 > prio2
	}
	return getNamespacedName(podInfo1.Pod) < getNamespacedName(podInfo2.Pod)
}

// GetNamespacedName returns the namespaced name.
func getNamespacedName(obj metav1.Object) string {
	return fmt.Sprintf("%v/%v", obj.GetNamespace(), obj.GetName())
}

func (fs *FluxionScheduler) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	return framework.NewStatus(framework.Success)
}

// New returns an empty FluxionScheduler plugin, which only provides a queue sort!
func New(_ context.Context, _ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
	return &FluxionScheduler{}, nil
}
