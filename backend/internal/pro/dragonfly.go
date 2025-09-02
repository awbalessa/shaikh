package pro

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	flyDialTimout      time.Duration = 1 * time.Second
	flyPoolTimeout     time.Duration = 1 * time.Second
	flyReadTimeout     time.Duration = 500 * time.Millisecond
	flyWriteTimeout    time.Duration = 500 * time.Millisecond
	flyConnMaxIdleTime time.Duration = 5 * time.Minute
	flyConnMaxLifetime time.Duration = 30 * time.Minute
	flyPoolSize        int           = 10
	flyMinIdleConns    int           = 2
)

type DragonflyCache struct {
	Fly *redis.Client
}

func NewDragonflyCache() *DragonflyCache {
	fly := redis.NewClient(&redis.Options{
		Addr:                  os.Getenv("DRAGONFLY_ADDR"),
		ContextTimeoutEnabled: true,
		DialTimeout:           flyDialTimout,
		PoolTimeout:           flyPoolTimeout,
		ReadTimeout:           flyReadTimeout,
		WriteTimeout:          flyWriteTimeout,
		ConnMaxIdleTime:       flyConnMaxIdleTime,
		ConnMaxLifetime:       flyConnMaxLifetime,
		PoolSize:              flyPoolSize,
		MinIdleConns:          flyMinIdleConns,
		ClientName:            "shaikh-api",
	})

	return &DragonflyCache{
		Fly: fly,
	}
}

func (f *DragonflyCache) Set(
	ctx context.Context,
	key string,
	value []byte,
	expr time.Duration,
) error {
	cmd := f.Fly.Set(ctx, key, value, expr)
	if err := cmd.Err(); err != nil {
		return err
	}

	return nil
}

func (f *DragonflyCache) Get(
	ctx context.Context,
	key string,
) ([]byte, error) {
	cmd := f.Fly.Get(ctx, key)
	bytes, err := cmd.Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	return bytes, nil
}

func (f *DragonflyCache) SetNX(
	ctx context.Context,
	key string,
	value []byte,
	expr time.Duration,
) (bool, error) {
	cmd := f.Fly.SetNX(ctx, key, value, expr)
	ok, err := cmd.Result()
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (f *DragonflyCache) Del(
	ctx context.Context,
	key string,
) (int64, error) {
	cmd := f.Fly.Del(ctx, key)

	n, err := cmd.Result()
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (f *DragonflyCache) RefreshTTL(
	ctx context.Context,
	key string,
	ttl time.Duration,
) (bool, error) {
	cmd := f.Fly.Expire(ctx, key, ttl)
	ok, err := cmd.Result()
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (f *DragonflyCache) Exists(
	ctx context.Context,
	key string,
) (bool, error) {
	cmd := f.Fly.Exists(ctx, key)
	n, err := cmd.Result()
	if err != nil {
		return false, err
	}

	exists := n > 0
	return exists, nil
}

func (f *DragonflyCache) Ping(ctx context.Context) error {
	res, err := f.Fly.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("dragonfly ping failed: %w", err)
	}

	if res != "PONG" {
		return fmt.Errorf("dragonfly ping unexpected response: %s", res)
	}

	return nil
}

func (f *DragonflyCache) Name() string {
	return "cache"
}
