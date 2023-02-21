package auth

import (
	"bridge/api/v1/pb"
	"bridge/internal/repository"
	"bridge/internal/rpc_error"
	"bridge/internal/utils"
	"context"
	"database/sql"
	"errors"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type service struct {
	OverrideAuthFunc
	pb.UnimplementedAuthServiceServer

	jwtManager JWTManager
	l          zerolog.Logger
	rs         repository.Store
}

func (s *service) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	var (
		sanitizedReq = &pb.RegisterRequest{
			Name:        req.Name,
			Email:       req.Email,
			PhoneNumber: req.PhoneNumber,
		}

		l = s.l.With().Str("action", "register user").Interface("req", sanitizedReq).Logger()
	)

	if req.Password != req.ConfirmPassword {
		l.Err(errors.New("passwords do not match")).Msg("password mismatch")
		return nil, rpc_error.ErrPasswordConfirmationMismatch
	}

	user := &pb.User{
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
	}

	if err := s.rs.UserRepo.Exists(ctx, user); err != nil {
		l.Err(err).Msg("user row exists")
		return nil, err
	}

	passwordHash, err := utils.HashString(req.Password)
	if err != nil {
		l.Err(err).Msg("failed to hash password")
		return nil, rpc_error.ErrServerError
	}

	user.Name = req.Name
	user.Password = passwordHash
	user.AccountStatus = pb.User_PENDING_ACTIVE
	user.CreatedAt = timestamppb.New(time.Now())
	user.UpdatedAt = timestamppb.New(time.Now())

	if err = s.rs.UserRepo.Create(ctx, user); err != nil {
		l.Err(err).Msg("failed to create user")
		return nil, utils.ParseDBError(err)
	}

	l = l.With().Interface("user", user).Logger()

	token, err := s.jwtManager.Generate(user, 60*time.Minute)
	if err != nil {
		l.Err(err).Msg("failed to generate access token")
		return nil, rpc_error.ErrServerError
	}

	l.Info().Msg("user registered successfully")

	return &pb.RegisterResponse{
		User:        user,
		AccessToken: token,
	}, nil
}

func (s *service) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	l := s.l.With().Str("action", "login user").Str("email", req.Email).Logger()

	credentials, err := s.rs.UserRepo.Authenticate(ctx, req.GetEmail())
	if err != nil {
		l.Err(err).Msg("failed to authenticate user")
		if errors.Is(err, sql.ErrNoRows) {
			return nil, rpc_error.ErrUnauthenticated
		}
		return nil, rpc_error.ErrServerError
	}

	if !utils.CompareHash(credentials.Password, req.Password) {
		l.Err(errors.New("passwords don't match")).Msg("passwords hash mismatch")
		return nil, rpc_error.ErrUnauthenticated
	}

	user, err := s.rs.UserRepo.FindByID(ctx, credentials.ID)
	if err != nil {
		l.Err(err).Msg("failed to find user")
		return nil, rpc_error.ErrServerError
	}

	l = l.With().Interface("user", user).Logger()

	if user.AccountStatus == pb.User_INACTIVE {
		return nil, rpc_error.ErrInactiveAccount
	}

	token, err := s.jwtManager.Generate(user, 60*time.Minute)
	if err != nil {
		l.Err(err).Msg("failed to generate access token")
		return nil, rpc_error.ErrServerError
	}

	l.Info().Msg("user authenticated successfully")

	return &pb.LoginResponse{
		User:        user,
		AccessToken: token,
	}, nil
}

func NewService(jwtManager JWTManager, l zerolog.Logger, rs repository.Store) pb.AuthServiceServer {
	return &service{
		jwtManager: jwtManager,
		l:          l.With().Str("service", "auth").Logger(),
		rs:         rs,
	}
}
