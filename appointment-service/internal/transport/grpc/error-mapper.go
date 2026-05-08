package grpc

import (
	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapErrorToStatus(err error) error {
	if err == nil {
		return nil
	}

	switch err {
	case model.ErrTitleRequired, model.ErrDoctorIDRequired, model.ErrInvalidID, model.ErrInvalidStatus:
		return status.Error(codes.InvalidArgument, err.Error())

	case model.ErrAppointmentNotFound:
		return status.Error(codes.NotFound, err.Error())

	case model.ErrDoctorServiceUnavailable:
		return status.Error(codes.Unavailable, "doctor service is currently unreachable")

	case model.ErrDoctorNotFoundRemote:
		return status.Error(codes.FailedPrecondition, err.Error())

	case model.ErrInvalidStatusTransition:
		return status.Error(codes.InvalidArgument, err.Error())

	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
