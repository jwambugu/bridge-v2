package testutils

import (
	"bridge/api/v1/pb"
	"bridge/internal/interceptors"
	"bridge/internal/repository"
	"bridge/internal/servers"
	"bridge/services/auth"
	"bridge/services/user"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

var lis net.Listener

// TestGRPCSrv creates and runs a grpc.Server. Returns the address of the server.
func TestGRPCSrv(
	t *testing.T,
	jwtManager auth.JWTManager,
	l zerolog.Logger,
	rs repository.Store,
) string {
	var (
		authSvc              = auth.NewService(jwtManager, l, rs)
		userSvc              = user.NewService(rs)
		unarySrvInterceptors = interceptors.NewUnaryServerInterceptors()
		authProcessor        = auth.NewAuthProcessor(jwtManager, rs)
		srv                  = servers.NewGrpcSrv(authProcessor, unarySrvInterceptors)
		asserts              = assert.New(t)
	)

	pb.RegisterAuthServiceServer(srv, authSvc)
	pb.RegisterUserServiceServer(srv, userSvc)

	var err error
	if lis == nil {
		lis, err = net.Listen("tcp", ":0")
		asserts.NoError(err)
	}

	go func() {
		asserts.NoError(srv.Serve(lis))
	}()

	return lis.Addr().String()
}
