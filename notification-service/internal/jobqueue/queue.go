package jobqueue

import (
	"context"
	"net/http"

	"github.com/IsFariza/ap2-Caching-Strategies/notification-service/internal/jobqueue/idempotency"
	"github.com/IsFariza/ap2-Caching-Strategies/notification-service/internal/models"
	"github.com/redis/go-redis/v9"
)

type Manager struct {
	pool  *WorkerPool
	store *idempotency.Store
}

func NewManager(rdb *redis.Client, bufferSize int) *Manager {
	return &Manager{
		pool: &WorkerPool{
			rdb:    rdb,
			jobs:   make(chan models.Job, bufferSize),
			client: &http.Client{},
		},
		store: idempotency.NewStore(rdb),
	}
}
func (m *Manager) Start(workers int) {
	m.pool.Start(workers)
}
func (m *Manager) Dispatch(job models.Job) {
	unique, err := m.store.IsUnique(context.Background(), job.AppointmentID)
	if err == nil && unique {
		m.pool.Enqueue(job)
	}
}
