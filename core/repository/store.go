package repository

import (
	"bridge/api/v1/pb"
	"context"
)

type User interface {
	Create(ctx context.Context, user *pb.User) error
	Find(ctx context.Context, user *pb.User) error
}

type Store struct {
	UserRepo User
}

func NewStore() *Store {
	return &Store{}
}
