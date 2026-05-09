package middleware

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func RateLimitInterceptor(rdb *redis.Client, limit int) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		if rdb == nil {
			return handler(ctx, req)
		}
		p, ok := peer.FromContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Internal, "peer info missing")
		}

		host, _, err := net.SplitHostPort(p.Addr.String())
		if err != nil {
			host = p.Addr.String()
		}
		key := fmt.Sprintf("rate_limit:%s", host)
		now := time.Now().UnixNano()
		window := now - int64(time.Minute)

		pipe := rdb.Pipeline()
		pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", window))
		pipe.ZCard(ctx, key)
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
		pipe.Expire(ctx, key, time.Minute)

		cmds, err := pipe.Exec(ctx)
		if err != nil {
			fmt.Printf("rate limiter error: %v\n", err)
			return handler(ctx, req)
		}
		count := cmds[1].(*redis.IntCmd).Val()

		if int(count) >= limit {
			return nil, status.Errorf(codes.ResourceExhausted, "limit of %d RPM exceeded", limit)
		}

		return handler(ctx, req)
	}
}
