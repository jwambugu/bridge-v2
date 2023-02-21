package repository

import (
	"bridge/api/v1/pb"
	"context"
)

type User interface {
	Authenticate(ctx context.Context, email string) (*pb.User, error)
	Create(ctx context.Context, user *pb.User) error
	Exists(ctx context.Context, user *pb.User) error
	FindByEmail(ctx context.Context, email string) (*pb.User, error)
	FindByID(ctx context.Context, id string) (*pb.User, error)
	FindByPhoneNumber(ctx context.Context, phoneNumber string) (*pb.User, error)
	Update(ctx context.Context, user *pb.User) error
}

type Store struct {
	UserRepo User
}

func NewStore() Store {
	return Store{}
}
