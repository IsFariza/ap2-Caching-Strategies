package models

type Job struct {
	AppointmentID string `json:"appointment_id"`
	NewStatus     string `json:"new_status"`
}
