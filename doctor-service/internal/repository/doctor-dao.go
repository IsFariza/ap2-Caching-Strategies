package repository

import (
	"time"

	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/model"
)

type doctorRow struct {
	ID             string    `db:"id"`
	FullName       string    `db:"full_name"`
	Specialization string    `db:"specialization"`
	Email          string    `db:"email"`
	CreatedAt      time.Time `db:"created_at"`
}

func (r *doctorRow) ToDomain() *model.Doctor {
	return &model.Doctor{
		ID:             r.ID,
		FullName:       r.FullName,
		Specialization: r.Specialization,
		Email:          r.Email,
		CreatedAt:      r.CreatedAt,
	}
}
