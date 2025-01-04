package controller

import (
	"context"
	"reflect"

	ctrl "sigs.k8s.io/controller-runtime"

	api "github.com/converged-computing/fluxqueue/api/v1alpha1"
)

// UpdateStatus updates the status if it is different from the current
func (r *FluxJobReconciler) updateStatus(spec *api.FluxJob, updateStatus api.SubmitStatus) (ctrl.Result, error) {

	status := api.FluxJobStatus{SubmitStatus: updateStatus}
	if !reflect.DeepEqual(spec.Status, status) {
		spec.Status = status
		err := r.Status().Update(context.Background(), spec)
		if err != nil {
			rlog.Error(err, "Failed to update PodSet status")
			return ctrl.Result{}, err
		}
	}
	rlog.Info("Updated FluxJob", "Name", spec.Name, "Namespace", spec.Namespace, "Status", spec.Status.SubmitStatus)
	return ctrl.Result{}, nil
}
