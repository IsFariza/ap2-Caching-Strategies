package idempotency

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store struct {
	rdb *redis.Client
}

func NewStore(rdb *redis.Client) *Store {
	return &Store{rdb: rdb}
}
func (s *Store) IsUnique(ctx context.Context, key string) (bool, error) {
	return s.rdb.SetNX(ctx, "idempotency:"+key, "processed", 24*time.Hour).Result()
}
