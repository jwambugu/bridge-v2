package rpc_error

import "google.golang.org/grpc/codes"

var (
	ErrUnauthenticated              = NewError(codes.Unauthenticated, codes.Unauthenticated.String())
	ErrCreateResourceFailed         = NewError(codes.Internal, "Failed to create specified resource.")
	ErrPasswordConfirmationMismatch = NewError(codes.InvalidArgument, "The password confirmation does not match.")
	ErrServerError                  = NewError(codes.Internal, "Internal server error.")
)
