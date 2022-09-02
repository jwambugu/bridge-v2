package auth

import (
	"bridge/api/v1/pb"
	"bridge/core/repository"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

type server struct {
	pb.UnimplementedAuthServiceServer

	rs repository.Store
}

func (s *server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user := &pb.User{
		Email: req.GetEmail(),
	}

	if err := s.rs.UserRepo.Find(ctx, user); err != nil {
		log.Printf("failed to find user: %v", err.Error())
		return nil, status.Errorf(codes.Internal, "error finding user: %v", err)
	}

	res := &pb.LoginResponse{
		User:        user,
		AccessToken: "token",
	}
	return res, nil
}

func NewService(rs repository.Store) *server {
	return &server{rs: rs}
}
