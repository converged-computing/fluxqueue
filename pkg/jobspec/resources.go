package jobspec

import (
	corev1 "k8s.io/api/core/v1"
)

// https://github.com/kubernetes/kubectl/blob/master/pkg/describe/describe.go#L4211-L4213
// We want each slot to coincide with one pod. We will ask flux for all N slots to schedule,
// and then based on the response we get back, assign specific nodes. This currently
// assumes slots are each the same, but this is subject to change.
// QUESTION: what to add here for labels?
type Resources struct {
	Labels []string
	Slot   Slot
	Count  int32
}

type Slot struct {
	Cpu     int32
	Memory  int64
	Gpu     int64
	Storage int64
}

// GeneratePodResources returns resources for a pod, which can
// be used to populate the flux JobSpec.
func GeneratePodResources(containers []corev1.Container, slots int32) *Resources {

	// We will sum cpu and memory across containers
	// For GPU, we could make a more complex jobspec, but for now
	// assume one container is representative for GPU needed.
	resources := Resources{Slot: Slot{}, Count: slots}

	for _, container := range containers {

		// Add on Cpu, Memory, GPU from container requests
		// This is a limited set of resources owned by the pod
		resources.Slot.Cpu += int32(container.Resources.Requests.Cpu().Value())
		resources.Slot.Memory += container.Resources.Requests.Memory().Value()
		resources.Slot.Storage += container.Resources.Requests.StorageEphemeral().Value()

		// We assume that a pod (node) only has access to the same GPU
		gpus, ok := container.Resources.Limits["nvidia.com/gpu"]
		if ok {
			resources.Slot.Gpu += gpus.Value()
		}
	}

	// If we have zero cpus, assume 1
	if resources.Slot.Cpu == 0 {
		resources.Slot.Cpu = 1
	}
	return &resources
}
