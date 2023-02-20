package auth

import (
	"bridge/api/v1/pb"
	"bridge/internal/repository"
	"bridge/internal/rpc_error"
	"context"
	"database/sql"
	"errors"
	"google.golang.org/grpc/metadata"
	"strings"
)

// AuthenticatorFunc is the pluggable function that performs authentication.
//
// The passed in `Context` will contain the gRPC metadata.MD object (for header-based authentication) and
// the peer.Peer information that can contain transport-based credentials (e.g. `credentials.AuthInfo`).
//
// The returned context will be propagated to handlers, allowing user changes to `Context`. However,
// please make sure that the `Context` returned is a child `Context` of the one passed in.
//
// If error is returned, its `grpc.Code()` will be returned to the user as well as the verbatim message.
// Please make sure you use `codes.Unauthenticated` (lacking auth) and `codes.PermissionDenied`
// (authed, but lacking perms) appropriately.
type AuthenticatorFunc func(ctx context.Context) (context.Context, error)

// ServiceAuthFuncOverride allows a given gRPC service implementation to override the global AuthenticatorFunc.
//
// If a service implements the AuthFuncOverride method, it takes precedence over the `AuthenticatorFunc` method,
// and will be called instead of AuthenticatorFunc for all method invocations within that service.
type ServiceAuthFuncOverride interface {
	AuthenticatorFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error)
}

// Authenticator provides methods for authentication.
//
// Authenticate implements AuthenticatorFunc
type Authenticator interface {
	Authenticate() AuthenticatorFunc
}

var (
	headerAuthorize     = "authorization"
	authorizationScheme = "bearer"
)

type authProcessor struct {
	jwtManager JWTManager
	rs         repository.Store
}

func (ap *authProcessor) Authenticate() AuthenticatorFunc {
	return func(ctx context.Context) (context.Context, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return ctx, rpc_error.ErrMissingCtxAuthMetadata
		}

		header := md.Get(headerAuthorize)
		if len(header) != 1 {
			return ctx, rpc_error.ErrMissingAuthHeader
		}

		splits := strings.SplitN(header[0], " ", 2)
		if len(splits) < 2 {
			return ctx, rpc_error.ErrMissingMalformedToken
		}

		if !strings.EqualFold(splits[0], authorizationScheme) {
			return ctx, rpc_error.ErrInvalidAuthorizationScheme
		}

		claims, err := ap.jwtManager.Verify(splits[1])
		if err != nil {
			switch err {
			case ErrInvalidToken:
				return nil, rpc_error.ErrInvalidToken
			case ErrExpiredToken:
				return nil, rpc_error.ErrExpiredToken
			default:
				return nil, rpc_error.ErrServerError
			}
		}

		u, err := ap.rs.UserRepo.Find(ctx, claims.User.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ctx, rpc_error.ErrUnauthenticated
			}
			return nil, rpc_error.ErrServerError
		}

		if u.AccountStatus == pb.User_INACTIVE {
			return ctx, rpc_error.ErrInactiveAccount
		}

		return ctx, nil
	}
}

// NewAuthProcessor instantiates a new Authenticator
func NewAuthProcessor(jwtManager JWTManager, rs repository.Store) Authenticator {
	return &authProcessor{
		jwtManager: jwtManager,
		rs:         rs,
	}
}

// OverrideAuthFunc overrides global AuthenticatorFunc
type OverrideAuthFunc struct{}

func (OverrideAuthFunc) AuthenticatorFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	return ctx, nil
}
