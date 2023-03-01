package category

import (
	"bridge/api/v1/pb"
	"bridge/internal/repository"
	"bridge/internal/rpc_error"
	"bridge/internal/utils"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type service struct {
	pb.UnimplementedCategoryServiceServer

	l  zerolog.Logger
	rs repository.Store
}

func (s *service) CreateCategory(
	ctx context.Context,
	req *pb.CreateCategoryRequest,
) (*pb.CreateCategoryResponse, error) {
	var (
		l = s.l.With().Str("action", "create category").
			Interface("req", fmt.Sprintf("%+v", req)).
			Logger()

		name     = req.Name
		now      = timestamppb.New(time.Now())
		category = &pb.Category{
			Name:      name,
			Slug:      utils.Slugify(name),
			Status:    pb.Category_ACTIVE,
			CreatedAt: now,
			UpdatedAt: now,
		}
	)

	if err := s.rs.CategoryRepo.Create(ctx, category); err != nil {
		l.Err(err).Msg("failed to create category")
		return nil, utils.ParseDBError(err)
	}

	l.Info().Interface("category", fmt.Sprintf("%+v", category)).Msg("action completed")
	return &pb.CreateCategoryResponse{Category: category}, nil
}

func (s *service) GetCategories(ctx context.Context, req *pb.GetCategoriesRequest) (*pb.GetCategoriesResponse, error) {
	l := s.l.With().Str("action", "get categories").Interface("req", fmt.Sprintf("%+v", req)).Logger()

	categories, err := s.rs.CategoryRepo.All(ctx)
	if err != nil {
		l.Err(err).Msg("failed to get categories")
		return nil, utils.ParseDBError(err)
	}

	l.Info().Int("categories", len(categories)).Msg("action completed")
	return &pb.GetCategoriesResponse{Categories: categories}, nil
}

func (s *service) GetCategory(ctx context.Context, req *pb.GetCategoryByIDRequest) (*pb.GetCategoryResponse, error) {
	l := s.l.With().Str("action", "get category").Interface("req", fmt.Sprintf("%+v", req)).Logger()

	category, err := s.rs.CategoryRepo.FindByID(ctx, req.ID)
	if err != nil {
		l.Err(err).Msg("failed to find category by id")
		return nil, utils.ParseDBError(err)
	}

	l.Info().Interface("category", fmt.Sprintf("%+v", category)).Msg("action completed")
	return &pb.GetCategoryResponse{Category: category}, nil
}

func (s *service) UpdateCategory(
	ctx context.Context,
	req *pb.UpdateCategoryRequest,
) (*pb.UpdateCategoryResponse, error) {
	var (
		l    = s.l.With().Str("action", "update category").Interface("req", fmt.Sprintf("%+v", req)).Logger()
		slug = utils.Slugify(req.Name)
	)

	category, err := s.rs.CategoryRepo.FindByID(ctx, req.ID)
	if err != nil {
		l.Err(err).Msg("failed to find category by id")
		if errors.Is(err, sql.ErrNoRows) {
			return nil, rpc_error.ErrCategoryNotFound
		}
		return nil, utils.ParseDBError(err)
	}

	if category.Name != req.Name {
		category, err := s.rs.CategoryRepo.FindBySlug(ctx, slug)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			l.Err(err).Msg("failed to find category by slug")
			return nil, utils.ParseDBError(err)
		}

		if category != nil {
			l.Err(rpc_error.ErrCategoryExists).Interface("category", category).Msg("")
			return nil, rpc_error.ErrCategoryExists
		}
	}

	category.Name = req.Name
	category.Slug = slug
	category.Status = req.Status
	if req.Meta != nil {
		if category.Meta == nil {
			category.Meta = &pb.CategoryMeta{}
		}

		category.Meta.Icon = req.Meta.Icon
	}

	if err = s.rs.CategoryRepo.Update(ctx, category); err != nil {
		l.Err(err).Msg("failed to update category")
		return nil, utils.ParseDBError(err)
	}

	l.Info().Interface("category", fmt.Sprintf("%+v", category)).Msg("action completed")
	return &pb.UpdateCategoryResponse{Category: category}, nil
}

func NewService(l zerolog.Logger, rs repository.Store) pb.CategoryServiceServer {
	return &service{
		l:  l.With().Str("service", "category").Logger(),
		rs: rs,
	}
}
