package servers

import (
	"bridge/pkg/servers/interceptors"
	"google.golang.org/grpc"
)

// NewGrpcSrv creates a new grpc server with required server options set up.
func NewGrpcSrv() *grpc.Server {
	return grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.UnaryServerValidator()),
	)
}
