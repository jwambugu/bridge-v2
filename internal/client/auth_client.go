package client

import (
	"bridge/api/v1/pb"
	"context"
	"google.golang.org/grpc"
)

type AuthClient interface {
	Login(ctx context.Context) (*pb.LoginResponse, error)
}

type authClient struct {
	svc pb.AuthServiceClient

	email    string
	password string
}

func (cl *authClient) Login(ctx context.Context) (*pb.LoginResponse, error) {
	req := &pb.LoginRequest{Email: cl.email, Password: cl.password}
	return cl.svc.Login(ctx, req)
}

func NewClient(cc *grpc.ClientConn, email string, password string) AuthClient {
	svc := pb.NewAuthServiceClient(cc)

	return &authClient{
		svc:      svc,
		email:    email,
		password: password,
	}
}
