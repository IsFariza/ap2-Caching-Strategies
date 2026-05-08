package event

import (
	"encoding/json"
	"time"

	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/model"
	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/model/interfaces"
	"github.com/nats-io/nats.go"
)

type doctorPublisher struct {
	nc *nats.Conn
}

func NewDoctorPublisher(nc *nats.Conn) interfaces.DoctorPublisher {
	return &doctorPublisher{nc: nc}
}

func (p *doctorPublisher) PublishDoctorCreated(doc *model.Doctor) error {

	subject := "doctors.created"

	eventPayload := struct {
		EventType      string `json:"event_type"`
		OccurredAt     string `json:"occurred_at"`
		ID             string `json:"id"`
		FullName       string `json:"full_name"`
		Specialization string `json:"specialization"`
		Email          string `json:"email"`
	}{
		EventType:      "doctors.created",
		OccurredAt:     time.Now().Format(time.RFC3339),
		ID:             doc.ID,
		FullName:       doc.FullName,
		Specialization: doc.Specialization,
		Email:          doc.Email,
	}

	data, err := json.Marshal(eventPayload)
	if err != nil {
		return err
	}

	return p.nc.Publish(subject, data)
}
