package grpc

import (
	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapErrorToStatus(err error) error {
	if err == nil {
		return nil
	}

	switch err {
	case model.ErrNameRequired, model.ErrEmailRequired, model.ErrInvalidEmail, model.ErrInvalidID:
		return status.Error(codes.InvalidArgument, err.Error())

	case model.ErrDoctorNotFound:
		return status.Error(codes.NotFound, err.Error())

	case model.ErrEmailUsed:
		return status.Error(codes.AlreadyExists, err.Error())

	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
