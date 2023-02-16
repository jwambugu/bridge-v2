package auth

import (
	"bridge/api/v1/pb"
	"bridge/core/repository"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type server struct {
	pb.UnimplementedAuthServiceServer

	jwtManager JWTManager
	l          zerolog.Logger
	rs         repository.Store
}

func (s *server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	l := s.l.With().Str("action", "register user").Str("req", fmt.Sprintf("%+v", req)).Logger()
	_ = l
	return nil, nil
}

func (s *server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	l := s.l.With().Str("action", "login user").Str("req", fmt.Sprintf("%+v", req)).Logger()

	credentials, err := s.rs.UserRepo.Authenticate(ctx, req.GetEmail())
	if err != nil {
		l.Err(err).Msg("failed to authenticate user")
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.Unauthenticated, "unauthenticated")
		}
		return nil, status.Errorf(codes.Internal, "error finding user: %v", err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(credentials.Password), []byte(req.GetPassword())); err != nil {
		l.Err(err).Msg("failed to compare passwords")
		return nil, status.Error(codes.Unauthenticated, codes.Unauthenticated.String())
	}

	user, err := s.rs.UserRepo.Find(ctx, credentials.ID)
	if err != nil {
		l.Err(err).Msg("failed to find user")
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.Unauthenticated, "unauthenticated")
		}
		return nil, status.Errorf(codes.Internal, "error finding user: %v", err)
	}

	token, err := s.jwtManager.Generate(user, 60*time.Minute)
	if err != nil {
		l.Err(err).Msg("failed to generate access token")
		return nil, status.Errorf(codes.Internal, "error generating access token: %v", err)
	}

	res := &pb.LoginResponse{
		User:        user,
		AccessToken: token,
	}
	return res, nil
}

func NewServer(jwtManager JWTManager, l zerolog.Logger, rs repository.Store) pb.AuthServiceServer {
	return &server{
		jwtManager: jwtManager,
		l:          l.With().Str("service", "auth").Logger(),
		rs:         rs,
	}
}
