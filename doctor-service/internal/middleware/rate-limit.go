package middleware

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type RateLimiter struct {
	client  *redis.Client
	limit   int64
	enabled bool
}

func NewRateLimiter(ctx context.Context) *RateLimiter {
	limit := int64(100)
	if raw := os.Getenv("RATE_LIMIT_RPM"); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	url := os.Getenv("REDIS_URL")
	if url == "" {
		log.Printf("warning: REDIS_URL is empty; rate limiter disabled")
		return &RateLimiter{limit: limit}
	}

	opts, err := redis.ParseURL(url)
	if err != nil {
		opts = &redis.Options{Addr: url}
	}
	client := redis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("warning: redis unavailable; rate limiter disabled: %v", err)
		return &RateLimiter{client: client, limit: limit}
	}
	return &RateLimiter{client: client, limit: limit, enabled: true}
}

func (r *RateLimiter) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if !r.enabled {
			return handler(ctx, req)
		}

		clientIP := "unknown"
		if p, ok := peer.FromContext(ctx); ok {
			host, _, err := net.SplitHostPort(p.Addr.String())
			if err == nil {
				clientIP = host
			} else {
				clientIP = p.Addr.String()
			}
		}

		now := time.Now()
		nowMillis := now.UnixMilli()
		windowStart := now.Add(-time.Minute).UnixMilli()
		key := "rate_limit:doctor:" + clientIP

		if err := r.client.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10)).Err(); err != nil {
			log.Printf("rate limiter redis error key=%s error=%v", key, err)
			return handler(ctx, req)
		}

		count, err := r.client.ZCard(ctx, key).Result()
		if err != nil {
			log.Printf("rate limiter redis error key=%s error=%v", key, err)
			return handler(ctx, req)
		}
		log.Printf(`{"service":"doctor-service","component":"rate_limiter","algorithm":"sliding_window","client_ip":"%s","key":"%s","count":%d,"limit":%d}`, clientIP, key, count, r.limit)

		if count >= r.limit {
			retryAfter := 60
			oldest, err := r.client.ZRangeWithScores(ctx, key, 0, 0).Result()
			if err == nil && len(oldest) > 0 {
				waitMillis := int64(oldest[0].Score) + int64(time.Minute/time.Millisecond) - nowMillis
				if waitMillis > 0 {
					retryAfter = int(waitMillis/time.Second.Milliseconds()) + 1
				}
			}
			log.Printf(`{"service":"doctor-service","component":"rate_limiter","algorithm":"sliding_window","status":"blocked","client_ip":"%s","retry_after_seconds":%d}`, clientIP, retryAfter)
			return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded; retry after %d seconds", retryAfter)
		}

		member := fmt.Sprintf("%d-%d", nowMillis, now.UnixNano())
		if err := r.client.ZAdd(ctx, key, redis.Z{Score: float64(nowMillis), Member: member}).Err(); err != nil {
			log.Printf("rate limiter redis error key=%s error=%v", key, err)
			return handler(ctx, req)
		}
		_ = r.client.Expire(ctx, key, time.Minute).Err()

		return handler(ctx, req)
	}
}
