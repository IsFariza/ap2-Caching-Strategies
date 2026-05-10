package repository

import (
	"context"
	"log"

	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/model"
	"github.com/IsFariza/ap2-Caching-Strategies/appointment-service/internal/model/interfaces"
)

type cachedAppointmentRepo struct {
	next  interfaces.AppointmentRepo
	cache interfaces.Repository
}

func NewCachedAppointmentRepository(next interfaces.AppointmentRepo, cacheRepo interfaces.Repository) interfaces.AppointmentRepo {
	return &cachedAppointmentRepo{next: next, cache: cacheRepo}
}

func (r *cachedAppointmentRepo) Create(ctx context.Context, appointment *model.Appointment) error {
	if err := r.next.Create(ctx, appointment); err != nil {
		return err
	}

	if err := r.cache.Delete(ctx, "appointments:list"); err != nil {
		log.Printf("cache delete failed key=appointments:list error=%v", err)
	} else {
		logCache("appointment-service", "delete", "appointments:list")
	}

	return nil
}

func (r *cachedAppointmentRepo) GetById(ctx context.Context, id string) (*model.Appointment, error) {
	key := "appointment:" + id
	var appointment model.Appointment
	hit, err := r.cache.Get(ctx, key, &appointment)
	if err != nil {
		log.Printf("cache get failed key=%s error=%v", key, err)
	}
	if hit {
		logCache("appointment-service", "hit", key)
		return &appointment, nil
	}
	logCache("appointment-service", "miss", key)

	found, err := r.next.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := r.cache.Set(ctx, key, found); err != nil {
		log.Printf("cache set failed key=%s error=%v", key, err)
	} else {
		logCache("appointmet-service", "set", key)
	}
	return found, nil
}

func (r *cachedAppointmentRepo) GetAll(ctx context.Context) ([]*model.Appointment, error) {
	const key = "appointments:list"
	var appointments []*model.Appointment
	hit, err := r.cache.Get(ctx, key, &appointments)
	if err != nil {
		log.Printf("cache get failed key=%s error=%v", key, err)
	}
	if hit {
		logCache("appointment-service", "hit", key)
		return appointments, nil
	}
	logCache("appointment-service", "miss", key)

	appointments, err = r.next.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	if err := r.cache.Set(ctx, key, appointments); err != nil {
		log.Printf("cache set failed key=%s error=%v", key, err)
	} else {
		logCache("appointment-service", "set", key)
	}
	return appointments, nil
}

func (r *cachedAppointmentRepo) Update(ctx context.Context, id string, newStatus model.Status) error {
	if err := r.next.Update(ctx, id, newStatus); err != nil {
		return err
	}

	key := "appointment:" + id

	updated, err := r.next.GetById(ctx, id)
	if err != nil {
		log.Printf("cache refresh skipped key=%s error=%v", key, err)
	} else if err := r.cache.Set(ctx, key, updated); err != nil {
		log.Printf("cache set failed key=%s error=%v", key, err)
	} else {
		logCache("appointment-service", "set", key)
	}

	if err := r.cache.Delete(ctx, "appointments:list"); err != nil {
		log.Printf("cache delete failed key=appointments:list error=%v", err)
	} else {
		logCache("appointment-service", "delete", "appointments:list")
	}

	return nil
}

func logCache(service, status, key string) {
	log.Printf(`{"service":"%s","component":"cache","status":"%s","key":"%s"}`, service, status, key)
}
