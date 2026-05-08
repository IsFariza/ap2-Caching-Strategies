package event

import (
	"encoding/json"
	"time"

	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/model"
	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/model/interfaces"
	"github.com/nats-io/nats.go"
)

type appointmentPublisher struct {
	nc *nats.Conn
}

func NewAppointmentPublisher(nc *nats.Conn) interfaces.AppointmentPublisher {
	return &appointmentPublisher{nc: nc}
}
func (p *appointmentPublisher) PublishCreated(appt *model.Appointment) error {
	payload := struct {
		EventType  string `json:"event_type"`
		OccurredAt string `json:"occurred_at"`
		ID         string `json:"id"`
		Title      string `json:"title"`
		DoctorID   string `json:"doctor_id"`
		Status     string `json:"status"`
	}{
		EventType:  "appointments.created",
		OccurredAt: time.Now().Format(time.RFC3339),
		ID:         appt.ID,
		Title:      appt.Title,
		DoctorID:   appt.DoctorID,
		Status:     string(appt.Status),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return p.nc.Publish("appointments.created", data)
}

func (p *appointmentPublisher) PublishStatusUpdated(id string, oldS, newS model.Status) {
	payload := struct {
		EventType  string       `json:"event_type"`
		OccurredAt string       `json:"occurred_at"`
		ID         string       `json:"id"`
		OldStatus  model.Status `json:"old_status"`
		NewStatus  model.Status `json:"new_status"`
	}{
		EventType:  "appointments.status_updated",
		OccurredAt: time.Now().Format(time.RFC3339),
		ID:         id,
		OldStatus:  oldS,
		NewStatus:  newS,
	}
	data, _ := json.Marshal(payload)
	p.nc.Publish("appointments.status_updated", data)
}
