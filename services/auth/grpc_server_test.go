package auth_test

import (
	"bridge/api/v1/pb"
	"bridge/core/factory"
	"bridge/core/repository"
	"bridge/services/auth"
	"bridge/services/user"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func startGrpcServer(t *testing.T, rs *repository.Store) string {
	t.Helper()

	var (
		authSrv = auth.NewServer(rs)
		srv     = grpc.NewServer()
	)

	pb.RegisterAuthServiceServer(srv, authSrv)

	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	go func() {
		require.NoError(t, srv.Serve(lis))
	}()

	return lis.Addr().String()
}

func testGrpcAuthClient(t *testing.T, addr string) pb.AuthServiceClient {
	t.Helper()

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	return pb.NewAuthServiceClient(conn)
}

func TestServer_Login(t *testing.T) {
	u := factory.NewUser()
	rs := repository.NewStore()
	rs.UserRepo = user.NewTestRepo(u)

	var (
		srvAddr    = startGrpcServer(t, rs)
		authClient = testGrpcAuthClient(t, srvAddr)
		ctx        = context.Background()

		req = &pb.LoginRequest{
			Email:    u.GetEmail(),
			Password: "password",
		}
	)

	res, err := authClient.Login(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
}
