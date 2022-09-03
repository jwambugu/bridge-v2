package auth

import (
	"bridge/api/v1/pb"
	"bridge/core/repository"
	"context"
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

type server struct {
	pb.UnimplementedAuthServiceServer

	rs         *repository.Store
	jwtManager JWTManager
}

func (s *server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	credentials, err := s.rs.UserRepo.Authenticate(ctx, req.GetEmail())
	if err != nil {
		log.Printf("failed to authenticate user: %v", err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.Unauthenticated, "unauthenticated")
		}
		return nil, status.Errorf(codes.Internal, "error finding user: %v", err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(credentials.Password), []byte(req.GetPassword())); err != nil {
		return nil, status.Error(codes.Unauthenticated, codes.Unauthenticated.String())
	}

	user, err := s.rs.UserRepo.Find(ctx, credentials.ID)
	if err != nil {
		log.Printf("failed to find user: %v", err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.Unauthenticated, "unauthenticated")
		}
		return nil, status.Errorf(codes.Internal, "error finding user: %v", err)
	}

	token, err := s.jwtManager.Generate(user, 60*time.Minute)
	if err != nil {
		log.Printf("failed to generate access token: %v", err.Error())
		return nil, status.Errorf(codes.Internal, "error generating access token: %v", err)
	}

	res := &pb.LoginResponse{
		User:        user,
		AccessToken: token,
	}
	return res, nil
}

func NewServer(rs *repository.Store, jwtManager JWTManager) *server {
	return &server{
		rs:         rs,
		jwtManager: jwtManager,
	}
}
