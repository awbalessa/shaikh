package pro

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/dom"
	"github.com/redis/go-redis/v9"
)

type DragonflyCache struct {
	Fly *redis.Client
}

func NewDragonflyCache() *DragonflyCache {
	fly := redis.NewClient(&redis.Options{
		ClientName:            os.Getenv("SERVICE_NAME"),
		Addr:                  os.Getenv("DRAGONFLY_ADDR"),
		ContextTimeoutEnabled: true,
		DialTimeout:           1 * time.Second,
		PoolTimeout:           1 * time.Second,
		ReadTimeout:           500 * time.Millisecond,
		WriteTimeout:          500 * time.Millisecond,
		ConnMaxIdleTime:       5 * time.Minute,
		ConnMaxLifetime:       30 * time.Minute,
		PoolSize:              10,
		MinIdleConns:          2,
	})

	return &DragonflyCache{Fly: fly}
}

func (f *DragonflyCache) Set(
	ctx context.Context,
	key string,
	value []byte,
	expr time.Duration,
) error {
	if err := f.Fly.Set(ctx, key, value, expr).Err(); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("dragonfly set: %w", dom.ErrTimeout)
		}
		return fmt.Errorf("dragonfly set: %w", dom.ErrInternal)
	}
	return nil
}

func (f *DragonflyCache) Get(
	ctx context.Context,
	key string,
) ([]byte, error) {
	bytes, err := f.Fly.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // cache miss, not an error
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("dragonfly get: %w", dom.ErrTimeout)
		}
		return nil, fmt.Errorf("dragonfly get: %w", dom.ErrInternal)
	}
	return bytes, nil
}

func (f *DragonflyCache) SetNX(
	ctx context.Context,
	key string,
	value []byte,
	expr time.Duration,
) (bool, error) {
	ok, err := f.Fly.SetNX(ctx, key, value, expr).Result()
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return false, fmt.Errorf("dragonfly setnx: %w", dom.ErrTimeout)
		}
		return false, fmt.Errorf("dragonfly setnx: %w", dom.ErrInternal)
	}
	return ok, nil
}

func (f *DragonflyCache) Del(
	ctx context.Context,
	key string,
) (int64, error) {
	n, err := f.Fly.Del(ctx, key).Result()
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return 0, fmt.Errorf("dragonfly del: %w", dom.ErrTimeout)
		}
		return 0, fmt.Errorf("dragonfly del: %w", dom.ErrInternal)
	}
	return n, nil
}

func (f *DragonflyCache) RefreshTTL(
	ctx context.Context,
	key string,
	ttl time.Duration,
) (bool, error) {
	ok, err := f.Fly.Expire(ctx, key, ttl).Result()
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return false, fmt.Errorf("dragonfly refresh ttl: %w", dom.ErrTimeout)
		}
		return false, fmt.Errorf("dragonfly refresh ttl: %w", dom.ErrInternal)
	}
	return ok, nil
}

func (f *DragonflyCache) Exists(
	ctx context.Context,
	key string,
) (bool, error) {
	n, err := f.Fly.Exists(ctx, key).Result()
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return false, fmt.Errorf("dragonfly exists: %w", dom.ErrTimeout)
		}
		return false, fmt.Errorf("dragonfly exists: %w", dom.ErrInternal)
	}
	return n > 0, nil
}

func (f *DragonflyCache) Ping(ctx context.Context) error {
	res, err := f.Fly.Ping(ctx).Result()
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("dragonfly ping: %w", dom.ErrTimeout)
		}
		return fmt.Errorf("dragonfly ping: %w", dom.ErrUnavailable)
	}
	if res != "PONG" {
		return fmt.Errorf("dragonfly ping unexpected: %s", res)
	}
	return nil
}

func (f *DragonflyCache) Close() error {
	if f.Fly != nil {
		return f.Fly.Close()
	}
	return nil
}

func (f *DragonflyCache) Name() string {
	return "cache"
}
