package testutils

import (
	"bridge/api/v1/pb"
	"bridge/internal/client"
	"bridge/internal/interceptors"
	"bridge/internal/repository"
	"bridge/internal/server"
	"bridge/services/auth"
	"bridge/services/public"
	"bridge/services/user"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
		authSvc   = auth.NewService(jwtManager, l, rs)
		userSvc   = user.NewService(l, rs)
		publicSvc = public.NewService(l, rs)

		unarySrvInterceptors = interceptors.NewUnaryServerInterceptors()
		authProcessor        = auth.NewAuthProcessor(jwtManager, l, rs)
		srv                  = server.NewGrpcSrv(authProcessor, unarySrvInterceptors)
		asserts              = assert.New(t)
	)

	pb.RegisterAuthServiceServer(srv, authSvc)
	pb.RegisterUserServiceServer(srv, userSvc)
	pb.RegisterPublicServiceServer(srv, publicSvc)

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

func TestClientConnWithToken(t *testing.T, addr, email, password string) *grpc.ClientConn {
	t.Helper()
	asserts := assert.New(t)

	transportOpt := grpc.WithTransportCredentials(insecure.NewCredentials())
	cc, err := grpc.Dial(addr, transportOpt)
	asserts.NoError(err)

	authClient := client.NewClient(cc, email, password)
	authInterceptor := client.NewAuthInterceptor(authClient)

	cc, err = grpc.Dial(addr, transportOpt, grpc.WithChainUnaryInterceptor(authInterceptor.UnaryInterceptor()))
	asserts.NoError(err)
	return cc
}
