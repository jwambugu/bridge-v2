package user

import (
	"bridge/api/v1/pb"
	"bridge/core/repository"
	"bridge/core/rpc_error"
	"context"
	"log"
)

type server struct {
	pb.UnimplementedUserServiceServer

	rs repository.Store
}

func (s *server) Create(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	user := req.GetUser()

	if err := s.rs.UserRepo.Create(ctx, user); err != nil {
		log.Printf("error creating user: %v", err.Error())
		return nil, rpc_error.ErrCreateResourceFailed
	}

	log.Printf("user created successfully - %+v", user)
	return &pb.CreateUserResponse{User: user}, nil
}

func (s *server) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	user := req.GetUser()

	if err := s.rs.UserRepo.Update(ctx, user); err != nil {
		log.Printf("error updating user: %v", err.Error())
		return nil, rpc_error.ErrServerError
	}
	return &pb.UpdateResponse{User: user}, nil
}

func NewServer(rs repository.Store) pb.UserServiceServer {
	return &server{
		rs: rs,
	}
}
