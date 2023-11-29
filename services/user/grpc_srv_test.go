package user_test

import (
	"bridge/api/v1/pb"
	"bridge/internal/config"
	"bridge/internal/config/vault"
	"bridge/internal/factory"
	"bridge/internal/logger"
	"bridge/internal/repository"
	"bridge/internal/rpc_error"
	"bridge/internal/testutils"
	"bridge/internal/testutils/docker_test"
	"bridge/services/auth"
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"os"
	"testing"
	"time"
)

type testService struct {
	db    *sqlx.DB
	vault *docker_test.VaultClient
}

var testSvc = &testService{}

func testMain(m *testing.M) int {
	pgSrv, postgresCleanup, err := docker_test.NewPostgresSrv()
	if err != nil {
		log.Fatalln(err)
	}

	defer func() {
		if err = postgresCleanup(); err != nil {
			log.Fatalln(err)
		}
	}()

	vaultClient, vaultCleanup, err := docker_test.NewVaultClient()
	if err != nil {
		log.Fatalln(err)
	}

	defer func() {
		if err = vaultCleanup(); err != nil {
			log.Fatalln(err)
		}
	}()

	testSvc.db = pgSrv.DB
	testSvc.vault = vaultClient
	return m.Run()
}

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func TestServer_Create(t *testing.T) {
	var (
		asserts = assert.New(t)
		ctx     = context.Background()
		l       = logger.TestLogger
	)

	userRepo, err := repository.NewTestUserRepo(ctx, testSvc.db)
	asserts.NoError(err)

	vaultClient, err := vault.NewProvider(testSvc.vault.Address, testSvc.vault.Path, testSvc.vault.Token)
	asserts.NoError(err)

	configProvider := config.NewConfig(vaultClient)

	jwtKey, err := configProvider.Get(ctx, "JWT_SYMMETRIC_KEY")
	asserts.NoError(err)

	jwtManager, err := auth.NewPasetoToken(jwtKey)
	asserts.NoError(err)

	rs := repository.NewStore()
	rs.UserRepo = userRepo

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

				err := userRepo.Create(ctx, u1)
				asserts.NoError(err)

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

			admin := factory.NewUser()

			err = userRepo.Create(ctx, admin)
			asserts.NoError(err)

			var (
				srvAddr    = testutils.TestGRPCSrv(t, jwtManager, l, rs)
				cc         = testutils.TestClientConnWithToken(t, srvAddr, admin.Email, factory.DefaultPassword)
				userClient = pb.NewUserServiceClient(cc)

				u   = tt.getUser()
				req = &pb.CreateUserRequest{
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
	var (
		asserts = assert.New(t)
		ctx     = context.Background()
		u       = factory.NewUser()
	)

	userRepo, err := repository.NewTestUserRepo(ctx, testSvc.db, u)
	asserts.NoError(err)

	rs := repository.NewStore()
	rs.UserRepo = userRepo

	vaultClient, err := vault.NewProvider(testSvc.vault.Address, testSvc.vault.Path, testSvc.vault.Token)
	asserts.NoError(err)

	configProvider := config.NewConfig(vaultClient)

	jwtKey, err := configProvider.Get(ctx, "JWT_SYMMETRIC_KEY")
	asserts.NoError(err)

	jwtManager, err := auth.NewPasetoToken(jwtKey)
	asserts.NoError(err)

	var (
		srvAddr    = testutils.TestGRPCSrv(t, jwtManager, logger.TestLogger, rs)
		cc         = testutils.TestClientConnWithToken(t, srvAddr, u.Email, factory.DefaultPassword)
		userClient = pb.NewUserServiceClient(cc)

		req = &pb.UpdateRequest{
			User: &pb.User{
				ID:            u.ID,
				Name:          "Rick Sanchez",
				Email:         u.Email,
				PhoneNumber:   u.PhoneNumber,
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
	asserts.NoError(err)
	asserts.NotNil(res)
	asserts.Equal(req.User.Name, res.User.GetName())
	asserts.Equal(pb.User_ACTIVE, res.User.GetAccountStatus())
	asserts.Equal(req.User.Meta.KycData.IdNumber, res.User.Meta.KycData.IdNumber)
}
