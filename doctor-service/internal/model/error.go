package model

import "errors"

var (
	ErrNameRequired  = errors.New("doctor name is required")
	ErrEmailRequired = errors.New("doctor email is required")
	ErrInvalidEmail  = errors.New("invalid email format")
	ErrInvalidID     = errors.New("invalid doctor ID format")

	ErrEmailUsed = errors.New("a doctor with this email already exists")

	ErrDoctorNotFound = errors.New("doctor not found")

	ErrInternal = errors.New("an internal server error occurred")
)
