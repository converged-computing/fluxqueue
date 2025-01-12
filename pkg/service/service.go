package service

import (
	"os"

	"github.com/converged-computing/fluxqueue/pkg/defaults"
	pb "github.com/converged-computing/fluxqueue/pkg/service-grpc"
	ctrl "sigs.k8s.io/controller-runtime"

	"context"
)

var (
	slog = ctrl.Log.WithName("worker")
)

type ExternalService struct {
	pb.UnimplementedExternalPluginServiceServer
}

// Init is a helper function for any startup stuff, for which now we have none :)
func (f *ExternalService) Init() {
	slog.Info("created external service.")
}

// GetGroup gets and returns the group info
// TODO no good way to look up group - we would need to ask Fluxion directly OR put the grpc
// service alongside the scheduler plugin, which seems like a bad design
func (s *ExternalService) GetGroup(ctx context.Context, in *pb.GroupRequest) (*pb.GroupResponse, error) {
	slog.Info("calling get group endpoint", "Input", in)

	// Prepare an empty match response (that can still be serialized)
	emptyResponse := &pb.GroupResponse{}
	return emptyResponse, nil
}

// List group returns existing groups
func (s *ExternalService) ListGroups(ctx context.Context, in *pb.GroupRequest) (*pb.GroupResponse, error) {

	emptyResponse := &pb.GroupResponse{}

	// Prepare an empty match response (that can still be serialized)
	slog.Info("calling list groups endpoint", "Input", in)
	return emptyResponse, nil
}

// GetResources gets the current Kubernetes Json Graph Format JGF
// This should be created on init of the scheduler
func (s *ExternalService) GetResources(ctx context.Context, in *pb.ResourceRequest) (*pb.ResourceResponse, error) {

	emptyResponse := &pb.ResourceResponse{}

	// Prepare an empty match response (that can still be serialized)
	slog.Info("calling get resources endpoint", "Input", in)

	jgf, err := os.ReadFile(defaults.KubernetesJsonGraphFormat)
	if err != nil {
		slog.Error(err, "reading JGF")
		return emptyResponse, err
	}
	emptyResponse.Graph = string(jgf)
	return emptyResponse, nil
}
