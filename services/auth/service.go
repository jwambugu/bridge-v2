package auth

import (
	"bridge/api/v1/pb"
	"bridge/pkg/repository"
	"bridge/pkg/rpc_error"
	"bridge/pkg/util"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type service struct {
	pb.UnimplementedAuthServiceServer

	jwtManager JWTManager
	l          zerolog.Logger
	rs         repository.Store
}

func (s *service) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	l := s.l.With().Str("action", "register user").Str("req", fmt.Sprintf("%+v", req)).Logger()

	if req.Password != req.ConfirmPassword {
		l.Err(errors.New("passwords do not match")).Msg("password mismatch")
		return nil, rpc_error.ErrPasswordConfirmationMismatch
	}

	passwordHash, err := util.HashString(req.Password)
	if err != nil {
		l.Err(err).Msg("failed to hash password")
		return nil, rpc_error.ErrServerError
	}

	user := &pb.User{
		Name:          req.Name,
		Email:         req.Email,
		PhoneNumber:   req.PhoneNumber,
		Password:      string(passwordHash),
		AccountStatus: pb.User_PENDING_ACTIVE,
		CreatedAt:     timestamppb.New(time.Now()),
		UpdatedAt:     timestamppb.New(time.Now()),
	}

	if err = s.rs.UserRepo.Create(ctx, user); err != nil {
		l.Err(err).Msg("failed to create user")
		return nil, rpc_error.ErrCreateResourceFailed
	}

	token, err := s.jwtManager.Generate(user, 60*time.Minute)
	if err != nil {
		l.Err(err).Msg("failed to generate access token")
		return nil, rpc_error.ErrServerError
	}

	return &pb.RegisterResponse{
		User:        user,
		AccessToken: token,
	}, nil
}

func (s *service) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	l := s.l.With().Str("action", "login user").Str("req", fmt.Sprintf("%+v", req)).Logger()

	credentials, err := s.rs.UserRepo.Authenticate(ctx, req.GetEmail())
	if err != nil {
		l.Err(err).Msg("failed to authenticate user")
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.Unauthenticated, "unauthenticated")
		}
		return nil, rpc_error.ErrServerError
	}

	if !util.CompareHash(credentials.Password, req.Password) {
		l.Err(errors.New("passwords don't match")).Msg("passwords hash mismatch")
		return nil, rpc_error.ErrUnauthenticated
	}

	user, err := s.rs.UserRepo.Find(ctx, credentials.ID)
	if err != nil {
		l.Err(err).Msg("failed to find user")
		if errors.Is(err, sql.ErrNoRows) {
			return nil, rpc_error.ErrUnauthenticated
		}
		return nil, rpc_error.ErrServerError
	}

	token, err := s.jwtManager.Generate(user, 60*time.Minute)
	if err != nil {
		l.Err(err).Msg("failed to generate access token")
		return nil, rpc_error.ErrServerError
	}

	return &pb.LoginResponse{
		User:        user,
		AccessToken: token,
	}, nil
}

func NewAuthService(jwtManager JWTManager, l zerolog.Logger, rs repository.Store) pb.AuthServiceServer {
	return &service{
		jwtManager: jwtManager,
		l:          l.With().Str("service", "auth").Logger(),
		rs:         rs,
	}
}
