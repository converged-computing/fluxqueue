package v1alpha1

import (
	"context"
	"fmt"

	"github.com/converged-computing/fluxqueue/pkg/defaults"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

// EnqueuePod submits a flux job for the pod
func (a *jobReceiver) EnqueuePod(ctx context.Context, pod *corev1.Pod) error {

	// Pods associated with a parent object often don't have a name
	if pod.Name == "" {
		return nil
	}
	logger.Info("Contender pod", "Name", pod.Name, "Namespace", pod.Namespace)

	// Check if we have a label from another abstraction. E.g.,, a job that includes
	// pods should not schedule the pods twice
	if pod.ObjectMeta.Labels == nil {
		pod.ObjectMeta.Labels = map[string]string{}
	}

	// We've already seen this pod elsewhere, exit without re-submit
	_, ok := pod.ObjectMeta.Labels[defaults.SeenLabel]
	if ok {
		return nil
	}
	// Mark the pod now as seen
	pod.ObjectMeta.Labels[defaults.SeenLabel] = "yes"

	// Add scheduling gate to the pod
	if pod.Spec.SchedulingGates == nil {
		pod.Spec.SchedulingGates = []corev1.PodSchedulingGate{}
	}

	fluxqGate := corev1.PodSchedulingGate{Name: defaults.SchedulingGateName}
	pod.Spec.SchedulingGates = append(pod.Spec.SchedulingGates, fluxqGate)

	// Ensure the pod gets scheduled with fluxion scheduler
	pod.Spec.SchedulerName = defaults.SchedulerName
	logger.Info("received pod and added gate", "Name", pod.Name)

	return SubmitFluxJob(
		ctx,
		JobWrappedPod,
		pod.Name,
		pod.Namespace,
		1,
		pod.Spec.Containers,
	)
}

// EnqueueJob suspends the job and creates a FluxJob
func (a *jobReceiver) EnqueueJob(ctx context.Context, job *batchv1.Job) error {
	logger.Info("Contender job", "Name", job.Name, "Namespace", job.Namespace)

	// Add labels to the pod so they don't trigger another submit/schedule
	// Check if we have a label from another abstraction. E.g.,, a job that includes
	// pods should not schedule the pods twice
	if job.Spec.Template.ObjectMeta.Labels == nil {
		job.Spec.Template.ObjectMeta.Labels = map[string]string{}
	}
	if job.ObjectMeta.Labels == nil {
		job.ObjectMeta.Labels = map[string]string{}
	}

	// Cut out early if we are getting hit again
	_, ok := job.ObjectMeta.Labels[defaults.SeenLabel]
	if ok {
		return nil
	}
	job.ObjectMeta.Labels[defaults.SeenLabel] = "yes"
	job.Spec.Template.ObjectMeta.Labels[defaults.SeenLabel] = "yes"
	job.Spec.Template.Spec.SchedulerName = defaults.SchedulerName

	// Suspend the job
	suspended := true
	job.Spec.Suspend = &suspended

	logger.Info("received job and suspended", "Name", job.Name)
	return SubmitFluxJob(
		ctx,
		JobWrappedJob,
		job.Name,
		job.Namespace,
		*job.Spec.Parallelism,
		job.Spec.Template.Spec.Containers,
	)
}

// EnqueueDeployment gates pods that are associated with a deployment
// This is a slightly different design, because we cannot suspend the upper level
// abstraction. Instead, we allow the deployment to be created, but gate the pods
// that belong to it.
func (a *jobReceiver) EnqueueDeployment(ctx context.Context, deployment *appsv1.Deployment) error {
	logger.Info("Contender deployment", "Name", deployment.Name, "Namespace", deployment.Namespace)

	// Add labels to the pod so they don't trigger another submit/schedule
	// Check if we have a label from another abstraction. E.g.,, a job that includes
	// pods should not schedule the pods twice
	if deployment.Spec.Template.ObjectMeta.Labels == nil {
		deployment.Spec.Template.ObjectMeta.Labels = map[string]string{}
	}
	if deployment.ObjectMeta.Labels == nil {
		deployment.ObjectMeta.Labels = map[string]string{}
	}

	// Cut out early if we are getting hit again
	_, ok := deployment.ObjectMeta.Labels[defaults.SeenLabel]
	if ok {
		return nil
	}
	deployment.ObjectMeta.Labels[defaults.SeenLabel] = "yes"
	deployment.Spec.Template.ObjectMeta.Labels[defaults.SeenLabel] = "yes"
	deployment.Spec.Template.Spec.SchedulerName = defaults.SchedulerName

	// Gate the pods under the deployment
	if deployment.Spec.Template.Spec.SchedulingGates == nil {
		deployment.Spec.Template.Spec.SchedulingGates = []corev1.PodSchedulingGate{}
	}
	fluxqGate := corev1.PodSchedulingGate{Name: defaults.SchedulingGateName}
	deployment.Spec.Template.Spec.SchedulingGates = append(deployment.Spec.Template.Spec.SchedulingGates, fluxqGate)

	// We will use this later as a selector to get pods associated with the deployment
	selector := fmt.Sprintf("deployment-%s-%s", deployment.Name, deployment.Namespace)
	deployment.Spec.Template.ObjectMeta.Labels[defaults.SelectorLabel] = selector

	logger.Info("received deployment and gated pods", "Name", deployment.Name)
	return SubmitFluxJob(
		ctx,
		JobWrappedDeployment,
		deployment.Name,
		deployment.Namespace,
		*deployment.Spec.Replicas,
		deployment.Spec.Template.Spec.Containers,
	)
}
