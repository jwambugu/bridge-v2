package repository_test

import (
	"bridge/api/v1/pb"
	"bridge/internal/factory"
	"bridge/internal/repository"
	"bridge/internal/rpc_error"
	"bridge/internal/testutils/docker_test"
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

var testDB *sqlx.DB

func testMain(m *testing.M) (code int, err error) {
	pgSrv, cleanup, err := docker_test.NewPostgresSrv()
	if err != nil {
		return 0, err
	}

	defer func() {
		err = cleanup()
	}()

	testDB = pgSrv.DB
	return m.Run(), err
}

func TestMain(m *testing.M) {
	code, err := testMain(m)
	if err != nil {
		log.Fatalln(err)
	}

	os.Exit(code)
}

func TestUserRepo_Authenticate(t *testing.T) {
	t.Parallel()

	var (
		asserts = assert.New(t)
		ctx     = context.Background()
		u       = factory.NewUser()
	)

	repo, err := repository.NewTestUserRepo(ctx, testDB, u)
	asserts.NoError(err)

	gotUser, err := repo.Authenticate(ctx, u.Email)
	asserts.NoError(err)
	asserts.NotNil(gotUser)
	asserts.Equal(u.ID, gotUser.ID)
}

func TestUserRepo_FindByID(t *testing.T) {
	t.Parallel()

	var (
		asserts = assert.New(t)
		ctx     = context.Background()
		u       = factory.NewUser()
	)

	repo, err := repository.NewTestUserRepo(ctx, testDB, u)
	asserts.NoError(err)

	gotUser, err := repo.FindByID(ctx, u.ID)
	asserts.NoError(err)
	asserts.NotNil(gotUser)
	asserts.Equal(u.ID, gotUser.ID)
}

func TestUserRepo_FindByEmail(t *testing.T) {
	t.Parallel()

	var (
		asserts = assert.New(t)
		ctx     = context.Background()
		u       = factory.NewUser()
	)

	repo, err := repository.NewTestUserRepo(ctx, testDB, u)
	asserts.NoError(err)

	gotUser, err := repo.FindByEmail(ctx, u.Email)
	asserts.NoError(err)
	asserts.NotNil(gotUser)
	asserts.Equal(u.ID, gotUser.ID)
}

func TestUserRepo_FindByPhoneNumber(t *testing.T) {
	t.Parallel()

	var (
		asserts = assert.New(t)
		ctx     = context.Background()

		u = factory.NewUser()
	)

	repo, err := repository.NewTestUserRepo(ctx, testDB, u)
	asserts.NoError(err)

	gotUser, err := repo.FindByPhoneNumber(ctx, u.PhoneNumber)
	asserts.NoError(err)
	asserts.NotNil(gotUser)
	asserts.Equal(u.ID, gotUser.ID)
}

func TestUserRepo_Exists(t *testing.T) {
	var (
		asserts = assert.New(t)
		ctx     = context.Background()
	)

	repo, err := repository.NewTestUserRepo(ctx, testDB)
	asserts.NoError(err)

	tests := []struct {
		name    string
		getUser func() *pb.User
		wantErr error
	}{
		{
			name: "similar user does not exist",
			getUser: func() *pb.User {
				return factory.NewUser()
			},
		},
		{
			name: "email already exists",
			getUser: func() *pb.User {
				var (
					u  = factory.NewUser()
					u1 = factory.NewUser()
				)

				err := repo.Create(ctx, u)
				asserts.NoError(err)

				u.PhoneNumber = u1.PhoneNumber
				return u
			},
			wantErr: rpc_error.ErrEmailExists,
		},
		{
			name: "phone number already exists",
			getUser: func() *pb.User {
				var (
					u  = factory.NewUser()
					u1 = factory.NewUser()
				)

				err := repo.Create(ctx, u)
				asserts.NoError(err)

				u.Email = u1.Email
				return u
			},
			wantErr: rpc_error.ErrPhoneNumberExists,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			user := tt.getUser()

			err = repo.Exists(ctx, user)
			if wantErr := tt.wantErr; wantErr != nil {
				asserts.Error(err)
				asserts.EqualError(err, wantErr.Error())
				return
			}

			asserts.NoError(err)
		})
	}
}

func TestUserRepo_Update(t *testing.T) {
	t.Parallel()

	var (
		asserts = assert.New(t)
		ctx     = context.Background()

		u  = factory.NewUser()
		u1 = factory.NewUser()
	)

	repo, err := repository.NewTestUserRepo(ctx, testDB, u)
	asserts.NoError(err)

	u.Email = u1.Email
	u.AccountStatus = pb.User_INACTIVE

	err = repo.Update(ctx, u)
	asserts.NoError(err)

	gotUser, err := repo.FindByPhoneNumber(ctx, u.PhoneNumber)
	asserts.NoError(err)
	asserts.NotNil(gotUser)
	asserts.Equal(u1.Email, gotUser.Email)
	asserts.Equal(pb.User_INACTIVE, gotUser.AccountStatus)
}
