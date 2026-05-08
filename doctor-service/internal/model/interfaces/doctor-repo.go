package interfaces

import (
	"context"

	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/model"
)

type DoctorRepository interface {
	Create(ctx context.Context, doctor *model.Doctor) error
	GetById(ctx context.Context, id string) (*model.Doctor, error)
	GetAll(ctx context.Context) ([]*model.Doctor, error)
	GetByEmail(ctx context.Context, email string) (*model.Doctor, error)
}
