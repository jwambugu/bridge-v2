package repository

import (
	"bridge/api/v1/pb"
	"context"
)

type User interface {
	Authenticate(ctx context.Context, email string) (*pb.User, error)
	Create(ctx context.Context, user *pb.User) error
	Find(ctx context.Context, id uint64) (*pb.User, error)
	Update(ctx context.Context, user *pb.User) error
}

type Store struct {
	UserRepo User
}

func NewStore() Store {
	return Store{}
}
