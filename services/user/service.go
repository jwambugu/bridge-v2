package user

import (
	"bridge/api/v1/pb"
	"bridge/internal/repository"
	"bridge/internal/rpc_error"
	"bridge/internal/utils"
	"bridge/services/auth"
	"context"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type service struct {
	auth.OverrideAuthFunc
	pb.UnimplementedUserServiceServer

	l  zerolog.Logger
	rs repository.Store
}

func (s *service) Create(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	var (
		l = s.l.With().Str("action", "create user").Interface("req", req).Logger()
		u = &pb.User{
			Email:       req.Email,
			PhoneNumber: req.PhoneNumber,
		}
	)

	if err := s.rs.UserRepo.Exists(ctx, u); err != nil {
		l.Err(err).Msg("user row exists")
		return nil, err
	}

	password := utils.String(8)
	passwordHash, err := utils.HashString(password)
	if err != nil {
		l.Err(err).Msg("failed to hash password")
		return nil, rpc_error.ErrServerError
	}

	u.Name = req.Name
	u.Password = passwordHash
	u.Meta = req.Meta
	u.CreatedAt = timestamppb.New(time.Now())
	u.UpdatedAt = timestamppb.New(time.Now())

	if err = s.rs.UserRepo.Create(ctx, u); err != nil {
		l.Err(err).Msg("failed to create user")
		return nil, utils.ParseDBError(err)
	}

	l.Info().Interface("user", u).Msg("user created successfully")
	return &pb.CreateUserResponse{User: u}, nil
}

func (s *service) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	var (
		l = s.l.With().Str("action", "update user").Interface("req", req).Logger()
		u = req.User
	)
	if err := s.rs.UserRepo.Update(ctx, u); err != nil {
		l.Err(err).Msg("failed to update user")
		return nil, utils.ParseDBError(err)
	}

	l.Info().Interface("user", u).Msg("user updated successfully")
	return &pb.UpdateResponse{User: u}, nil
}

func NewService(l zerolog.Logger, rs repository.Store) pb.UserServiceServer {
	return &service{
		l:  l.With().Str("service", "user").Logger(),
		rs: rs,
	}
}
