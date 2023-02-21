package server

import (
	"bridge/internal/interceptors"
	"bridge/services/auth"
	"google.golang.org/grpc"
)

// NewGrpcSrv creates a new grpc server with required server options set up.
func NewGrpcSrv(authFunc auth.Authenticator, unarySrvInterceptors interceptors.UnaryServerInterceptor) *grpc.Server {
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			unarySrvInterceptors.UnaryServerValidator(),
			// TODO: add after implementing an authenticator
			unarySrvInterceptors.UnaryServerAuthenticator(authFunc.Authenticate()),
		),
	}
	return grpc.NewServer(opts...)
}
