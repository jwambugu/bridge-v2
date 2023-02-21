package client

import (
	"bridge/services/auth"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
)

// AuthInterceptor methods intercept the RPC call on the client adding authentication headers.
//
// UnaryInterceptor intercepts the execution of a unary RPC.
type AuthInterceptor interface {
	UnaryInterceptor() grpc.UnaryClientInterceptor
}

type authInterceptor struct {
	authClient AuthClient

	accessToken string
}

func (ai *authInterceptor) generateAccessToken(ctx context.Context) {
	res, err := ai.authClient.Login(ctx)
	if err != nil {
		log.Panicf("auth client: login - %v", err.Error())
	}

	ai.accessToken = res.AccessToken
}

func (ai *authInterceptor) UnaryInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		ai.generateAccessToken(ctx)
		ctx = metadata.AppendToOutgoingContext(ctx, auth.HeaderAuthorize, auth.AppendBearerPrefix(ai.accessToken))
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func NewAuthInterceptor(authClient AuthClient) AuthInterceptor {
	return &authInterceptor{authClient: authClient}
}
