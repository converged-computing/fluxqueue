package v1alpha1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	cli  client.Client
	slog = ctrl.Log.WithName("submit")
)

// SubmitFluxJob wraps a pod or job spec into a FluxJob
// We essentially create a CRD for a a FluxJob
func SubmitFluxJob(
	ctx context.Context,
	jobType JobWrapped,
	spec []byte,
	name string,
	namespace string,
	nodes int32,
) error {

	// Ensure we have a client that knows how to create FluxJob
	if cli == nil {
		cli = mgr.GetClient()
	}

	// Check for existing Flux Job
	jobName := GetJobName(jobType, name)
	existing := &FluxJob{}

	// Don't allow a job (of same name, type, namespace) to be submit twice
	err := cli.Get(ctx, types.NamespacedName{Name: jobName, Namespace: namespace}, existing)
	if err == nil {
		slog.Info("Job already exists, will not submit again", "Namespace", namespace, "Name", jobName)
		return err
	}
	// If it's not an issue of not being found, this should not happen
	if !errors.IsNotFound(err) {
		slog.Error(err, "Issue with getting job", "Namespace", namespace, "Name", jobName)
		return err
	}

	// If we get here, create!
	slog.Info("Creating flux job ", "Namespace", namespace, "Name", jobName)

	// Define the Flux Job
	fluxjob := &FluxJob{
		ObjectMeta: metav1.ObjectMeta{Name: jobName, Namespace: namespace},
		Spec: FluxJobSpec{
			JobSpec: "",
			Object:  spec,
			Nodes:   nodes,
			Type:    jobType,
		},
	}
	err = cli.Create(ctx, fluxjob)
	if err != nil {
		slog.Error(err, "Issue with creating job", "Namespace", namespace, "Name", jobName)
		return err
	}
	slog.Info("Created flux job", "Namespace", namespace, "Name", jobName)
	return nil
}
