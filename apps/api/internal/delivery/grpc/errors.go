package grpcdelivery

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

func toGRPCError(err error) error {
	switch {
	case pkgerrors.Is(err, pkgerrors.ErrNotFound):
		return status.Error(codes.NotFound, "not found")
	case pkgerrors.Is(err, pkgerrors.ErrValidation):
		return status.Error(codes.InvalidArgument, err.Error())
	case pkgerrors.Is(err, pkgerrors.ErrBadRequest):
		return status.Error(codes.InvalidArgument, err.Error())
	case pkgerrors.Is(err, pkgerrors.ErrConflict):
		return status.Error(codes.AlreadyExists, err.Error())
	case pkgerrors.Is(err, pkgerrors.ErrUnauthorized):
		return status.Error(codes.Unauthenticated, err.Error())
	case pkgerrors.Is(err, pkgerrors.ErrForbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
