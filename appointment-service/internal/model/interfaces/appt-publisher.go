package interfaces

import "github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/model"

type AppointmentPublisher interface {
	PublishCreated(appt *model.Appointment) error
	PublishStatusUpdated(id string, oldS, newS model.Status)
}
