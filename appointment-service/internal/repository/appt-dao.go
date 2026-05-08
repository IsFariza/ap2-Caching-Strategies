package repository

import (
	"time"

	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/model"
)

type apptRow struct {
	ID          string       `db:"id"`
	Title       string       `db:"title"`
	Description string       `db:"description"`
	DoctorID    string       `db:"doctor_id"`
	Status      model.Status `db:"status"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
}

func (r *apptRow) ToDomain() *model.Appointment {
	return &model.Appointment{
		ID:          r.ID,
		Title:       r.Title,
		Description: r.Description,
		DoctorID:    r.DoctorID,
		Status:      r.Status,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}
