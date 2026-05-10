package repository

import (
	"context"
	"log"

	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/model"
	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/model/interfaces"
)

type cachedDoctorRepository struct {
	next  interfaces.DoctorRepository
	cache interfaces.Repository
}

func NewCachedDoctorRepository(next interfaces.DoctorRepository, cacheRepo interfaces.Repository) interfaces.DoctorRepository {
	return &cachedDoctorRepository{next: next, cache: cacheRepo}
}

func (r *cachedDoctorRepository) Create(ctx context.Context, doctor *model.Doctor) error {
	if err := r.next.Create(ctx, doctor); err != nil {
		return err
	}

	key := "doctor:" + doctor.ID
	if err := r.cache.Set(ctx, key, doctor); err != nil {
		log.Printf("cache set failed key=%s error=%v", key, err)
	} else {
		logCache("doctor-service", "set", key)
	}

	if err := r.cache.Delete(ctx, "doctors:list"); err != nil {
		log.Printf("cache delete failed key=doctors:list error=%v", err)
	} else {
		logCache("doctor-service", "delete", "doctors:list")
	}

	return nil
}

func (r *cachedDoctorRepository) GetById(ctx context.Context, id string) (*model.Doctor, error) {
	key := "doctor:" + id

	var doctor model.Doctor
	hit, err := r.cache.Get(ctx, key, &doctor)
	if err != nil {
		log.Printf("cache get failed key=%s error=%v", key, err)
	}

	if hit {
		logCache("doctor-service", "hit", key)
		return &doctor, nil
	}

	logCache("doctor-service", "miss", key)

	found, err := r.next.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := r.cache.Set(ctx, key, found); err != nil {
		log.Printf("cache set failed key=%s error=%v", key, err)
	} else {
		logCache("doctor-service", "set", key)
	}

	return found, nil
}

func (r *cachedDoctorRepository) GetAll(ctx context.Context) ([]*model.Doctor, error) {
	const key = "doctors:list"

	var doctors []*model.Doctor
	hit, err := r.cache.Get(ctx, key, &doctors)
	if err != nil {
		log.Printf("cache get failed key=%s error=%v", key, err)
	}

	if hit {
		logCache("doctor-service", "hit", key)
		return doctors, nil
	}

	logCache("doctor-service", "miss", key)

	doctors, err = r.next.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	if err := r.cache.Set(ctx, key, doctors); err != nil {
		log.Printf("cache set failed key=%s error=%v", key, err)
	} else {
		logCache("doctor-service", "set", key)
	}

	return doctors, nil
}

func (r *cachedDoctorRepository) GetByEmail(ctx context.Context, email string) (*model.Doctor, error) {
	return r.next.GetByEmail(ctx, email)
}

func logCache(service, status, key string) {
	log.Printf(`{"service":"%s","component":"cache","status":"%s","key":"%s"}`, service, status, key)
}
