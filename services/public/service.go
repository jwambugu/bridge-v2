package public

import (
	"bridge/api/v1/pb"
	"bridge/internal/repository"
	"bridge/internal/rpc_error"
	"bridge/services/auth"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
)

type service struct {
	auth.OverrideAuthFunc
	pb.UnimplementedPublicServiceServer

	l  zerolog.Logger
	rs repository.Store
}

func (s *service) GetCategoryBySlug(
	ctx context.Context,
	req *pb.GetCategoryBySlugRequest,
) (*pb.GetCategoryResponse, error) {
	l := s.l.With().Str("action", "get category by slug").Interface("req", fmt.Sprintf("%+v", req)).Logger()

	category, err := s.rs.CategoryRepo.FindBySlug(ctx, req.Slug)
	if err != nil {
		l.Err(err).Msg("failed to find category")
		if errors.Is(err, sql.ErrNoRows) {
			return nil, rpc_error.ErrCategoryNotFound
		}
		return nil, rpc_error.ErrServerError
	}

	l.Info().Interface("category", category)

	return &pb.GetCategoryResponse{
		Category: category,
	}, nil
}

func (s *service) GetCategories(ctx context.Context, req *pb.GetCategoriesRequest) (*pb.GetCategoriesResponse, error) {
	l := s.l.With().Str("action", "get categories").Interface("req", fmt.Sprintf("%+v", req)).Logger()

	categories, err := s.rs.CategoryRepo.All(ctx)
	if err != nil {
		l.Err(err).Msg("failed to get categories")
		return nil, rpc_error.ErrServerError
	}

	l.Info().Int("categories_len", len(categories))

	return &pb.GetCategoriesResponse{
		Categories: categories,
	}, nil
}

func NewService(l zerolog.Logger, rs repository.Store) pb.PublicServiceServer {
	return &service{
		l:  l.With().Str("service", "public").Logger(),
		rs: rs,
	}
}
