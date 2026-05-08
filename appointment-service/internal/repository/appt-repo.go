package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/model"
	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/model/interfaces"
)

type appointmentRepo struct {
	db *sql.DB
}

func NewAppointmentRepository(db *sql.DB) interfaces.AppointmentRepo {
	return &appointmentRepo{
		db: db,
	}
}

func (r *appointmentRepo) Create(ctx context.Context, appt *model.Appointment) error {
	query := `INSERT INTO appointments (id, title, description, doctor_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	if appt.CreatedAt.IsZero() {
		appt.CreatedAt = time.Now()
	}
	appt.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		appt.ID,
		appt.Title,
		appt.Description,
		appt.DoctorID,
		appt.Status,
		appt.CreatedAt,
		appt.UpdatedAt,
	)
	return err

}

func (r *appointmentRepo) GetById(ctx context.Context, id string) (*model.Appointment, error) {
	query := `SELECT id, title, description, doctor_id, status, created_at, updated_at FROM appointments WHERE id = $1`
	var appt apptRow
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&appt.ID,
		&appt.Title,
		&appt.Description,
		&appt.DoctorID,
		&appt.Status,
		&appt.CreatedAt,
		&appt.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrAppointmentNotFound
		}
		return nil, err
	}
	return appt.ToDomain(), nil
}

func (r *appointmentRepo) GetAll(ctx context.Context) ([]*model.Appointment, error) {
	query := `SELECT id, title, description, doctor_id, status, created_at, updated_at FROM appointments`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*model.Appointment
	for rows.Next() {
		var row apptRow
		if err := rows.Scan(&row.ID, &row.Title, &row.Description, &row.DoctorID, &row.Status, &row.CreatedAt, &row.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, row.ToDomain())
	}

	return result, nil
}

func (r *appointmentRepo) Update(ctx context.Context, id string, newStatus model.Status) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var currentStatus string
	checkQuery := `SELECT status FROM appointments WHERE id = $1 FOR UPDATE`
	err = tx.QueryRowContext(ctx, checkQuery, id).Scan(&currentStatus)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrAppointmentNotFound
		}
		return err
	}
	query := `UPDATE appointments SET status = $1, updated_at = $2 WHERE id = $3`

	_, err = tx.ExecContext(ctx, query, string(newStatus), time.Now(), id)
	if err != nil {
		return err
	}

	return tx.Commit()
}
