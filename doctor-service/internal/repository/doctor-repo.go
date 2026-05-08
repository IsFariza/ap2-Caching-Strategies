package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/model"
	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/model/interfaces"
)

type doctorRepository struct {
	db *sql.DB
}

func NewDoctorRepository(db *sql.DB) interfaces.DoctorRepository {
	return &doctorRepository{
		db: db,
	}
}

func (r *doctorRepository) Create(ctx context.Context, doctor *model.Doctor) error {

	query := `
		INSERT INTO doctors (id, full_name, specialization, email, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	if doctor.CreatedAt.IsZero() {
		doctor.CreatedAt = time.Now()
	}

	_, err := r.db.ExecContext(ctx, query,
		doctor.ID,
		doctor.FullName,
		doctor.Specialization,
		doctor.Email,
		doctor.CreatedAt,
	)
	return err
}

func (r *doctorRepository) GetById(ctx context.Context, id string) (*model.Doctor, error) {
	query := `SELECT id, full_name, specialization, email, created_at FROM doctors WHERE id = $1`

	var row doctorRow
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&row.ID,
		&row.FullName,
		&row.Specialization,
		&row.Email,
		&row.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrDoctorNotFound
		}
		return nil, err
	}

	return row.ToDomain(), nil
}

func (r *doctorRepository) GetAll(ctx context.Context) ([]*model.Doctor, error) {
	query := `SELECT id, full_name, specialization, email, created_at FROM doctors`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*model.Doctor
	for rows.Next() {
		var row doctorRow
		if err := rows.Scan(&row.ID, &row.FullName, &row.Specialization, &row.Email, &row.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, row.ToDomain())
	}

	return result, nil
}

func (r *doctorRepository) GetByEmail(ctx context.Context, email string) (*model.Doctor, error) {
	query := `SELECT id, full_name, specialization, email, created_at FROM doctors WHERE email = $1`

	var row doctorRow
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&row.ID,
		&row.FullName,
		&row.Specialization,
		&row.Email,
		&row.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrDoctorNotFound
		}
		return nil, err
	}

	return row.ToDomain(), nil
}
