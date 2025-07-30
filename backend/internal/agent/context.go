package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/google/uuid"
	"github.com/oklog/ulid"
)

const (
	contextCacheTTL6Hrs  time.Duration = 6 * time.Hour
	contextCacheTTL12Hrs time.Duration = 12 * time.Hour
)

type sessionSummary struct {
	title        string
	lastAccessed time.Time
	summary      string
}

type inputPrompt struct {
	systemInput string
	userInput   string
}

type interaction struct {
	input       inputPrompt
	modelOutput string
}

type contextWindow struct {
	sessionSummaries []sessionSummary
	history          []interaction
	input            inputPrompt
	tokenCount       int
}

type sessionContext struct {
	UserID    uuid.UUID     `json:"user_id"`
	SessionID ulid.ULID     `json:"session_id"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Window    contextWindow `json:"context_window"`
}

func (a *Agent) setContextCache(ctx context.Context, sc *sessionContext) error {
	const method = "setContextCache"
	log := a.logger.With(
		slog.String("method", method),
		slog.String("user_id", sc.UserID.String()),
		slog.String("session_id", sc.SessionID.String()),
		slog.Time("created_at", sc.CreatedAt),
		slog.Time("updated_at", sc.UpdatedAt),
		slog.String("token_count", humanize.Comma(int64(sc.Window.tokenCount))),
	)

	bytes, err := json.Marshal(sc)
	if err != nil {
		log.With(
			"err", err,
		).ErrorContext(ctx, "failed to set context cache")
		return fmt.Errorf("failed to set context cache: %w", err)
	}

	key := createContextCacheKey(sc)

	if err = a.store.Fly.Set(ctx, key, bytes, contextCacheTTL6Hrs); err != nil {
		log.With(
			"err", err,
		).ErrorContext(ctx, "failed to set context cache")
		return fmt.Errorf("failed to set context cache: %w", err)
	}

	log.DebugContext(ctx, "set context cache successfully")

	return nil
}

func (a *Agent) getContextCache(ctx context.Context, key string) (*sessionContext, error) {
	const method = "getContextCache"
	log := a.logger.With(
		slog.String("method", method),
		slog.String("key", key),
	)

	bytes, err := a.store.Fly.Get(ctx, key)
	if err != nil {
		log.With(
			"err", err,
		).ErrorContext(ctx, "failed to get context cache")
		return nil, fmt.Errorf("failed to get context cache: %w", err)
	}

	if bytes == nil {
		log.WarnContext(ctx, "missed context cache: returning nil...")
		return nil, nil
	}

	var sc sessionContext
	if err = json.Unmarshal(bytes, &sc); err != nil {
		log.With(
			"err", err,
		).ErrorContext(ctx, "failed to get context cache")
		return nil, fmt.Errorf("failed to get context cache: %w", err)
	}

	log.With(
		slog.String("user_id", sc.UserID.String()),
		slog.String("session_id", sc.SessionID.String()),
		slog.Time("created_at", sc.CreatedAt),
		slog.Time("updated_at", sc.UpdatedAt),
		slog.String("token_count", humanize.Comma(int64(sc.Window.tokenCount))),
	).DebugContext(ctx, "context cache retrieved successfully")

	return &sc, nil
}

func createContextCacheKey(sc *sessionContext) string {
	return fmt.Sprintf("shaikh:user:%s:session:%s:context", sc.UserID.String(), sc.SessionID.String())
}
