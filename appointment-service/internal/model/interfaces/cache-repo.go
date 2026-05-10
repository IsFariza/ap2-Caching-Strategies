package interfaces

import "context"

type Repository interface {
	Get(ctx context.Context, key string, dest any) (bool, error)
	Set(ctx context.Context, key string, value any) error
	Delete(ctx context.Context, keys ...string) error
}
