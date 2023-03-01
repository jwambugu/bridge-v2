package category_test

import (
	"bridge/api/v1/pb"
	"bridge/internal/config"
	"bridge/internal/factory"
	"bridge/internal/logger"
	"bridge/internal/repository"
	"bridge/internal/testutils"
	"bridge/services/auth"
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func TestService_CreateCategory(t *testing.T) {
	var (
		l = logger.NewTestLogger
	)

	tests := []struct {
		name        string
		getCategory func() *pb.Category
		wantErrCode codes.Code
	}{
		{
			name: "creates a category successfully",
			getCategory: func() *pb.Category {
				return factory.NewCategory()
			},
		},
		{
			name: "fails to create a category if it exists",
			getCategory: func() *pb.Category {
				c := factory.NewCategory()
				repository.NewTestCategoryRepo(l, c)
				return c
			},
			wantErrCode: codes.AlreadyExists,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			admin := factory.NewUser()

			rs := repository.NewStore()
			rs.UserRepo = repository.NewTestUserRepo(l, admin)
			rs.CategoryRepo = repository.NewTestCategoryRepo(l)

			jwtKey := config.Get[string](config.JWTKey, "")
			jwtManager, err := auth.NewPasetoToken(jwtKey)
			require.NoError(t, err)

			var (
				srvAddr        = testutils.TestGRPCSrv(t, jwtManager, l, rs)
				cc             = testutils.TestClientConnWithToken(t, srvAddr, admin.Email, factory.DefaultPassword)
				categoryClient = pb.NewCategoryServiceClient(cc)
				ctx            = context.Background()
				c              = tt.getCategory()
				req            = &pb.CreateCategoryRequest{
					Name: c.Name,
				}
			)

			res, err := categoryClient.CreateCategory(ctx, req)
			if wantErrCode := tt.wantErrCode; wantErrCode != codes.OK {
				require.Error(t, err)
				statusFromError, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, wantErrCode, statusFromError.Code())
				require.Nil(t, res)
				return
			}

			require.NoError(t, err)
			require.Equal(t, c.Slug, res.Category.Slug)

		})
	}
}

func TestService_GetCategory(t *testing.T) {
	var (
		l = logger.NewTestLogger
	)

	tests := []struct {
		name        string
		getCategory func() *pb.Category
		wantErrCode codes.Code
	}{
		{
			name: "gets a category successfully",
			getCategory: func() *pb.Category {
				c := factory.NewCategory()
				repository.NewTestCategoryRepo(l, c)
				return c
			},
		},
		{
			name: "fails to find a category if does not exist",
			getCategory: func() *pb.Category {
				c := factory.NewCategory()
				c.ID = "d7d8d9ec-c3ee-4a88-9ceb-5d11b0b63c96"
				return c
			},
			wantErrCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			admin := factory.NewUser()

			rs := repository.NewStore()
			rs.UserRepo = repository.NewTestUserRepo(l, admin)
			rs.CategoryRepo = repository.NewTestCategoryRepo(l)

			jwtKey := config.Get[string](config.JWTKey, "")
			jwtManager, err := auth.NewPasetoToken(jwtKey)
			require.NoError(t, err)

			var (
				srvAddr        = testutils.TestGRPCSrv(t, jwtManager, l, rs)
				cc             = testutils.TestClientConnWithToken(t, srvAddr, admin.Email, factory.DefaultPassword)
				categoryClient = pb.NewCategoryServiceClient(cc)
				ctx            = context.Background()
				c              = tt.getCategory()
				req            = &pb.GetCategoryByIDRequest{
					ID: c.ID,
				}
			)

			res, err := categoryClient.GetCategory(ctx, req)
			if wantErrCode := tt.wantErrCode; wantErrCode != codes.OK {
				require.Error(t, err)
				statusFromError, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, wantErrCode.String(), statusFromError.Code().String())
				require.Nil(t, res)
				return
			}

			require.NoError(t, err)
			require.Equal(t, c.ID, res.Category.ID)
			require.Equal(t, c.Slug, res.Category.Slug)

		})
	}
}

func TestService_GetCategories(t *testing.T) {
	t.Parallel()
	var (
		l     = logger.NewTestLogger
		admin = factory.NewUser()
		c     = factory.NewCategory()
		c1    = factory.NewCategory()
		c2    = factory.NewCategory()
	)

	rs := repository.NewStore()
	rs.UserRepo = repository.NewTestUserRepo(l, admin)
	rs.CategoryRepo = repository.NewTestCategoryRepo(l, c, c1, c2)

	jwtKey := config.Get[string](config.JWTKey, "")
	jwtManager, err := auth.NewPasetoToken(jwtKey)
	require.NoError(t, err)

	var (
		srvAddr        = testutils.TestGRPCSrv(t, jwtManager, l, rs)
		cc             = testutils.TestClientConnWithToken(t, srvAddr, admin.Email, factory.DefaultPassword)
		categoryClient = pb.NewCategoryServiceClient(cc)
		ctx            = context.Background()
		req            = &pb.GetCategoriesRequest{}
	)

	res, err := categoryClient.GetCategories(ctx, req)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(res.Categories), 3)
}

func TestService_UpdateCategory(t *testing.T) {
	var (
		l = logger.NewTestLogger
	)

	tests := []struct {
		name        string
		getCategory func() *pb.Category
		wantErrCode codes.Code
	}{
		{
			name: "updates a category successfully",
			getCategory: func() *pb.Category {
				var (
					c  = factory.NewCategory()
					c1 = factory.NewCategory()
				)

				repository.NewTestCategoryRepo(l, c)

				c.Name = c1.Name
				c.Slug = c1.Slug
				c.Status = pb.Category_INACTIVE
				c.Meta.Icon = c1.Meta.Icon
				return c
			},
		},
		{
			name: "fails to find a category if does not exist",
			getCategory: func() *pb.Category {
				c := factory.NewCategory()
				c.ID = "d7d8d9ec-c3ee-4a88-9ceb-5d11b0b63c96"
				return c
			},
			wantErrCode: codes.NotFound,
		},
		{
			name: "fails to update a category if the slug exists",
			getCategory: func() *pb.Category {
				var (
					c  = factory.NewCategory()
					c1 = factory.NewCategory()
				)

				repository.NewTestCategoryRepo(l, c, c1)

				c.Name = c1.Name
				return c
			},
			wantErrCode: codes.AlreadyExists,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			admin := factory.NewUser()

			rs := repository.NewStore()
			rs.UserRepo = repository.NewTestUserRepo(l, admin)
			rs.CategoryRepo = repository.NewTestCategoryRepo(l)

			jwtKey := config.Get[string](config.JWTKey, "")
			jwtManager, err := auth.NewPasetoToken(jwtKey)
			require.NoError(t, err)

			var (
				srvAddr        = testutils.TestGRPCSrv(t, jwtManager, l, rs)
				cc             = testutils.TestClientConnWithToken(t, srvAddr, admin.Email, factory.DefaultPassword)
				categoryClient = pb.NewCategoryServiceClient(cc)
				ctx            = context.Background()
				c              = tt.getCategory()
				req            = &pb.UpdateCategoryRequest{
					ID:     c.ID,
					Name:   c.Name,
					Status: c.Status,
					Meta:   c.Meta,
				}
			)

			res, err := categoryClient.UpdateCategory(ctx, req)
			if wantErrCode := tt.wantErrCode; wantErrCode != codes.OK {
				require.Error(t, err)
				statusFromError, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, wantErrCode.String(), statusFromError.Code().String())
				require.Nil(t, res)
				return
			}

			require.NoError(t, err)
			require.Equal(t, c.Slug, res.Category.Slug)
			require.Equal(t, c.Status, res.Category.Status)
			require.Equal(t, c.Meta.Icon, res.Category.Meta.Icon)
		})
	}
}
