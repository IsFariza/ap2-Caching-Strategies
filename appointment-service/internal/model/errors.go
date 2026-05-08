package model

import "errors"

var (
	ErrAppointmentNotFound = errors.New("appointment not found")
	ErrTitleRequired       = errors.New("title is required")
	ErrDoctorIDRequired    = errors.New("doctor_id is required")
	ErrInvalidID           = errors.New("invalid id format")

	ErrDoctorNotFoundRemote     = errors.New("the specified doctor does not exist")
	ErrDoctorServiceUnavailable = errors.New("could not reach doctor service")

	ErrInvalidStatusTransition = errors.New("invalid status transition: cannot go from done back to new")
	ErrInvalidStatus           = errors.New("status must be new, in_progress, or done")
)
