package controller

import (
	"fmt"

	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"

	api "github.com/converged-computing/fluxqueue/api/v1alpha1"
	"github.com/converged-computing/fluxqueue/pkg/fluxqueue/types"
)

// submitJob submits the job to the queue
func (r *FluxJobReconciler) submitJob(spec *api.FluxJob) (ctrl.Result, error) {

	klog.Infof("Preparing to submit FluxJob %s/%s", spec.Namespace, spec.Name)
	result := ctrl.Result{}

	// Add the pod to the provisional queue
	enqueueStatus, err := r.Queue.Enqueue(spec)
	if err != nil {
		rlog.Error(err, "enqueue for job was not successful", "Namespace", spec.Namespace, "Name", spec.Name)
		return result, err
	}
	rlog.Info("Enqueue for job was successful", "Namespace", spec.Namespace, "Name", spec.Name)

	// If we cannot schedule "unsatisfiable" we delete
	deleteJob := false

	// If the group is already in pending we reject it. We do not
	// currently support expanding groups that are undergoing processing, unless
	// it is an explicit update to an object (TBA).
	if enqueueStatus == types.JobAlreadyInPending {
		rlog.Error(err, "job already in pending queue", "Namespace", spec.Namespace, "Name", spec.Name)
		deleteJob = true

	} else if enqueueStatus == types.JobInvalid {
		rlog.Error(err, "job is invalid or erroneous", "Namespace", spec.Namespace, "Name", spec.Name)
		deleteJob = true

	} else if enqueueStatus == types.JobEnqueueSuccess {
		rlog.Info("job was added to pending", "Namespace", spec.Namespace, "Name", spec.Name)

		// This is usually a database error or similar
	} else if enqueueStatus == types.Unknown {
		rlog.Info("job had unknown issue adding to pending", "Namespace", spec.Namespace, "Name", spec.Name)
	}
	// TODO delete / cleanup from pending
	fmt.Println(deleteJob)
	// TODO(vsoch) need method to keep queue moving here (running Schedule()
	//err = r.Queue.Schedule()
	//if err != nil {
	//	rlog.Error(err, "Issue with fluxnetes Schedule")
	//}
	return result, nil
}
