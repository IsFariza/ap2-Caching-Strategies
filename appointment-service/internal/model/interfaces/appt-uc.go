package interfaces

import (
	"context"

	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/model"
)

type AppointmentUsecase interface {
	Create(ctx context.Context, appointment *model.Appointment) error
	GetById(ctx context.Context, id string) (*model.Appointment, error)
	GetAll(ctx context.Context) ([]*model.Appointment, error)
	Update(ctx context.Context, id string, newStatus model.Status) error
}
