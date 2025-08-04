package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/database"
	"github.com/dustin/go-humanize"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/sync/errgroup"
	"google.golang.org/genai"
)

const (
	gccTTL30Mins         time.Duration = 30 * time.Minute
	gccTTL1Hr            time.Duration = 1 * time.Hour
	contextCacheTTL6Hrs  time.Duration = 6 * time.Hour
	contextCacheTTL12Hrs time.Duration = 12 * time.Hour
	memories100          int           = 100
	sessions5            int           = 5
)

type userMemory struct {
	updatedAt time.Time
	memory    string
}

type previousSession struct {
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
	userMemories     []userMemory
	previousSessions []previousSession
	history          []interaction
}

type gcc struct {
	resourceName string
	expiresAt    time.Time
}

type sessionContext struct {
	UserID    uuid.UUID          `json:"user_id"`
	SessionID uuid.UUID          `json:"session_id"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
	GCCMap    map[agentName]*gcc `json:"gcc_map"`
	Window    *contextWindow     `json:"context_window"`
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

	if len(cw.userMemories) > 0 {
		var parts []*genai.Part
		for _, m := range cw.userMemories {
			partText := fmt.Sprintf("As of %s, %s",
				humanize.Time(m.updatedAt),
				m.memory,
			)
			parts = append(parts, genai.NewPartFromText(partText))
		}
		contents = append(contents, &genai.Content{
			Role:  genai.RoleUser,
			Parts: parts,
		})
	}

	if len(cw.previousSessions) > 0 {
		var parts []*genai.Part
		for _, s := range cw.previousSessions {
			partText := fmt.Sprintf("Last Accessed: %s\nSummary: %s",
				humanize.Time(s.lastAccessed),
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

func (a *Agent) setContext(
	ctx context.Context,
	sc *sessionContext,
	cw []*genai.Content,
) error {
	for ag := range sc.GCCMap {
		if err := a.setgcc(ctx, sc, cw, ag); err != nil {
			return fmt.Errorf("failed to set context: %w", err)
		}
	}

	if err := a.setFlyContext(ctx, sc); err != nil {
		return fmt.Errorf("failed to set context: %w", err)
	}

	return nil
}

func (a *Agent) setgcc(
	ctx context.Context,
	sc *sessionContext,
	cw []*genai.Content,
	agent agentName,
) error {
	prof, err := a.getProfile(agent)
	if err != nil {
		return fmt.Errorf("failed to set gcc for %s: %w", agent, err)
	}

	if existing := sc.GCCMap[agent]; existing != nil {
		if _, err := a.gc.client.Caches.Delete(ctx, existing.resourceName, nil); err != nil {
			return fmt.Errorf("failed to set gcc for %s: %w", agent, err)
		}
	}

	expireTime := time.Now().Add(gccTTL30Mins)
	createReq := &genai.CreateCachedContentConfig{
		ExpireTime:        expireTime,
		Contents:          cw,
		SystemInstruction: prof.config.SystemInstruction,
		Tools:             prof.config.Tools,
	}

	res, err := a.gc.client.Caches.Create(ctx, string(prof.model), createReq)
	if err != nil {
		return fmt.Errorf("failed to set gcc for %s: %w", agent, err)
	}

	newGCC := &gcc{
		resourceName: res.Name,
		expiresAt:    res.ExpireTime,
	}
	sc.GCCMap[agent] = newGCC

	return nil
}

func (a *Agent) setFlyContext(ctx context.Context, sc *sessionContext) error {
	const method = "setContextCache"
	log := a.logger.With(
		slog.String("method", method),
		slog.String("user_id", sc.UserID.String()),
		slog.String("session_id", sc.SessionID.String()),
		slog.Time("created_at", sc.CreatedAt),
		slog.Time("updated_at", sc.UpdatedAt),
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

func (a *Agent) applygcc(
	agent agentName,
	sc *sessionContext,
	cw []*genai.Content,
	parts []*genai.Part,
) (*agentProfile, []*genai.Content, error) {
	prof, err := a.getProfile(agent)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to apply gcc for %s: %w", agent, err)
	}

	gcc := sc.GCCMap[agent]
	var fullContext []*genai.Content
	now := time.Now()

	if gcc != nil && now.Before(gcc.expiresAt) {
		prof.config.CachedContent = gcc.resourceName
		prof.config.SystemInstruction = nil
		prof.config.Tools = nil

		fullContext = []*genai.Content{
			{
				Role:  genai.RoleUser,
				Parts: parts,
			},
		}
	} else {
		fullContext = append(cw, &genai.Content{
			Role:  genai.RoleUser,
			Parts: parts,
		})
	}

	return prof, fullContext, nil
}

func (a *Agent) getContext(ctx context.Context) (*sessionContext, []*genai.Content, error) {
	userID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	sessionID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	sc, err := a.getFlyContext(ctx, userID, sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get context: %w", err)
	}

	if sc == nil {
		window, err := a.getDbContext(ctx, userID, sessionID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get context: %w", err)
		}

		sc = &sessionContext{
			UserID:    userID,
			SessionID: sessionID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			GCCMap:    make(map[agentName]*gcc),
			Window:    window,
		}
	}

	cw, err := a.buildContextWindow(ctx, sc.Window, searcherAgent)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get context: %w", err)
	}

	return sc, cw, nil
}

func (a *Agent) getFlyContext(ctx context.Context, userID, sessionID uuid.UUID) (*sessionContext, error) {
	const method = "getFlyContext"
	key := createContextCacheKey(userID, sessionID)
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
	).DebugContext(ctx, "context cache retrieved successfully")

	return &sc, nil
}

func (a *Agent) getDbContext(
	ctx context.Context,
	userID, sessionID uuid.UUID,
) (*contextWindow, error) {
	userUUID := pgtype.UUID{
		Bytes: userID,
		Valid: true,
	}

	var (
		memories     []userMemory
		sessions     []previousSession
		interactions []interaction
	)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		mem, err := a.store.Pg.GetMemoriesByUserID(ctx, database.GetMemoriesByUserIDParams{
			NumberOfMemories: int32(memories100),
			UserID:           userUUID,
		})
		if err != nil {
			return fmt.Errorf("failed to get context from db: %w", err)
		}

		local := make([]userMemory, 0, len(mem))
		for _, m := range mem {
			local = append(local, userMemory{
				updatedAt: m.UpdatedAt.Time,
				memory:    m.Memory,
			})
		}
		memories = local
		return nil
	})

	g.Go(func() error {
		prev, err := a.store.Pg.GetSessionsByUserID(ctx, database.GetSessionsByUserIDParams{
			NumberOfSessions: int32(sessions5),
			UserID:           userUUID,
		})
		if err != nil {
			return fmt.Errorf("failed to get context from db: %w", err)
		}

		local := make([]previousSession, 0, len(prev))
		for _, p := range prev {
			local = append(local, previousSession{
				lastAccessed: p.UpdatedAt.Time,
				summary:      p.Summary.String,
			})
		}
		sessions = local
		return nil
	})

	g.Go(func() error {
		messages, err := a.store.Pg.GetMessagesBySessionID(ctx, pgtype.UUID{
			Bytes: sessionID,
			Valid: true,
		})
		if err != nil {
			return fmt.Errorf("failed to get context from db: %w", err)
		}

		local := make([]interaction, 0)
		var current inputPrompt

		for _, m := range messages {
			switch m.Role {
			case "user":
				current.userInput = genai.NewPartFromText(m.Content)

			case "system":
				var responseMap map[string]any
				if err := json.Unmarshal(m.FunctionResponse, &responseMap); err != nil {
					return fmt.Errorf("failed to get context from db: %w", err)
				}
				current.functionResponse = genai.NewPartFromFunctionResponse(
					m.FunctionName.String,
					responseMap,
				)

			case "model":
				local = append(local, interaction{
					input:       current,
					modelOutput: genai.NewPartFromText(m.Content),
				})
				current = inputPrompt{}
			}
		}

		interactions = local
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &contextWindow{
		userMemories:     memories,
		previousSessions: sessions,
		history:          interactions,
	}, nil
}

func createContextCacheKey(userID uuid.UUID, sessionID uuid.UUID) string {
	return fmt.Sprintf("shaikh:user:%s:session:%s:context", userID.String(), sessionID.String())
}
