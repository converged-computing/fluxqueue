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

package controller

import (
	"context"

	fluxion "github.com/converged-computing/fluxion/pkg/client"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	api "github.com/converged-computing/fluxqueue/api/v1alpha1"
	"github.com/converged-computing/fluxqueue/pkg/fluxqueue"
)

var (
	rlog = ctrl.Log.WithName("fluxqueue")
)

// FluxJobReconciler reconciles a FluxJob object
type FluxJobReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	RESTClient rest.Interface
	RESTConfig *rest.Config
	Queue      *fluxqueue.Queue
	Fluxion    fluxion.Client

	// Defaults for fluxion / scheduling
	Duration int
}

// NewFluxJobReconciler creates a new reconciler with clients, a Queue, and Fluxion client
func NewFluxJobReconciler(
	client client.Client,
	scheme *runtime.Scheme,
	restConfig *rest.Config,
	restClient rest.Interface,
	queue *fluxqueue.Queue,
	fluxCli fluxion.Client,
) *FluxJobReconciler {
	return &FluxJobReconciler{
		Client:     client,
		Scheme:     scheme,
		RESTClient: restClient,
		RESTConfig: restConfig,
		Queue:      queue,
		Fluxion:    fluxCli,
	}
}

// +kubebuilder:rbac:groups="",resources=nodes;events,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=pods/log,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods/exec,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups=batch,resources=jobs/log,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs/exec,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups=jobs.converged-computing.org,resources=fluxjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=jobs.converged-computing.org,resources=fluxjobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=jobs.converged-computing.org,resources=fluxjobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the FluxJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *FluxJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// Create a new MetricSet
	var spec api.FluxJob

	// Keep developer informed what is going on.
	rlog.Info("🌀 Event received by FluxJob controller!")
	rlog.Info("Request: ", "req", req)

	// Does the metric exist yet (based on name and namespace)
	err := r.Get(ctx, req.NamespacedName, &spec)
	if err != nil {

		// Create it, doesn't exist yet
		if errors.IsNotFound(err) {
			rlog.Info("🔴 FluxJob not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		rlog.Info("🔴 Failed to get FluxJob. Re-running reconcile.")
		return ctrl.Result{Requeue: true}, err
	}
	rlog.Info("Found FluxJob", "Name", spec.Name, "Namespace", spec.Namespace, "Status", spec.Status.SubmitStatus)
	result := ctrl.Result{}

	// If the job is already submit, continue
	if spec.Status.SubmitStatus == api.SubmitStatusSubmit {
		return result, nil
	}

	// Submit the job to the queue - TODO if error, should delete?
	// If we are successful, update the status
	result, err = r.submitJob(&spec)
	if err == nil {
		err = r.updateStatus(&spec, api.SubmitStatusSubmit)
	} else {
		err = r.updateStatus(&spec, api.SubmitStatusError)
	}
	return result, err
}

// Set default duration
func (r *FluxJobReconciler) SetDuration(duration int) {
	r.Duration = duration
}

// SetupWithManager sets up the controller with the Manager.
func (r *FluxJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.FluxJob{}).
		Complete(r)
}
