package dragonfly

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

func (f *dragonflyClient) Set(
	ctx context.Context,
	key string,
	value []byte,
	expr time.Duration,
) error {
	const method = "Set"
	log := f.logger.With(
		slog.String("method", method),
		slog.String("key", key),
		slog.String("expiration", expr.String()),
	)

	cmd := f.cli.Set(ctx, key, value, expr)
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

func (f *dragonflyClient) Get(
	ctx context.Context,
	key string,
) ([]byte, error) {
	const method = "Get"
	log := f.logger.With(
		slog.String("method", method),
		slog.String("key", key),
	)

	cmd := f.cli.Get(ctx, key)
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

func (f *dragonflyClient) SetNX(
	ctx context.Context,
	key string,
	value []byte,
	expr time.Duration,
) (bool, error) {
	const method = "SetNX"
	log := f.logger.With(
		slog.String("method", method),
		slog.String("key", key),
		slog.String("expiration", expr.String()),
	)

	cmd := f.cli.SetNX(ctx, key, value, expr)
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

func (f *dragonflyClient) Del(
	ctx context.Context,
	key string,
) error {
	const method = "Del"
	log := f.logger.With(
		slog.String("method", method),
		slog.String("key", key),
	)

	cmd := f.cli.Del(ctx, key)
	n, err := cmd.Result()
	if err != nil {
		log.With("err", err).ErrorContext(ctx, "failed to delete cache")
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	log.With(slog.Int64("keys_deleted", n)).DebugContext(ctx, "cache key deleted")
	return nil
}

func (f *dragonflyClient) RefreshTTL(
	ctx context.Context,
	key string,
	ttl time.Duration,
) error {
	const method = "RefreshTTL"
	log := f.logger.With(
		slog.String("method", method),
		slog.String("key", key),
		slog.String("new_ttl", ttl.String()),
	)

	cmd := f.cli.Expire(ctx, key, ttl)
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

func (f *dragonflyClient) Exists(
	ctx context.Context,
	key string,
) (bool, error) {
	const method = "Exists"
	log := f.logger.With(
		slog.String("method", method),
		slog.String("key", key),
	)

	cmd := f.cli.Exists(ctx, key)
	n, err := cmd.Result()
	if err != nil {
		log.With("err", err).ErrorContext(ctx, "failed to check existence")
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}

	exists := n > 0
	log.With(slog.Bool("exists", exists)).DebugContext(ctx, "key existence checked")
	return exists, nil
}
