package store

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

func (f *dragonflyClient) Set(
	ctx context.Context,
	key string,
	value any,
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
