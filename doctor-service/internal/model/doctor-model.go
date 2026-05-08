package model

import "time"

type Doctor struct {
	ID             string
	FullName       string
	Specialization string
	Email          string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
