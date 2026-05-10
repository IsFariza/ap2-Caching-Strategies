package jobqueue

import (
	"context"
	"os"
	"time"

	"github.com/IsFariza/ap2-Caching-Strategies/notification-service/internal/models"
)

func (q *Queue) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-q.jobs:
			q.process(ctx, job)
		}
	}
}

func (q *Queue) process(ctx context.Context, job models.Job) {
	backoffs := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second}
	var lastErr error

	for attempt := 1; attempt <= 3; attempt++ {
		q.log(os.Stdout, "info", job.IdempotencyKey, attempt, "processing", nil)
		err := q.callGateway(ctx, job)
		if err == nil {
			if markErr := q.markDone(ctx, job.IdempotencyKey); markErr != nil {
				q.log(os.Stdout, "warn", job.IdempotencyKey, attempt, "retry", markErr)
			}
			q.log(os.Stdout, "info", job.IdempotencyKey, attempt, "success", nil)
			return
		}

		lastErr = err
		q.log(os.Stdout, "warn", job.IdempotencyKey, attempt, "retry", err)
		if attempt < 3 {
			time.Sleep(backoffs[attempt-1])
		}
	}

	q.log(os.Stderr, "error", job.IdempotencyKey, 3, "dead_letter", lastErr)
}
