package auth_test

import (
	"bridge/api/v1/pb"
	"bridge/pkg/config"
	"bridge/pkg/factory"
	"bridge/pkg/logger"
	"bridge/pkg/repository"
	"bridge/pkg/servers"
	"bridge/pkg/util"
	"bridge/services/auth"
	"bridge/services/user"
	"context"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"net"
	"testing"
)

func startServer(t *testing.T, rs repository.Store, l zerolog.Logger, jwtManager auth.JWTManager) string {
	t.Helper()

	var (
		authSrv = auth.NewAuthService(jwtManager, l, rs)
		srv     = servers.NewGrpcSrv()
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
				return newUser, "test_password"
			},
			wantCode: codes.Unauthenticated,
		},
		{
			name: "authentication fails if user does not exists",
			setup: func(t *testing.T) (*pb.User, string) {
				t.Helper()

				newUser := factory.NewUser()
				return newUser, "test_password"
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
				srvAddr    = startServer(t, rs, logger.NewTestLogger, jwtManager)
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

func TestServer_Register(t *testing.T) {
	tests := []struct {
		name      string
		createReq func() *pb.RegisterRequest
		wantCode  codes.Code
	}{
		{
			name: "registers a user successfully",
			createReq: func() *pb.RegisterRequest {
				u := factory.NewUser()
				return &pb.RegisterRequest{
					Name:            u.Name,
					Email:           u.Email,
					PhoneNumber:     u.PhoneNumber,
					Password:        factory.DefaultPassword,
					ConfirmPassword: factory.DefaultPassword,
				}
			},
			wantCode: codes.OK,
		},
		{
			name: "request fails if passwords don't match",
			createReq: func() *pb.RegisterRequest {
				u := factory.NewUser()
				return &pb.RegisterRequest{
					Name:            u.Name,
					Email:           u.Email,
					PhoneNumber:     u.PhoneNumber,
					Password:        factory.DefaultPassword,
					ConfirmPassword: "DefaultPassword",
				}
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "request fails if validation rules are not met",
			createReq: func() *pb.RegisterRequest {
				u := factory.NewUser()
				return &pb.RegisterRequest{
					PhoneNumber: u.PhoneNumber,
					Password:    factory.DefaultPassword,
				}
			},
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			rs := repository.NewStore()
			rs.UserRepo = user.NewTestRepo()

			jwtKey := config.Get[string](config.JWTKey, "")
			jwtManager, err := auth.NewPasetoToken(jwtKey)
			assert.NoError(t, err)

			var (
				srvAddr    = startServer(t, rs, logger.NewTestLogger, jwtManager)
				authClient = testAuthClient(t, srvAddr)
				ctx        = context.Background()
				req        = tt.createReq()
			)

			res, err := authClient.Register(ctx, req)

			if tt.wantCode != codes.OK {
				assert.Error(t, err)

				statusFromError, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantCode.String(), statusFromError.Code().String())
				assert.Nil(t, res)
				return
			}

			assert.NoError(t, err)

			assert.NotNil(t, res)

			gotUser, err := rs.UserRepo.Find(ctx, res.User.ID)
			assert.NoError(t, err)
			assert.Equal(t, req.Name, gotUser.Name)
			assert.Equal(t, req.Email, gotUser.Email)

			credentials, err := rs.UserRepo.Authenticate(ctx, req.Email)
			assert.NoError(t, err)
			assert.NotNil(t, credentials)
			assert.True(t, util.CompareHash(credentials.Password, req.Password))
		})
	}
}
