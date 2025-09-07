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
			return dom.NewTaggedError(dom.ErrTimeout, err)
		}
		return dom.NewTaggedError(dom.ErrInternal, err)
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
			return nil, dom.NewTaggedError(dom.ErrTimeout, err)
		}
		return nil, dom.NewTaggedError(dom.ErrInternal, err)
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
			return false, dom.NewTaggedError(dom.ErrTimeout, err)
		}
		return false, dom.NewTaggedError(dom.ErrInternal, err)
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
			return 0, dom.NewTaggedError(dom.ErrTimeout, err)
		}
		return 0, dom.NewTaggedError(dom.ErrInternal, err)
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
			return false, dom.NewTaggedError(dom.ErrTimeout, err)
		}
		return false, dom.NewTaggedError(dom.ErrInternal, err)
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
			return false, dom.NewTaggedError(dom.ErrTimeout, err)
		}
		return false, dom.NewTaggedError(dom.ErrInternal, err)
	}
	return n > 0, nil
}

func (f *DragonflyCache) Ping(ctx context.Context) error {
	res, err := f.Fly.Ping(ctx).Result()
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return dom.NewTaggedError(dom.ErrTimeout, err)
		}
		return dom.NewTaggedError(dom.ErrUnavailable, err)
	}
	if res != "PONG" {
		return dom.NewTaggedError(dom.ErrInternal, fmt.Errorf("unexpected ping response: %s", res))
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