package jobqueue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/IsFariza/ap2-Caching-Strategies/notification-service/internal/models"
	"github.com/redis/go-redis/v9"
)

type WorkerPool struct {
	rdb    *redis.Client
	jobs   chan models.Job
	client *http.Client
}

func NewWorkerPool(rdb *redis.Client, bufferSize int) *WorkerPool {
	return &WorkerPool{
		rdb:    rdb,
		jobs:   make(chan models.Job, bufferSize),
		client: &http.Client{Timeout: 5 * time.Second},
	}
}
func (wp *WorkerPool) Start(workerCount int) {
	for i := 0; i < workerCount; i++ {
		go wp.worker()
	}
}
func (wp *WorkerPool) Enqueue(job models.Job) {
	wp.jobs <- job
}
func (wp *WorkerPool) worker() {
	for job := range wp.jobs {
		lockKey := fmt.Sprintf("notified:%s", job.AppointmentID)
		success, err := wp.rdb.SetNX(context.Background(), lockKey, "processing", 24*time.Hour).Result()
		if err != nil || !success {
			log.Printf("Job for %s already processed or error: %v", job.AppointmentID, err)
			continue
		}
		err = wp.executewithRetry(job)
		if err != nil {
			logEntry := map[string]interface{}{
				"time":           time.Now().Format(time.RFC3339),
				"level":          "ERROR",
				"appointment_id": job.AppointmentID,
				"message":        "Permanent failure: Job moved to dead letter",
				"error":          err.Error(),
			}
			data, _ := json.Marshal(logEntry)
			fmt.Fprintln(os.Stderr, string(data))
		} else {
			wp.rdb.Set(context.Background(), lockKey, "completed", 24*time.Hour)
		}

	}
}
