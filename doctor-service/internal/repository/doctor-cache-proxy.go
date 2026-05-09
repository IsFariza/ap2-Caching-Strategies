package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/model"
	"github.com/IsFariza/ap2-Caching-Strategies/doctor-service/internal/model/interfaces"
	"github.com/redis/go-redis/v9"
)

type doctorCacheProxy struct {
	repo interfaces.DoctorRepository
	rdb  *redis.Client
	ttl  time.Duration
}

func NewDoctorCacheProxy(repo interfaces.DoctorRepository, rdb *redis.Client, ttlSeconds int) interfaces.DoctorRepository {
	return &doctorCacheProxy{
		repo: repo,
		rdb:  rdb,
		ttl:  time.Duration(ttlSeconds) * time.Second,
	}
}

func (p *doctorCacheProxy) GetById(ctx context.Context, id string) (*model.Doctor, error) {
	key := fmt.Sprintf("doctor:%s", id)
	val, err := p.rdb.Get(ctx, key).Result()
	if err == nil {
		var doc model.Doctor
		if err := json.Unmarshal([]byte(val), &doc); err == nil {
			return &doc, nil
		}
	}
	doc, err := p.repo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(doc)
	if err := p.rdb.Set(ctx, key, data, p.ttl).Err(); err != nil {
		log.Printf("Cache write failure: %v", err)
	}
	return doc, nil
}

func (p *doctorCacheProxy) GetAll(ctx context.Context) ([]*model.Doctor, error) {
	key := "doctor:list"
	val, err := p.rdb.Get(ctx, key).Result()
	if err == nil {
		var docs []*model.Doctor
		if err := json.Unmarshal([]byte(val), &docs); err == nil {
			return docs, nil
		}
	}
	docs, err := p.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(docs)
	p.rdb.Set(ctx, key, data, p.ttl)

	return docs, nil
}
func (p *doctorCacheProxy) Create(ctx context.Context, doc *model.Doctor) error {

	err := p.repo.Create(ctx, doc)
	if err != nil {
		return err
	}

	if err := p.rdb.Del(ctx, "doctors:list").Err(); err != nil {
		log.Printf("Cache invalidation failure: %v", err)
	}

	return nil
}
func (p *doctorCacheProxy) GetByEmail(ctx context.Context, email string) (*model.Doctor, error) {
	return p.repo.GetByEmail(ctx, email)
}
