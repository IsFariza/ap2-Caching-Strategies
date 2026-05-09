package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/model"
	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/model/interfaces"
	"github.com/redis/go-redis/v9"
)

type appointmentCacheProxy struct {
	repo interfaces.AppointmentRepo
	rdb  *redis.Client
	ttl  time.Duration
}

func NewAppointmentCacheProxy(repo interfaces.AppointmentRepo, rdb *redis.Client, ttlSeconds int) interfaces.AppointmentRepo {
	return &appointmentCacheProxy{
		repo: repo,
		rdb:  rdb,
		ttl:  time.Duration(ttlSeconds) * time.Second,
	}
}

func (p *appointmentCacheProxy) GetById(ctx context.Context, id string) (*model.Appointment, error) {
	key := fmt.Sprintf("appointment:%s", id)

	val, err := p.rdb.Get(ctx, key).Result()
	if err == nil {
		var appt model.Appointment
		if err := json.Unmarshal([]byte(val), &appt); err == nil {
			return &appt, nil
		}
	}
	appt, err := p.repo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(appt)
	p.rdb.Set(ctx, key, data, p.ttl)
	return appt, nil
}

func (p *appointmentCacheProxy) GetAll(ctx context.Context) ([]*model.Appointment, error) {
	key := "appointments:list"

	val, err := p.rdb.Get(ctx, key).Result()
	if err == nil {
		var appts []*model.Appointment
		if err := json.Unmarshal([]byte(val), &appts); err == nil {
			return appts, nil
		}
	}
	appts, err := p.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(appts)
	p.rdb.Set(ctx, key, data, p.ttl)
	return appts, nil
}

func (p *appointmentCacheProxy) Create(ctx context.Context, appt *model.Appointment) error {
	if err := p.repo.Create(ctx, appt); err != nil {
		return err
	}
	p.rdb.Del(ctx, "appointments:list")
	return nil
}

func (p *appointmentCacheProxy) Update(ctx context.Context, id string, newStatus model.Status) error {
	if err := p.repo.Update(ctx, id, newStatus); err != nil {
		return err
	}
	p.rdb.Del(ctx, fmt.Sprintf("appointment:%s", id), "appointments:list")
	return nil
}
