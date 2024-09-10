package ratelimiter

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type (
	Strategy interface {
		Inc(ctx context.Context, value string) (int64, error)
	}

	redisStrategy struct {
		key      string
		rc       *redis.Client
		interval time.Duration
	}
)

func NewRedisStrategy(client *redis.Client, key string, interval time.Duration) Strategy {
	return &redisStrategy{
		key:      key,
		rc:       client,
		interval: interval,
	}
}

func (r *redisStrategy) Inc(ctx context.Context, value string) (int64, error) {
	key := r.key + value

	p := r.rc.Pipeline()
	inc := p.Incr(ctx, key)
	p.Expire(ctx, key, r.interval)
	_, err := p.Exec(ctx)
	if err != nil {
		return 0, err
	}

	count, err := inc.Result()
	if err != nil {
		return 0, err
	}

	return count, nil
}
