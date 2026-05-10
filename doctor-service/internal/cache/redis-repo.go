package cache

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client  *redis.Client
	ttl     time.Duration
	enabled bool
}

func NewRedisRepository(ctx context.Context) *RedisRepository {
	ttl := 60 * time.Second
	if raw := os.Getenv("CACHE_TTL_SECONDS"); raw != "" {
		if seconds, err := strconv.Atoi(raw); err == nil && seconds > 0 {
			ttl = time.Duration(seconds) * time.Second
		}
	}

	url := os.Getenv("REDIS_URL")
	if url == "" {
		log.Printf("warning: REDIS_URL is empty; cache disabled")
		return &RedisRepository{ttl: ttl}
	}

	opts, err := redis.ParseURL(url)
	if err != nil {
		opts = &redis.Options{Addr: url}
	}

	client := redis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("warning: redis unavailable; cache disabled: %v", err)
		return &RedisRepository{client: client, ttl: ttl}
	}

	return &RedisRepository{client: client, ttl: ttl, enabled: true}
}

func (r *RedisRepository) Get(ctx context.Context, key string, dest any) (bool, error) {
	if !r.enabled {
		return false, nil
	}

	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return false, err
	}
	return true, nil
}

func (r *RedisRepository) Set(ctx context.Context, key string, value any) error {
	if !r.enabled {
		return nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, r.ttl).Err()
}

func (r *RedisRepository) Delete(ctx context.Context, keys ...string) error {
	if !r.enabled || len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}
