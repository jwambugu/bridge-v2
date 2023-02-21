package user_test

import (
	"bridge/api/v1/pb"
	"bridge/internal/config"
	"bridge/internal/factory"
	"bridge/internal/logger"
	"bridge/internal/repository"
	"bridge/internal/rpc_error"
	"bridge/internal/testutils"
	"bridge/services/auth"
	"bridge/services/user"
	"context"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
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
	var (
		l       = logger.NewTestLogger
		asserts = assert.New(t)
	)

	tests := []struct {
		name    string
		getUser func() *pb.User
		wantErr error
	}{
		{
			name: "creates a user successfully",
			getUser: func() *pb.User {
				return factory.NewUser()
			},
		},
		{
			name: "request fails if email exists",
			getUser: func() *pb.User {
				var (
					u  = factory.NewUser()
					u1 = factory.NewUser()
				)

				user.NewTestRepo(u1)
				u.Email = u1.Email
				return u
			},
			wantErr: rpc_error.ErrEmailExists,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rs := repository.NewStore()
			rs.UserRepo = user.NewTestRepo()

			jwtKey := config.Get[string](config.JWTKey, "")
			jwtManager, err := auth.NewPasetoToken(jwtKey)
			asserts.NoError(err)

			var (
				srvAddr    = testutils.TestGRPCSrv(t, jwtManager, l, rs)
				userClient = testUserClient(t, srvAddr)
				ctx        = context.Background()
				u          = tt.getUser()
				req        = &pb.CreateUserRequest{
					Name:        u.Name,
					Email:       u.Email,
					PhoneNumber: u.PhoneNumber,
					Meta: &pb.UserMeta{
						KycData: &pb.KYCData{
							IdNumber: u.Meta.KycData.IdNumber,
						},
					},
				}
			)

			res, err := userClient.Create(ctx, req)

			if wantErr := tt.wantErr; wantErr != nil {
				asserts.Error(err)

				statusFromError, ok := status.FromError(err)
				asserts.True(ok)
				asserts.EqualError(statusFromError.Err(), wantErr.Error())
				asserts.Nil(res)
				return
			}

			asserts.NoError(err)
			asserts.NotNil(res)

			gotUser, err := rs.UserRepo.FindByID(ctx, res.User.ID)
			asserts.NoError(err)
			asserts.NotNil(gotUser)
			asserts.Equal(req.Name, gotUser.Name)
			asserts.Equal(req.Email, gotUser.Email)
		})
	}
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
