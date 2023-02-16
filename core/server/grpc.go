package server

import (
	grpcvalidator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"google.golang.org/grpc"
)

func NewGrpcSrv() *grpc.Server {
	return grpc.NewServer(
		grpc.UnaryInterceptor(grpcvalidator.UnaryServerInterceptor()),
	)
}
