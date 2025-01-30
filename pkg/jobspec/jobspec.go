package jobspec

import (
	v1 "github.com/compspec/jobspec-go/pkg/jobspec/v1"
)

// NewJobSpec generates a jobspec for some number of slots in a cluster
// We associate each "slot" with a pod, so the request asks for a specific number of cpu.
// We also are assuming now that each pod is equivalent, so slots are equivalent.
// If we want to change this, we will need an ability to define slots of different types.
func NewJobspec(name string, command []string, resources *Resources) (*v1.Jobspec, error) {

	// This is creating the resources for the slot Cores are always set to minimally 1
	slotSpec := newSlotSpec(resources)

	// Create the top level resources spec (with a slot)
	rSpec := []v1.Resource{
		{
			Type:  "slot",
			Count: resources.Count,
			Label: "default",
			With:  slotSpec,
		},
	}

	// Create the task spec
	tasks := []v1.Tasks{
		{
			Command: command,
			Slot:    "default",
			Count:   v1.Count{PerSlot: 1},
		},
	}

	// Start preparing the spec
	spec := v1.Jobspec{
		Version:   1,
		Resources: rSpec,
		Tasks:     tasks,
	}

	// Attributes are for the system, we aren't going to add them yet
	// attributes:
	// system:
	//   duration: 3600.
	//   cwd: "/home/flux"
	//   environment:
	// 	HOME: "/home/flux"
	// This is verison 1 as defined by v1 above
	return &spec, nil
}

// newSlotSpec creates a spec for one slot, which is one pod (a set of containers)
func newSlotSpec(resources *Resources) []v1.Resource {
	slotSpec := []v1.Resource{
		{Type: "core", Count: resources.Slot.Cpu},
	}
	// If we have memory or gpu specified, they are appended
	if resources.Slot.Gpu > 0 {
		slotSpec = append(slotSpec, v1.Resource{Type: "gpu", Count: int32(resources.Slot.Gpu)})
	}
	if resources.Slot.Memory > 0 {
		toMB := resources.Slot.Memory >> 20
		slotSpec = append(slotSpec, v1.Resource{Type: "memory", Count: int32(toMB)})
	}
	return slotSpec
}
