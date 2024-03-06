package auth_test

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
	"bridge/internal/utils"
	"bridge/services/auth"
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"log"
	"os"
	"testing"
)

func testAuthClient(t *testing.T, addr string) pb.AuthServiceClient {
	t.Helper()

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	return pb.NewAuthServiceClient(conn)
}

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

	vaultProvider, err := vault.NewProvider(vaultClient.Address, vaultClient.Path, vaultClient.Token)
	if err != nil {
		log.Fatalln(err)
	}

	appConfig := config.NewConfig(vaultProvider)
	if err = appConfig.Load(context.Background(), ""); err != nil {
		log.Fatalln(err)
	}

	testSvc.db = pgSrv.DB
	testSvc.vault = vaultClient
	return m.Run()
}

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func TestServer_Login(t *testing.T) {
	t.Parallel()

	var (
		asserts = assert.New(t)
		ctx     = context.Background()
	)

	userRepo, err := repository.NewTestUserRepo(ctx, testSvc.db)
	asserts.NoError(err)

	rs := repository.NewStore()
	rs.UserRepo = userRepo

	jwtManager, err := auth.NewPasetoToken(config.EnvKey.JwtKey)
	asserts.NoError(err)

	var (
		srvAddr    = testutils.TestGRPCSrv(t, jwtManager, logger.TestLogger, rs)
		authClient = testAuthClient(t, srvAddr)
	)

	tests := []struct {
		name    string
		setup   func(t *testing.T) (*pb.User, string)
		wantErr error
	}{
		{
			name: "user can be authenticated successfully",
			setup: func(t *testing.T) (*pb.User, string) {
				t.Helper()
				u := factory.NewUser()

				err = userRepo.Create(ctx, u)
				asserts.NoError(err)

				return u, factory.DefaultPassword
			},
		},
		{
			name: "authentication fails if incorrect password is provided",
			setup: func(t *testing.T) (*pb.User, string) {
				t.Helper()
				u := factory.NewUser()

				err = userRepo.Create(ctx, u)
				asserts.NoError(err)

				return u, "test_password"
			},
			wantErr: rpc_error.ErrUnauthenticated,
		},
		{
			name: "authentication fails if user does not exists",
			setup: func(t *testing.T) (*pb.User, string) {
				t.Helper()
				return factory.NewUser(), "test_password"
			},
			wantErr: rpc_error.ErrUnauthenticated,
		},
		{
			name: "authentication fails if user account is inactive",
			setup: func(t *testing.T) (*pb.User, string) {
				t.Helper()

				u := factory.NewUser()
				u.AccountStatus = pb.User_INACTIVE

				err = userRepo.Create(ctx, u)
				asserts.NoError(err)

				return u, factory.DefaultPassword
			},
			wantErr: rpc_error.ErrInactiveAccount,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testUser, password := tt.setup(t)

			req := &pb.LoginRequest{
				Email:    testUser.Email,
				Password: password,
			}

			res, err := authClient.Login(ctx, req)
			if wantErr := tt.wantErr; wantErr != nil {
				statusFromError, ok := status.FromError(err)
				asserts.True(ok)
				asserts.EqualError(statusFromError.Err(), wantErr.Error())
				asserts.Nil(res)
				return
			}

			asserts.NoError(err)
			asserts.NotNil(res)

			accessTokenPayload, err := jwtManager.Verify(res.AccessToken)
			asserts.NoError(err)
			asserts.Equal(res.User.ID, accessTokenPayload.Subject)
			asserts.Equal(res.User, accessTokenPayload.User)
		})
	}
}

func TestServer_Register(t *testing.T) {
	t.Parallel()

	var (
		asserts = assert.New(t)
		ctx     = context.Background()
	)

	vaultClient, err := vault.NewProvider(testSvc.vault.Address, testSvc.vault.Path, testSvc.vault.Token)
	asserts.NoError(err)
	asserts.NotNil(vaultClient)

	userRepo, err := repository.NewTestUserRepo(ctx, testSvc.db)
	asserts.NoError(err)

	rs := repository.NewStore()
	rs.UserRepo = userRepo

	jwtManager, err := auth.NewPasetoToken(config.EnvKey.JwtKey)
	asserts.NoError(err)

	var (
		srvAddr    = testutils.TestGRPCSrv(t, jwtManager, logger.TestLogger, rs)
		authClient = testAuthClient(t, srvAddr)
	)

	tests := []struct {
		name      string
		createReq func() *pb.RegisterRequest
		wantErr   error
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
		},
		{
			name: "request fails email already exists",
			createReq: func() *pb.RegisterRequest {
				var (
					u  = factory.NewUser()
					u1 = factory.NewUser()
				)

				err = userRepo.Create(ctx, u1)
				asserts.NoError(err)

				return &pb.RegisterRequest{
					Name:            u.Name,
					Email:           u1.Email,
					PhoneNumber:     u.PhoneNumber,
					Password:        factory.DefaultPassword,
					ConfirmPassword: factory.DefaultPassword,
				}
			},
			wantErr: rpc_error.ErrEmailExists,
		},
		{
			name: "request fails phone number already exists",
			createReq: func() *pb.RegisterRequest {
				var (
					u  = factory.NewUser()
					u1 = factory.NewUser()
				)

				err = userRepo.Create(ctx, u1)
				asserts.NoError(err)

				return &pb.RegisterRequest{
					Name:            u.Name,
					Email:           u.Email,
					PhoneNumber:     u1.PhoneNumber,
					Password:        factory.DefaultPassword,
					ConfirmPassword: factory.DefaultPassword,
				}
			},
			wantErr: rpc_error.ErrPhoneNumberExists,
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
			wantErr: rpc_error.ErrPasswordConfirmationMismatch,
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
			wantErr: rpc_error.NewError(codes.InvalidArgument, "invalid RegisterRequest.Name: value length must be at least 3 runes"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := tt.createReq()
			res, err := authClient.Register(ctx, req)

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
			asserts.Equal(req.Name, gotUser.Name)
			asserts.Equal(req.Email, gotUser.Email)

			credentials, err := rs.UserRepo.Authenticate(ctx, req.Email)
			asserts.NoError(err)
			asserts.NotNil(credentials)
			asserts.True(utils.CompareHash(credentials.Password, req.Password))
		})
	}
}
