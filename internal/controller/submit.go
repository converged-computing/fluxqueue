package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	api "github.com/converged-computing/fluxqueue/api/v1alpha1"
	"github.com/converged-computing/fluxqueue/pkg/fluxqueue/types"
)

// submitJob submits the job to the queue
func (r *FluxJobReconciler) submitJob(spec *api.FluxJob) (ctrl.Result, error) {

	rlog.Info("Preparing to submit FluxJob", "Namespace", spec.Namespace, "Name", spec.Name)
	result := ctrl.Result{}

	// Add the pod to the provisional queue
	enqueueStatus, err := r.Queue.Enqueue(spec)
	if err != nil {
		rlog.Error(err, "enqueue for job was not successful", "Namespace", spec.Namespace, "Name", spec.Name)
		return result, err
	}
	rlog.Info("Enqueue for job was successful", "Namespace", spec.Namespace, "Name", spec.Name)

	// If the group is already in pending we reject it. We do not
	// currently support expanding groups that are undergoing processing, unless
	// it is an explicit update to an object (TBA). Note that after the submit
	// label is applied, hitting this state should never happen.
	if enqueueStatus == types.JobAlreadyInPending {
		rlog.Error(err, "Job already in pending queue", "Namespace", spec.Namespace, "Name", spec.Name)

	} else if enqueueStatus == types.JobInvalid {
		rlog.Error(err, "Job is invalid or erroneous", "Namespace", spec.Namespace, "Name", spec.Name)

	} else if enqueueStatus == types.JobEnqueueSuccess {
		rlog.Info("Job was added to pending", "Namespace", spec.Namespace, "Name", spec.Name)

		// This is usually a database error or similar
	} else if enqueueStatus == types.Unknown {
		rlog.Info("Job had unknown issue adding to pending", "Namespace", spec.Namespace, "Name", spec.Name)
	}

	// Keep queue moving here (running Schedule()
	err = r.Queue.Schedule()
	if err != nil {
		rlog.Error(err, "Issue with fluxqueue Schedule")
	}
	return result, nil
}
