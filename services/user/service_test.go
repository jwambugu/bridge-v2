package user_test

import (
	"bridge/api/v1/pb"
	"bridge/internal/config"
	"bridge/internal/factory"
	"bridge/internal/logger"
	"bridge/internal/repository"
	"bridge/internal/testutils"
	"bridge/services/auth"
	"bridge/services/user"
	"context"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

func testUserClient(t *testing.T, addr string) pb.UserServiceClient {
	t.Helper()

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	return pb.NewUserServiceClient(conn)
}

func TestServer_Create(t *testing.T) {
	rs := repository.NewStore()
	rs.UserRepo = user.NewTestRepo()

	jwtKey := config.Get[string](config.JWTKey, "")
	jwtManager, err := auth.NewPasetoToken(jwtKey)
	assert.NoError(t, err)

	var (
		srvAddr    = testutils.TestGRPCSrv(t, jwtManager, logger.NewTestLogger, rs)
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

	jwtKey := config.Get[string](config.JWTKey, "")
	jwtManager, err := auth.NewPasetoToken(jwtKey)
	assert.NoError(t, err)

	var (
		srvAddr    = testutils.TestGRPCSrv(t, jwtManager, logger.NewTestLogger, rs)
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
