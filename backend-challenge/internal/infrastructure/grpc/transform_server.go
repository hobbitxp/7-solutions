package grpc

import (
	// "context"

	// "backend-challenge/internal/domain/model"
	"backend-challenge/internal/domain/service"
	"google.golang.org/grpc"
	// "google.golang.org/grpc/codes"
	// "google.golang.org/grpc/status"
)

// TransformServer implements the TransformService gRPC service
type TransformServer struct {
	transformService service.TransformService
}

// NewTransformServer creates a new TransformServer
func NewTransformServer(transformService service.TransformService) *TransformServer {
	return &TransformServer{
		transformService: transformService,
	}
}

// RegisterTransform registers the transform gRPC server
func RegisterTransform(s *grpc.Server, transformService service.TransformService) {
	// Temporarily do nothing
	_ = s
	_ = transformService
}