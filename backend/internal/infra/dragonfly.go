package infra

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/config"
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
	Log *slog.Logger
}

func NewDragonflyCache(env *config.Env) *DragonflyCache {
	log := slog.Default().With(
		"component", "dragonfly",
	)

	fly := redis.NewClient(&redis.Options{
		Addr:                  env.DragonFlyAddress,
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
		Log: log,
	}
}

func (f *DragonflyCache) Set(
	ctx context.Context,
	key string,
	value []byte,
	expr time.Duration,
) error {
	const method = "Set"
	log := f.Log.With(
		slog.String("method", method),
		slog.String("key", key),
		slog.String("expiration", expr.String()),
	)

	cmd := f.Fly.Set(ctx, key, value, expr)
	if err := cmd.Err(); err != nil {
		log.With(
			"err", err,
		).ErrorContext(ctx, "failed to set cache")
		return fmt.Errorf("failed to set cache: %w", err)
	}

	log.With(
		slog.String("status", cmd.Val()),
	).DebugContext(ctx, "cache set successfully")

	return nil
}

func (f *DragonflyCache) Get(
	ctx context.Context,
	key string,
) ([]byte, error) {
	const method = "Get"
	log := f.Log.With(
		slog.String("method", method),
		slog.String("key", key),
	)

	cmd := f.Fly.Get(ctx, key)
	bytes, err := cmd.Bytes()
	if err != nil {
		if err == redis.Nil {
			log.With(
				"err", err,
			).WarnContext(ctx, "cache miss")
			return nil, nil
		}
		log.With("err", err).ErrorContext(ctx, "failed to get cache")
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	log.DebugContext(ctx, "cache retrieved successfully")

	return bytes, nil
}

func (f *DragonflyCache) SetNX(
	ctx context.Context,
	key string,
	value []byte,
	expr time.Duration,
) (bool, error) {
	const method = "SetNX"
	log := f.Log.With(
		slog.String("method", method),
		slog.String("key", key),
		slog.String("expiration", expr.String()),
	)

	cmd := f.Fly.SetNX(ctx, key, value, expr)
	ok, err := cmd.Result()
	if err != nil {
		log.With(
			"err", err,
		).ErrorContext(ctx, "failed to setnx cache")
		return false, fmt.Errorf("failed to setnx cache: %w", err)
	}

	log.With(
		slog.Bool("was_set", cmd.Val()),
	).DebugContext(ctx, "cache setnx successfully")

	return ok, nil
}

func (f *DragonflyCache) Del(
	ctx context.Context,
	key string,
) error {
	const method = "Del"
	log := f.Log.With(
		slog.String("method", method),
		slog.String("key", key),
	)

	cmd := f.Fly.Del(ctx, key)
	n, err := cmd.Result()
	if err != nil {
		log.With("err", err).ErrorContext(ctx, "failed to delete cache")
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	log.With(slog.Int64("keys_deleted", n)).DebugContext(ctx, "cache key deleted")
	return nil
}

func (f *DragonflyCache) RefreshTTL(
	ctx context.Context,
	key string,
	ttl time.Duration,
) error {
	const method = "RefreshTTL"
	log := f.Log.With(
		slog.String("method", method),
		slog.String("key", key),
		slog.String("new_ttl", ttl.String()),
	)

	cmd := f.Fly.Expire(ctx, key, ttl)
	ok, err := cmd.Result()
	if err != nil {
		log.With("err", err).ErrorContext(ctx, "failed to refresh TTL")
		return fmt.Errorf("failed to refresh TTL: %w", err)
	}

	if !ok {
		log.WarnContext(ctx, "key not found to refresh TTL")
		return nil
	}

	log.DebugContext(ctx, "TTL refreshed successfully")
	return nil
}

func (f *DragonflyCache) Exists(
	ctx context.Context,
	key string,
) (bool, error) {
	const method = "Exists"
	log := f.Log.With(
		slog.String("method", method),
		slog.String("key", key),
	)

	cmd := f.Fly.Exists(ctx, key)
	n, err := cmd.Result()
	if err != nil {
		log.With("err", err).ErrorContext(ctx, "failed to check existence")
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}

	exists := n > 0
	log.With(slog.Bool("exists", exists)).DebugContext(ctx, "key existence checked")
	return exists, nil
}
