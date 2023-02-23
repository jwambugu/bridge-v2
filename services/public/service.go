package public

import (
	"bridge/api/v1/pb"
	"bridge/internal/repository"
	"bridge/services/auth"
	"context"
	"github.com/rs/zerolog"
)

type service struct {
	auth.OverrideAuthFunc
	pb.UnimplementedPublicServiceServer

	l  zerolog.Logger
	rs repository.Store
}

func (s service) GetCategory(ctx context.Context, request *pb.GetCategoryRequest) (*pb.GetCategoryResponse, error) {
	//l := s.l.With().Str("action", "get category").Interface("req", fmt.Sprintf("%+v", request)).Logger()

	//TODO implement me
	panic("implement me")
}

func (s service) GetCategories(ctx context.Context, request *pb.GetCategoriesRequest) (*pb.GetCategoriesResponse, error) {
	//TODO implement me
	panic("implement me")
}

func NewService(l zerolog.Logger, rs repository.Store) pb.PublicServiceServer {
	return &service{
		l:  l.With().Str("service", "public").Logger(),
		rs: rs,
	}
}
