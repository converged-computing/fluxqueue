package v1alpha1

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/converged-computing/fluxqueue/pkg/defaults"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	logger = ctrl.Log.WithName("webhook")
	mgr    manager.Manager
)

// IMPORTANT: if you use the controller-runtime builder, it will derive this name automatically from the gvk (kind, version, etc. so find the actual created path in the logs)
// kubectl describe mutatingwebhookconfigurations.admissionregistration.k8s.io
// It will also only allow you to describe one object type with For()
// This is disabled so we manually manage it - multiple types to a list did not work: config/webhook/manifests.yaml
////kubebuilder:webhook:path=/mutate-v1-sidecar,mutating=true,failurePolicy=fail,sideEffects=None,groups=core;batch,resources=pods;jobs,verbs=create,versions=v1,name=morascache.kb.io,admissionReviewVersions=v1

// NewMutatingWebhook allows us to keep the jobReceiver private
// If it's public it's exported and kubebuilder tries to add to zz_generated_deepcopy
// and you get all kinds of terrible errors about admission.Decoder missing DeepCopyInto
func NewMutatingWebhook(mainManager manager.Manager) *jobReceiver {
	mgr = mainManager
	return &jobReceiver{decoder: admission.NewDecoder(mainManager.GetScheme())}
}

type jobReceiver struct {
	decoder admission.Decoder
}

// Handle is the main function to receive the object (por or job to start)
// Pods: we use scheduling gates to prevent from scheduling
// Jobs: we suspend
// TODO we need to catch remainder of types
func (a *jobReceiver) Handle(ctx context.Context, req admission.Request) admission.Response {

	// First try for job
	job := &batchv1.Job{}
	err := a.decoder.Decode(req, job)
	if err != nil {

		// Try for a pod next
		pod := &corev1.Pod{}
		err := a.decoder.Decode(req, pod)
		if err != nil {
			logger.Error(err, "admission error")
			return admission.Errored(http.StatusBadRequest, err)
		}

		// If we get here, we decoded a pod, and want to enqueue as a FluxJob
		err = a.EnqueuePod(ctx, pod)
		if err != nil {
			logger.Error(err, "inject pod")
			return admission.Errored(http.StatusBadRequest, err)
		}

		// Mutate the fields in pod
		marshalledPod, err := json.Marshal(pod)
		if err != nil {
			logger.Error(err, "marshalling pod")
			return admission.Errored(http.StatusInternalServerError, err)
		}
		return admission.PatchResponseFromRaw(req.Object.Raw, marshalledPod)
	}

	// If we get here, we found a job
	err = a.EnqueueJob(ctx, job)
	if err != nil {
		logger.Error(err, "inject job")
		return admission.Errored(http.StatusBadRequest, err)
	}
	marshalledJob, err := json.Marshal(job)
	if err != nil {
		logger.Error(err, "marshalling job")
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshalledJob)
}

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
