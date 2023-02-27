package public_test

import (
	"bridge/api/v1/pb"
	"bridge/internal/config"
	"bridge/internal/factory"
	"bridge/internal/logger"
	"bridge/internal/repository"
	"bridge/internal/rpc_error"
	"bridge/internal/testutils"
	"bridge/services/auth"
	"bridge/services/public"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
)

func testPublicClient(t *testing.T, addr string) pb.PublicServiceClient {
	t.Helper()

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	return pb.NewPublicServiceClient(conn)
}

func TestService_GetCategoryBySlug(t *testing.T) {
	var (
		l       = logger.NewTestLogger
		asserts = assert.New(t)
	)

	tests := []struct {
		name        string
		getCategory func() *pb.Category
		wantErr     error
	}{
		{
			name: "gets a category by slug",
			getCategory: func() *pb.Category {
				c := factory.NewCategory()
				repository.NewTestCategoryRepo(l, c)
				return c
			},
		},
		{
			name: "returns not found if category does not exist",
			getCategory: func() *pb.Category {
				return factory.NewCategory()
			},
			wantErr: rpc_error.ErrCategoryNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			jwtKey := config.Get[string](config.JWTKey, "")
			jwtManager, err := auth.NewPasetoToken(jwtKey)
			require.NoError(t, err)

			rs := repository.NewStore()
			rs.CategoryRepo = repository.NewTestCategoryRepo(l)

			var (
				c               = tt.getCategory()
				ctx             = context.Background()
				srvAddr         = testutils.TestGRPCSrv(t, jwtManager, l, rs)
				publicSvcClient = testPublicClient(t, srvAddr)
				req             = &pb.GetCategoryBySlugRequest{Slug: c.Slug}
			)

			res, err := publicSvcClient.GetCategoryBySlug(ctx, req)
			if wantErr := tt.wantErr; wantErr != nil {
				require.Error(t, err)
				require.EqualError(t, err, wantErr.Error())
				return
			}

			require.NoError(t, err)
			asserts.NotNil(res)
			asserts.Equal(c.ID, res.Category.ID)
		})
	}
}

func TestService_GetCategories(t *testing.T) {
	t.Parallel()

	var (
		l       = logger.NewTestLogger
		asserts = assert.New(t)
		c       = factory.NewCategory()
		c1      = factory.NewCategory()
		c2      = factory.NewCategory()
		ctx     = context.Background()
		req     = &pb.GetCategoriesRequest{}
	)

	rs := repository.NewStore()
	rs.CategoryRepo = repository.NewTestCategoryRepo(l, c, c1, c2)

	var ()

	publicSvc := public.NewService(l, rs)

	res, err := publicSvc.GetCategories(ctx, req)

	require.NoError(t, err)
	asserts.NotNil(res)
	asserts.GreaterOrEqual(len(res.Categories), 3)
}
