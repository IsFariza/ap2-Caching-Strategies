package interfaces

import "github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/model"

type AppointmentPublisher interface {
	PublishCreated(appt *model.Appointment) error
	PublishStatusUpdated(appt *model.Appointment, oldS, newS model.Status)
}
