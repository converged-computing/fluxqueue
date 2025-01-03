package v1alpha1

import (
	"context"
	"encoding/json"
	"net/http"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	logger             = ctrl.Log.WithName("webhook")
	schedulingGateName = "fluxqueue"
	schedulerName      = "fluxion"
	mgr                manager.Manager
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
		logger.Info("Admission pod success.")
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
	logger.Info("Admission job success.")
	return admission.PatchResponseFromRaw(req.Object.Raw, marshalledJob)
}

// InjectPod adds the sidecar container to a pod
func (a *jobReceiver) EnqueuePod(ctx context.Context, pod *corev1.Pod) error {
	logger.Info("Enqueue pod", "Name", pod.Name, "Namespace", pod.Namespace)

	// Add scheduling gate to the pod
	if pod.Spec.SchedulingGates == nil {
		pod.Spec.SchedulingGates = []corev1.PodSchedulingGate{}
	}
	fluxqGate := corev1.PodSchedulingGate{Name: schedulingGateName}
	pod.Spec.SchedulingGates = append(pod.Spec.SchedulingGates, fluxqGate)

	// Ensure the pod gets scheduled with fluxion scheduler
	pod.Spec.SchedulerName = schedulerName
	logger.Info("received pod", "Name", pod.Name)

	podBytes, err := json.Marshal(pod)
	if err != nil {
		return err
	}

	return SubmitFluxJob(
		ctx,
		JobWrappedPod,
		podBytes,
		pod.Name,
		pod.Namespace,
		1,
	)
}

// EnqueueJob suspends the job and creates a FluxJob
func (a *jobReceiver) EnqueueJob(ctx context.Context, job *batchv1.Job) error {

	logger.Info("Enqueue job", "Name", job.Name, "Namespace", job.Namespace)

	// Suspend the job
	suspended := true
	job.Spec.Suspend = &suspended

	// Convert to bytes to store with the queue
	jobBytes, err := json.Marshal(job)
	if err != nil {
		return err
	}

	logger.Info("received job", "Name", job.Name)
	return SubmitFluxJob(
		ctx,
		JobWrappedPod,
		jobBytes,
		job.Name,
		job.Namespace,
		*job.Spec.Parallelism,
	)
}
