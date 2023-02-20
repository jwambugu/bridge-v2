package user_test

import (
	"bridge/api/v1/pb"
	"bridge/internal/factory"
	"bridge/internal/logger"
	"bridge/internal/repository"
	"bridge/internal/servers"
	"bridge/services/auth"
	"bridge/services/user"
	"context"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net"
	"testing"
	"time"
)

func startServer(t *testing.T, rs repository.Store, l zerolog.Logger, jwtManager auth.JWTManager) string {
	t.Helper()

	var (
		authSrv = auth.NewAuthService(jwtManager, l, rs)
		userSrv = user.NewUserService(rs)
		srv     = servers.NewGrpcSrv()
	)

	pb.RegisterAuthServiceServer(srv, authSrv)
	pb.RegisterUserServiceServer(srv, userSrv)

	lis, err := net.Listen("tcp", ":0")
	assert.NoError(t, err)

	go func() {
		assert.NoError(t, srv.Serve(lis))
	}()

	return lis.Addr().String()
}

func testUserClient(t *testing.T, addr string) pb.UserServiceClient {
	t.Helper()

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	return pb.NewUserServiceClient(conn)
}

func TestServer_Create(t *testing.T) {
	rs := repository.NewStore()
	rs.UserRepo = user.NewTestRepo()

	var (
		srvAddr    = startServer(t, rs, logger.NewTestLogger, nil)
		userClient = testUserClient(t, srvAddr)
		ctx        = context.Background()
		testUser   = factory.NewUser()

		req = &pb.CreateUserRequest{
			User: &pb.User{
				Name:          testUser.Name,
				Email:         testUser.Email,
				PhoneNumber:   testUser.PhoneNumber,
				Password:      factory.DefaultPassword,
				AccountStatus: pb.User_PENDING_ACTIVE,
				Meta: &pb.UserMeta{
					KycData: &pb.KYCData{
						IdNumber: testUser.Meta.KycData.IdNumber,
					},
				},
				CreatedAt: timestamppb.New(time.Now()),
				UpdatedAt: timestamppb.New(time.Now()),
			},
		}
	)

	res, err := userClient.Create(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, testUser.Name, res.User.GetName())
	assert.NotZero(t, res.User.GetID())
}

func TestServer_Update(t *testing.T) {
	testUser := factory.NewUser()
	rs := repository.NewStore()
	rs.UserRepo = user.NewTestRepo(testUser)

	var (
		srvAddr    = startServer(t, rs, logger.NewTestLogger, nil)
		userClient = testUserClient(t, srvAddr)
		ctx        = context.Background()

		req = &pb.UpdateRequest{
			User: &pb.User{
				ID:            testUser.ID,
				Name:          "Rick Sanchez",
				Email:         testUser.Email,
				PhoneNumber:   testUser.PhoneNumber,
				AccountStatus: pb.User_ACTIVE,
				Meta: &pb.UserMeta{
					KycData: &pb.KYCData{
						IdNumber: "11223344",
					},
				},
				CreatedAt: timestamppb.New(time.Now()),
				UpdatedAt: timestamppb.New(time.Now()),
			},
		}
	)

	res, err := userClient.Update(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, req.User.Name, res.User.GetName())
	assert.Equal(t, pb.User_ACTIVE, res.User.GetAccountStatus())
	assert.Equal(t, req.User.Meta.KycData.IdNumber, res.User.Meta.KycData.IdNumber)
}
