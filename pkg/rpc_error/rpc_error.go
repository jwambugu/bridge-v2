package rpc_error

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewError creates a new instance of Error
func NewError(code codes.Code, msg string) error {
	return status.Error(code, msg)
}
