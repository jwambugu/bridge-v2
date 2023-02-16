package auth_test

import (
	"bridge/api/v1/pb"
	"bridge/core/config"
	"bridge/core/factory"
	"bridge/core/logger"
	"bridge/core/repository"
	"bridge/services/auth"
	"bridge/services/user"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func startServer(t *testing.T, rs repository.Store, jwtManager auth.JWTManager) string {
	t.Helper()

	var (
		authSrv = auth.NewServer(jwtManager, logger.NewTestLogger, rs)
		srv     = grpc.NewServer()
	)

	pb.RegisterAuthServiceServer(srv, authSrv)

	lis, err := net.Listen("tcp", ":0")
	assert.NoError(t, err)

	go func() {
		assert.NoError(t, srv.Serve(lis))
	}()

	return lis.Addr().String()
}

func testAuthClient(t *testing.T, addr string) pb.AuthServiceClient {
	t.Helper()

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	return pb.NewAuthServiceClient(conn)
}

func TestServer_Login(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) (*pb.User, string)
		wantCode codes.Code
	}{
		{
			name: "user can be authenticated successfully",
			setup: func(t *testing.T) (*pb.User, string) {
				t.Helper()

				newUser := factory.NewUser()
				user.NewTestRepo(newUser)
				return newUser, factory.DefaultPassword
			},
			wantCode: codes.OK,
		},
		{
			name: "authentication fails if incorrect password is provided",
			setup: func(t *testing.T) (*pb.User, string) {
				t.Helper()

				newUser := factory.NewUser()
				user.NewTestRepo(newUser)
				return newUser, "test"
			},
			wantCode: codes.Unauthenticated,
		},
		{
			name: "authentication fails if user does not exists",
			setup: func(t *testing.T) (*pb.User, string) {
				t.Helper()

				newUser := factory.NewUser()
				return newUser, "test"
			},
			wantCode: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rs := repository.NewStore()
			rs.UserRepo = user.NewTestRepo()
			testUser, password := tt.setup(t)

			jwtKey := config.Get[string](config.JWTKey, "")
			jwtManager, err := auth.NewPasetoToken(jwtKey)
			assert.NoError(t, err)

			var (
				srvAddr    = startServer(t, rs, jwtManager)
				authClient = testAuthClient(t, srvAddr)
				ctx        = context.Background()

				req = &pb.LoginRequest{
					Email:    testUser.Email,
					Password: password,
				}
			)

			res, err := authClient.Login(ctx, req)
			if tt.wantCode != codes.OK {
				statusFromError, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, codes.Unauthenticated, statusFromError.Code())
				assert.Nil(t, res)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)

			accessTokenPayload, err := jwtManager.Verify(res.AccessToken)
			assert.NoError(t, err)
			assert.Equal(t, res.User.ID, accessTokenPayload.Subject)
			assert.Equal(t, res.User, accessTokenPayload.User)
		})
	}

}
