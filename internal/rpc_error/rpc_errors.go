package rpc_error

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	//ErrCreateResourceFailed         = NewError(codes.Internal, "Failed to create specified resource.")
	ErrResourceNotFound             = NewError(codes.NotFound, "Resource not found.")
	ErrCategoryExists               = NewError(codes.AlreadyExists, "Category already exists.")
	ErrCategoryNotFound             = NewError(codes.NotFound, "Category not found.")
	ErrEmailExists                  = NewError(codes.AlreadyExists, "Email is already in use.")
	ErrExpiredToken                 = NewError(codes.Unauthenticated, "Expired access token provided.")
	ErrInactiveAccount              = NewError(codes.Unauthenticated, "Account has been deactivated.")
	ErrInvalidAuthorizationScheme   = NewError(codes.Unauthenticated, "Invalid authorization scheme provided.")
	ErrInvalidToken                 = NewError(codes.Unauthenticated, "Invalid access token provided.")
	ErrMissingAuthHeader            = NewError(codes.Unauthenticated, "Missing authorization header.")
	ErrMissingCtxAuthMetadata       = NewError(codes.Unauthenticated, "Missing context authentication metadata.")
	ErrMissingMalformedToken        = NewError(codes.Unauthenticated, "Malformed authorization token.")
	ErrPasswordConfirmationMismatch = NewError(codes.InvalidArgument, "The password confirmation does not match.")
	ErrPhoneNumberExists            = NewError(codes.AlreadyExists, "Phone number is already in use.")
	ErrServerError                  = NewError(codes.Internal, "Internal server error.")
	ErrUnauthenticated              = NewError(codes.Unauthenticated, codes.Unauthenticated.String())
)

// NewError creates an error representing code and msg.
func NewError(code codes.Code, msg string) error {
	return status.Error(code, msg)
}
