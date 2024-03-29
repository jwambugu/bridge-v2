package interceptors

import (
	"bridge/services/auth"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor provides a hook to intercept the execution of a unary RPC on the server.
//
// UnaryServerValidator returns a new unary server interceptor that validates incoming messages.
// Invalid messages will be rejected with `InvalidArgument` before reaching any userspace handlers.
//
// UnaryServerAuthenticator returns a new unary server interceptor that authenticates incoming messages.
//
// Invalid messages will be rejected with `Unauthenticated` before reaching any userspace handlers.
type UnaryServerInterceptor interface {
	UnaryServerValidator() grpc.UnaryServerInterceptor
	UnaryServerAuthenticator(authFunc auth.AuthenticatorFunc) grpc.UnaryServerInterceptor
}

type unaryInterceptor struct{}

// The validate interface starting with protoc-gen-validate v0.6.0.
// See https://github.com/envoyproxy/protoc-gen-validate/pull/455.
type validator interface {
	Validate(all bool) error
}

// The validate interface prior to protoc-gen-validate v0.6.0.
type validatorLegacy interface {
	Validate() error
}

func validate(req interface{}) error {
	switch v := req.(type) {
	case validatorLegacy:
		if err := v.Validate(); err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}
	case validator:
		if err := v.Validate(false); err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}
	}
	return nil
}

func (u *unaryInterceptor) UnaryServerValidator() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		if err = validate(req); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (u *unaryInterceptor) UnaryServerAuthenticator(authFunc auth.AuthenticatorFunc) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var newCtx context.Context
		var err error

		if overrideSrv, ok := info.Server.(auth.ServiceAuthFuncOverride); ok {
			newCtx, err = overrideSrv.AuthenticatorFuncOverride(ctx, info.FullMethod)
		} else {
			newCtx, err = authFunc(ctx)
		}
		if err != nil {
			return nil, err
		}
		return handler(newCtx, req)
	}
}

// NewUnaryServerInterceptors creates a new instance of UnaryServerInterceptor
func NewUnaryServerInterceptors() UnaryServerInterceptor {
	return &unaryInterceptor{}
}
