package jobqueue

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/redis/go-redis/v9"
)

func idempotencyKey(eventType, id, occurredAt string) string {
	sum := sha256.Sum256([]byte(eventType + id + occurredAt))
	return hex.EncodeToString(sum[:])
}
func (q *Queue) getKey(ctx context.Context, key string) (string, error) {
	if q.redisOK {
		value, err := q.client.Get(ctx, "notification:idempotency:"+key).Result()
		if err == redis.Nil {
			return "", nil
		}
		return value, err
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.localKeys[key], nil
}

func (q *Queue) reserveKey(ctx context.Context, key string) (bool, error) {
	if q.redisOK {
		return q.client.SetNX(ctx, "notification:idempotency:"+key, "queued", idempotencyTTL).Result()
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.localKeys[key] != "" {
		return false, nil
	}
	q.localKeys[key] = "queued"
	return true, nil
}

func (q *Queue) markDone(ctx context.Context, key string) error {
	if q.redisOK {
		return q.client.Set(ctx, "notification:idempotency:"+key, "done", idempotencyTTL).Err()
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	q.localKeys[key] = "done"
	return nil
}
