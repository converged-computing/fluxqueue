package v1alpha1

import (
	"context"
	"encoding/json"
	"net/http"

	appsv1 "k8s.io/api/apps/v1"
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

	// This main handle function goes through each abstraction type and tries to match the object
	// If we don't have an error, we return an admission success. If we have an error
	// and tryNext is true, we continue to try another.
	marshalledJob, tryNext, err := a.HandleJob(ctx, req)
	if err == nil {
		return admission.PatchResponseFromRaw(req.Object.Raw, marshalledJob)
	}
	// If we get here, err cannot be nil
	if !tryNext {
		return admission.Errored(http.StatusBadRequest, err)
	}

	marshalledDeployment, tryNext, err := a.HandleDeployment(ctx, req)
	if err == nil {
		return admission.PatchResponseFromRaw(req.Object.Raw, marshalledDeployment)
	}
	if !tryNext {
		return admission.Errored(http.StatusBadRequest, err)
	}

	marshalledPod, err := a.HandlePod(ctx, req)
	if err == nil {
		return admission.PatchResponseFromRaw(req.Object.Raw, marshalledPod)
	}
	// If we get here, we don't support the type
	return admission.Errored(http.StatusBadRequest, err)
}

// Handle Pod attempts to serialize a pod
// This function is different between we don't care about a retry - if it's not a pod (the lowest abstraction)
// We can't deal with it.
func (a *jobReceiver) HandlePod(ctx context.Context, req admission.Request) ([]byte, error) {

	marshalledPod := []byte{}

	// Try for a pod next
	pod := &corev1.Pod{}
	err := a.decoder.Decode(req, pod)
	if err != nil {
		logger.Error(err, "admission pod")
		return marshalledPod, err
	}

	// If we get here, we decoded a pod, and want to enqueue as a FluxJob
	err = a.EnqueuePod(ctx, pod)
	if err != nil {
		logger.Error(err, "enqueue pod")
		return marshalledPod, err
	}

	// Mutate the fields in pod
	marshalledPod, err = json.Marshal(pod)
	if err != nil {
		logger.Error(err, "marshalling pod")
	}
	return marshalledPod, err
}

// Handle Job attempts to serialize a job
func (a *jobReceiver) HandleJob(ctx context.Context, req admission.Request) ([]byte, bool, error) {

	marshalledJob := []byte{}
	job := &batchv1.Job{}
	err := a.decoder.Decode(req, job)

	// This isn't a job - keep going!
	if err != nil {
		return marshalledJob, true, err
	}

	// If we get here, we found a job. But if there is an error, we don't try other types
	err = a.EnqueueJob(ctx, job)
	if err != nil {
		logger.Error(err, "enqueue job")
		return marshalledJob, false, err
	}
	marshalledJob, err = json.Marshal(job)
	if err != nil {
		logger.Error(err, "marshalling job")
	}
	return marshalledJob, false, err
}

// HandleDeployment currently isn't supported because we cannot assign specific nodes to replicas in the
// deployment (replicaset)
func (a *jobReceiver) HandleDeployment(ctx context.Context, req admission.Request) ([]byte, bool, error) {

	marshalledDeployment := []byte{}
	deployment := &appsv1.Deployment{}
	err := a.decoder.Decode(req, deployment)
	if err != nil {
		return marshalledDeployment, true, err
	}

	// If we get here, we found a job. But if there is an error, we don't try other types
	err = a.EnqueueDeployment(ctx, deployment)
	if err != nil {
		logger.Error(err, "enqueue deployment")
		return marshalledDeployment, false, err
	}
	marshalledDeployment, err = json.Marshal(deployment)
	if err != nil {
		logger.Error(err, "marshalling job")
	}
	return marshalledDeployment, false, err
}
