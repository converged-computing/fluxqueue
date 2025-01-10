package jobspec

import (
	corev1 "k8s.io/api/core/v1"
)

// https://github.com/kubernetes/kubectl/blob/master/pkg/describe/describe.go#L4211-L4213
type Resources struct {
	Cpu     int32
	Memory  int64
	Gpu     int64
	Storage int64
	Labels  []string
}

// GeneratePodResources resturns resources for a pod, which can
// be used to populate the flux JobSpec
func GeneratePodResources(containers []corev1.Container) *Resources {

	// We will sum cpu and memory across containers
	// For GPU, we could make a more complex jobspec, but for now
	// assume one container is representative for GPU needed.
	resources := Resources{}

	for _, container := range containers {

		// Add on Cpu, Memory, GPU from container requests
		// This is a limited set of resources owned by the pod
		resources.Cpu += int32(container.Resources.Requests.Cpu().Value())
		resources.Memory += container.Resources.Requests.Memory().Value()
		resources.Storage += container.Resources.Requests.StorageEphemeral().Value()

		// We assume that a pod (node) only has access to the same GPU
		gpus, ok := container.Resources.Limits["nvidia.com/gpu"]
		if ok {
			resources.Gpu += gpus.Value()
		}
	}

	// If we have zero cpus, assume 1
	if resources.Cpu == 0 {
		resources.Cpu = 1
	}
	return &resources
}
