package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	rdb    *redis.Client
	limit  int
	window time.Duration
}

func NewRateLimiter(rdb *redis.Client, perMinute int) *RateLimiter {
	return &RateLimiter{rdb: rdb, limit: perMinute, window: time.Minute}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := realIP(r)
		key := fmt.Sprintf("rate:auth:%s", ip)
		ctx := r.Context()

		count, ttl, err := rl.increment(ctx, key)
		if err != nil {
			// Redis unavailable — fail open
			next.ServeHTTP(w, r)
			return
		}

		if count > int64(rl.limit) {
			w.Header().Set("Retry-After", strconv.FormatInt(int64(ttl.Seconds())+1, 10))
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
			w.Header().Set("X-RateLimit-Remaining", "0")
			http.Error(w, `{"error":{"code":"rate_limit_exceeded","message":"too many requests, try again later"}}`, http.StatusTooManyRequests)
			return
		}

		remaining := int64(rl.limit) - count
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		next.ServeHTTP(w, r)
	})
}

// increment atomically increments the counter and sets expiry only on first hit.
// Returns (count, ttl, error).
func (rl *RateLimiter) increment(ctx context.Context, key string) (int64, time.Duration, error) {
	pipe := rl.rdb.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.ExpireNX(ctx, key, rl.window)
	ttlCmd := pipe.TTL(ctx, key)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, 0, err
	}
	return incr.Val(), ttlCmd.Val(), nil
}

func realIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
