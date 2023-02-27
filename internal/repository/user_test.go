package repository_test

import (
	"bridge/api/v1/pb"
	"bridge/internal/db"
	"bridge/internal/factory"
	"bridge/internal/logger"
	"bridge/internal/repository"
	"bridge/internal/rpc_error"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUserRepo_Create(t *testing.T) {
	t.Parallel()

	var (
		l       = logger.NewTestLogger
		asserts = assert.New(t)
	)

	dbConn, err := db.NewConnection()
	require.NoError(t, err)

	var (
		repo = repository.NewUserRepo(dbConn, l)
		u    = factory.NewUser()
	)

	err = repo.Create(context.Background(), u)
	require.NoError(t, err)
	asserts.NotEmpty(u.ID)
}

func TestUserRepo_Authenticate(t *testing.T) {
	t.Parallel()

	var (
		asserts = assert.New(t)
		ctx     = context.Background()
		l       = logger.NewTestLogger
		u       = factory.NewUser()
		repo    = repository.NewTestUserRepo(l, u)
	)

	gotUser, err := repo.Authenticate(ctx, u.Email)
	require.NoError(t, err)
	asserts.NotNil(gotUser)
	asserts.Equal(u.ID, gotUser.ID)
}

func TestUserRepo_FindByID(t *testing.T) {
	t.Parallel()

	var (
		asserts = assert.New(t)
		ctx     = context.Background()
		l       = logger.NewTestLogger
		u       = factory.NewUser()
		repo    = repository.NewTestUserRepo(l, u)
	)

	gotUser, err := repo.FindByID(ctx, u.ID)
	require.NoError(t, err)
	asserts.NotNil(gotUser)
	asserts.Equal(u.ID, gotUser.ID)
}

func TestUserRepo_FindByEmail(t *testing.T) {
	t.Parallel()

	var (
		asserts = assert.New(t)
		ctx     = context.Background()
		l       = logger.NewTestLogger
		u       = factory.NewUser()
		repo    = repository.NewTestUserRepo(l, u)
	)

	gotUser, err := repo.FindByEmail(ctx, u.Email)
	require.NoError(t, err)
	asserts.NotNil(gotUser)
	asserts.Equal(u.ID, gotUser.ID)
}

func TestUserRepo_FindByPhoneNumber(t *testing.T) {
	t.Parallel()

	var (
		asserts = assert.New(t)
		ctx     = context.Background()
		l       = logger.NewTestLogger
		u       = factory.NewUser()
		repo    = repository.NewTestUserRepo(l, u)
	)

	gotUser, err := repo.FindByPhoneNumber(ctx, u.PhoneNumber)
	require.NoError(t, err)
	asserts.NotNil(gotUser)
	asserts.Equal(u.ID, gotUser.ID)
}

func TestUserRepo_Exists(t *testing.T) {
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
				repository.NewTestUserRepo(l, u)
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
				repository.NewTestUserRepo(l, u)
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

			var (
				ctx  = context.Background()
				u    = tt.getUser()
				repo = repository.NewTestUserRepo(l)
			)

			err := repo.Exists(ctx, u)
			if wantErr := tt.wantErr; wantErr != nil {
				require.Error(t, err)
				asserts.EqualError(err, wantErr.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestUserRepo_Update(t *testing.T) {
	t.Parallel()

	var (
		asserts = assert.New(t)
		ctx     = context.Background()
		l       = logger.NewTestLogger
		u       = factory.NewUser()
		u1      = factory.NewUser()
		repo    = repository.NewTestUserRepo(l, u)
	)

	u.Email = u1.Email
	u.AccountStatus = pb.User_INACTIVE

	err := repo.Update(ctx, u)
	require.NoError(t, err)

	gotUser, err := repo.FindByPhoneNumber(ctx, u.PhoneNumber)
	require.NoError(t, err)
	asserts.NotNil(gotUser)
	asserts.Equal(u1.Email, gotUser.Email)
	asserts.Equal(pb.User_INACTIVE, gotUser.AccountStatus)
}
