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
	contextCacheTTL6Hrs  time.Duration = 6 * time.Hour
	contextCacheTTL12Hrs time.Duration = 12 * time.Hour
)

type sessionSummary struct {
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
	previousSessions []sessionSummary
	history          []interaction
	tokenCount       int
}

type gcc struct {
	resourceName string
	createdAt    time.Time
}

type sessionContext struct {
	UserID             uuid.UUID     `json:"user_id"`
	SessionID          ulid.ULID     `json:"session_id"`
	GeminiContextCache gcc           `json:"gcc"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
	Window             contextWindow `json:"context_window"`
}

func (a *Agent) buildContextWindow(cw *contextWindow) []*genai.Content {
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

	// --- 2. Interaction History ---
	for _, inter := range cw.history {
		// Combine userInput and functionResponse into a single user content
		var userParts []*genai.Part
		if inter.input.functionResponse != nil {
			userParts = append(userParts, inter.input.functionResponse)
		}
		if inter.input.userInput != nil {
			userParts = append(userParts, inter.input.userInput)
		}

		if len(userParts) > 0 {
			contents = append(contents, &genai.Content{
				Role:  genai.RoleUser,
				Parts: userParts,
			})
		}

		if inter.modelOutput != nil {
			contents = append(contents, &genai.Content{
				Role:  genai.RoleModel,
				Parts: []*genai.Part{inter.modelOutput},
			})
		}
	}

	return contents
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

// func (a *Agent) setGeminiContextCache(ctx context.Context, sc *sessionContext) error {
// 	config := &genai.CreateCachedContentConfig{}
// }

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
