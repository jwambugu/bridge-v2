package repository_test

import (
	"bridge/api/v1/pb"
	"bridge/internal/db"
	"bridge/internal/factory"
	"bridge/internal/logger"
	"bridge/internal/repository"
	"bridge/internal/utils"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func createCategory(ctx context.Context, t *testing.T, c *pb.Category) (repository.Category, *pb.Category) {
	t.Helper()

	var (
		l       = logger.NewTestLogger
		asserts = assert.New(t)
	)

	dbConn, err := db.NewConnection()
	require.NoError(t, err)

	repo := repository.NewCategoryRepo(dbConn, l)

	err = repo.Create(ctx, c)
	require.NoError(t, err)
	asserts.NotEmpty(c.ID)
	return repo, c
}

func TestRepo_Create(t *testing.T) {
	t.Parallel()
	createCategory(context.Background(), t, factory.NewCategory())
}

func TestRepo_FindByID(t *testing.T) {
	t.Parallel()

	var (
		asserts = assert.New(t)
		ctx     = context.Background()
		repo, c = createCategory(ctx, t, factory.NewCategory())
	)

	gotCategory, err := repo.FindByID(ctx, c.ID)
	require.NoError(t, err)
	asserts.NotNil(gotCategory)
	asserts.Equal(c.ID, gotCategory.ID)
	asserts.Equal(c.Slug, gotCategory.Slug)
}

func TestRepo_FindBySlug(t *testing.T) {
	t.Parallel()

	var (
		asserts = assert.New(t)
		ctx     = context.Background()
		repo, c = createCategory(ctx, t, factory.NewCategory())
	)

	gotCategory, err := repo.FindBySlug(ctx, c.Slug)
	require.NoError(t, err)
	asserts.NotNil(gotCategory)
	asserts.Equal(c.ID, gotCategory.ID)
	asserts.Equal(c.Slug, gotCategory.Slug)
}

func TestRepo_FindByUpdate(t *testing.T) {
	t.Parallel()

	var (
		asserts = assert.New(t)
		ctx     = context.Background()
		repo, c = createCategory(ctx, t, factory.NewCategory())
		c1      = factory.NewCategory()
	)

	c.Name = c1.Name + ` Update`
	c.Status = pb.Category_ACTIVE

	err := repo.Update(ctx, c)
	require.NoError(t, err)

	gotCategory, err := repo.FindBySlug(ctx, c.Slug)
	require.NoError(t, err)
	asserts.NotNil(gotCategory)

	asserts.Equal(c.Name, gotCategory.Name)
	asserts.Equal(utils.Slugify(c.Name), gotCategory.Slug)
	asserts.Equal(c.Status, gotCategory.Status)
}
