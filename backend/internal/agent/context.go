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
	"google.golang.org/genai"
)

const (
	gccTTL30Mins         time.Duration = 30 * time.Minute
	gccTTL1Hr            time.Duration = 1 * time.Hour
	contextCacheTTL6Hrs  time.Duration = 6 * time.Hour
	contextCacheTTL12Hrs time.Duration = 12 * time.Hour
)

type previousSession struct {
	title        string
	lastAccessed time.Time
	summary      string
}

type inputPrompt struct {
	functionResponse *genai.Part
	userInput        *genai.Part
}

type interaction struct {
	input       inputPrompt
	modelOutput *genai.Part
}

type contextWindow struct {
	previousSessions []previousSession
	history          []interaction
	tokenCount       int
}

type gcc struct {
	resourceName string
	expiresAt    time.Time
}

type sessionContext struct {
	UserID             uuid.UUID     `json:"user_id"`
	SessionID          ulid.ULID     `json:"session_id"`
	GeminiContextCache gcc           `json:"gcc"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
	Window             contextWindow `json:"context_window"`
}

func (a *Agent) buildContextWindow(
	ctx context.Context,
	cw *contextWindow,
	agent agentName,
) ([]*genai.Content, error) {
	prof, err := a.getProfile(agent)
	if err != nil {
		return nil, fmt.Errorf("failed to build context window: %w", err)
	}

	var contents []*genai.Content

	if len(cw.previousSessions) > 0 {
		var parts []*genai.Part
		for _, s := range cw.previousSessions {
			partText := fmt.Sprintf("Session Title: %s\nLast Accessed: %s\nSummary: %s",
				s.title,
				formatRecency(s.lastAccessed),
				s.summary,
			)
			parts = append(parts, genai.NewPartFromText(partText))
		}
		contents = append(contents, &genai.Content{
			Role:  genai.RoleUser,
			Parts: parts,
		})
	}

	historyContents := make([]*genai.Content, 0, len(cw.history)*2)
	for _, inter := range cw.history {
		var userParts []*genai.Part
		if inter.input.functionResponse != nil {
			userParts = append(userParts, inter.input.functionResponse)
		}
		if inter.input.userInput != nil {
			userParts = append(userParts, inter.input.userInput)
		}
		if len(userParts) > 0 {
			historyContents = append(historyContents, &genai.Content{
				Role:  genai.RoleUser,
				Parts: userParts,
			})
		}
		if inter.modelOutput != nil {
			historyContents = append(historyContents, &genai.Content{
				Role:  genai.RoleModel,
				Parts: []*genai.Part{inter.modelOutput},
			})
		}
	}

	ctc := &genai.CountTokensConfig{
		SystemInstruction: prof.config.SystemInstruction,
		Tools:             prof.config.Tools,
	}

	for {
		fullContext := append(contents, historyContents...)

		tokResp, err := a.gc.client.Models.CountTokens(ctx, string(prof.model), fullContext, ctc)
		if err != nil {
			return nil, fmt.Errorf("failed to build context window: %w", err)
		}

		if tokResp.TotalTokens < 200_000 {
			contents = fullContext
			break
		}

		if len(historyContents) > 1 {
			historyContents = historyContents[2:]
		} else {
			historyContents = nil
			break
		}
	}

	return contents, nil
}

func (a *Agent) setgcc(
	ctx context.Context,
	cw []*genai.Content,
	agent agentName,
	key string,
) (*gcc, error) {
	prof, err := a.getProfile(agent)
	if err != nil {
		return nil, fmt.Errorf("failed to set gemini context cache: %w", err)
	}

	conf := &genai.CreateCachedContentConfig{
		ExpireTime:        time.Now().Add(gccTTL30Mins),
		DisplayName:       key,
		Contents:          cw,
		SystemInstruction: prof.config.SystemInstruction,
		Tools:             prof.config.Tools,
	}

	res, err := a.gc.client.Caches.Create(ctx, string(prof.model), conf)
	if err != nil {
		return nil, fmt.Errorf("failed to set gemini context cache: %w", err)
	}

	out := &gcc{
		resourceName: res.Name,
		expiresAt:    res.ExpireTime,
	}

	return out, nil
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
		slog.String("gcc_resource_name", sc.GeminiContextCache.resourceName),
		slog.Duration("gcc_ttl", time.Until(sc.GeminiContextCache.expiresAt)),
	)

	bytes, err := json.Marshal(sc)
	if err != nil {
		log.With(
			"err", err,
		).ErrorContext(ctx, "failed to set context cache")
		return fmt.Errorf("failed to set context cache: %w", err)
	}

	key := createContextCacheKey(sc.UserID, sc.SessionID)

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
		slog.String("gcc_resource_name", sc.GeminiContextCache.resourceName),
		slog.Duration("gcc_ttl", time.Until(sc.GeminiContextCache.expiresAt)),
	).DebugContext(ctx, "context cache retrieved successfully")

	return &sc, nil
}

func createContextCacheKey(userID uuid.UUID, sessionID ulid.ULID) string {
	return fmt.Sprintf("shaikh:user:%s:session:%s:context", userID.String(), sessionID.String())
}

func formatRecency(t time.Time) string {
	duration := time.Since(t)

	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	case duration < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	case duration < 7*24*time.Hour:
		return fmt.Sprintf("%d days ago", int(duration.Hours()/24))
	case duration < 30*24*time.Hour:
		return fmt.Sprintf("%d weeks ago", int(duration.Hours()/(24*7)))
	default:
		return t.Format("2006-01-02")
	}
}
