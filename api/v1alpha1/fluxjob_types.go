/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// "Enum" to determine what we are wrapping
type JobWrapped int
type SubmitStatus string

const (
	JobWrappedPod JobWrapped = iota
	JobWrappedJob
	JobWrappedDeployment
	JobWrappedReplicaSet
	JobWrappedStatefulSet

	SubmitStatusNew    SubmitStatus = "statusNew"
	SubmitStatusSubmit SubmitStatus = "statusSubmit"
	SubmitStatusError  SubmitStatus = "statusError"
	SubmitStatusCancel SubmitStatus = "statusCancel"
)

// String returns the stringified type
func (w JobWrapped) String() string {
	return fmt.Sprintf("%d", w)
}

// GetJobName creates a job name that appends the type
func GetJobName(jobType JobWrapped, name string) string {
	switch jobType {
	case JobWrappedJob:
		return fmt.Sprintf("%s-job", name)
	case JobWrappedPod:
		return fmt.Sprintf("%s-pod", name)
	case JobWrappedDeployment:
		return fmt.Sprintf("%s-deployment", name)
	case JobWrappedReplicaSet:
		return fmt.Sprintf("%s-replicaset", name)
	case JobWrappedStatefulSet:
		return fmt.Sprintf("%s-statefulset", name)
	}
	return fmt.Sprintf("%s-unknown", name)
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FluxJobSpec defines the desired state of FluxJob
// A user is not expected to create these -
//  1. A job or similar will be submit
//  2. It will be intercepted by a webhook here
//  3. The webhook will generate the object
//  4. The job will go into the fluxqueue
//  5. When scheduled, it gets sent with an exact node assignment
//     to the custom scheduler plugin.
//  6. Cleanup will need to be handled
//
// A FluxJob is a mapping of a Kubernetes abstraction (e.g., job)
// into a Flux JobSpec, one that Fluxion can digest.
type FluxJobSpec struct {

	// JobSpec is the Flux jobspec
	// +optional
	JobSpec string `json:"jobspec,omitempty"`

	// Type of object that is wrapped
	// +optional
	Type JobWrapped `json:"type"`

	// Original name of the job
	// +optional
	Name string `json:"name"`

	// Object is the underlying pod/job/object specification
	// This currently is assumed that one job has equivalent pods under it
	// +optional
	Object []byte `json:"object,omitempty"`

	// Duration is the maximum runtime of the job
	// +optional
	Duration int32 `json:"duration,omitempty"`

	// If true, we are allowed to ask fluxion for
	// a reservation
	// +optional
	Reservation bool `json:"reservation,omitempty"`

	// Slots needed for the job
	// +optional
	Slots int32 `json:"nodes"`

	// Cores per pod (slot)
	// +optional
	Cores int32 `json:"cores"`

	// Resources assigned
	// +optional
	Resources Resources `json:"resources"`
}

type Resources struct {

	// Nodes assigned to the job
	// +optional
	Nodes []string `json:"nodes"`
}

// FluxJobStatus defines the observed state of FluxJob
type FluxJobStatus struct {
	SubmitStatus SubmitStatus `json:"submitStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// FluxJob is the Schema for the fluxjobs API
type FluxJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FluxJobSpec   `json:"spec,omitempty"`
	Status FluxJobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FluxJobList contains a list of FluxJob
type FluxJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FluxJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FluxJob{}, &FluxJobList{})
}
