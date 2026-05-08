package interfaces

import (
	"context"

	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/model"
)

type DoctorUseCase interface {
	Create(ctx context.Context, doc *model.Doctor) (*model.Doctor, error)
	GetByID(ctx context.Context, id string) (*model.Doctor, error)
	List(ctx context.Context) ([]*model.Doctor, error)
}
