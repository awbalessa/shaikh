package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"log/slog"
	"strings"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/dom"
	svc "github.com/awbalessa/shaikh/backend/internal/svc/search"
	"github.com/dustin/go-humanize"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"
	"google.golang.org/genai"
)

type AskSvc struct {
	LLM       dom.LLM
	Cache     dom.Cache
	Publisher dom.Publisher
	SearchSvc *svc.SearchSvc
}

func (a *AskSvc) Ask(
	ctx context.Context,
	prompt string,
) iter.Seq2[string, error] {
	return iter.Seq2[string, error](func(yield func(string, error) bool) {
		const method = "Ask"
		start := time.Now()
		cc, win, err := a.getContext(ctx)
		if err != nil {
			yield("", err)
			return
		}

		userIn := genai.NewPartFromText(prompt)
		var fnOut *genai.Part
		var modelOut strings.Builder

		log := a.logger.With(
			slog.String("method", method),
			slog.String("userID", cc.UserID.String()),
			slog.String("sessionID", cc.SessionID.String()),
			slog.String("created_at", cc.CreatedAt.Format(time.RFC822)),
			slog.String("updated_at", cc.UpdatedAt.Format(time.RFC822)),
			slog.Int("turn", cc.Window.turns+1),
		)

		log.DebugContext(ctx, "asking agent...")
		gotFirst := false
		var ttft time.Duration

		for resp, err := range a.ask(ctx, win, userIn, &fnOut) {
			if err != nil {
				yield("", err)
				return
			}

			if !gotFirst {
				ttft = time.Since(start)
				gotFirst = true
			}
			modelOut.WriteString(resp)
			if !yieldOk(ctx, yield, resp) {
				return
			}
		}

		totalTime := time.Since(start)

		log.With(
			slog.String("ttft", ttft.String()),
			slog.String("total_time", totalTime.String()),
		).DebugContext(ctx, "response recieved: updating context...")

		modelOutPart := genai.NewPartFromText(modelOut.String())
		lastInteraction := &Interaction{
			Input: inputPrompt{
				FunctionResponse: fnOut,
				UserInput:        userIn,
			},
			ModelOutput: modelOutPart,
			TurnNumber:  cc.Window.turns + 1,
		}
		cc.Window.history = append(cc.Window.history, lastInteraction)

		if err = a.setContext(ctx, cc, lastInteraction); err != nil {
			yield("", err)
			return
		}

		log.DebugContext(ctx, "context updated: returning...")
	})
}

func (a *AskSvc) ask(
	ctx context.Context,
	win []*genai.Content,
	prompt *genai.Part,
	fnRes **genai.Part,
) iter.Seq2[string, error] {
	return iter.Seq2[string, error](func(yield func(string, error) bool) {
		const method = "ask"
		log := a.logger.With(
			"method", method,
		)
		prof, err := a.getProfile(searcherAgent)
		if err != nil {
			yield("", err)
			return
		}

		full := append(win, &genai.Content{
			Role:  genai.RoleUser,
			Parts: []*genai.Part{prompt},
		})

		for resp, err := range a.gc.client.Models.GenerateContentStream(
			ctx,
			string(prof.model),
			full,
			prof.config,
		) {
			if err != nil {
				log.With(
					"err", err,
				).ErrorContext(ctx, "failed to ask agent")
				yield("", fmt.Errorf("failed to ask agent: %w", err))
				return
			}

			if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
				continue
			}

			for _, part := range resp.Candidates[0].Content.Parts {
				switch {
				case part.FunctionCall != nil:
					log.With(
						"name", part.FunctionCall.Name,
					).DebugContext(ctx, "agent called function")
					fnResponse, err := a.handleFunctionCall(
						ctx,
						win,
						prompt,
						part.FunctionCall,
						yield,
					)
					if err != nil {
						yield("", err)
						return
					}

					*fnRes = fnResponse
					return

				case part.Text != "":
					if !yieldOk(ctx, yield, part.Text) {
						return
					}
				}
			}
		}
	})
}

func (a *AskSvc) handleFunctionCall(
	ctx context.Context,
	win []*genai.Content,
	prompt *genai.Part,
	fnCall *genai.FunctionCall,
	yield func(string, error) bool,
) (*genai.Part, error) {
	const method = "handleFunctionCall"
	log := a.logger.With(
		"method", method,
	)

	prof, err := a.getProfile(generatorAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to handle function %s: %w", fnCall.Name, err)
	}

	fn, err := a.getFunction(dom.FunctionName(fnCall.Name))
	if err != nil {
		return nil, fmt.Errorf("failed to handle function %s: %w", fnCall.Name, err)
	}

	results, err := fn.call(ctx, fnCall.Args)
	if err != nil {
		log.With(
			"name", fnCall.Name,
			"args", fnCall.Args,
		).ErrorContext(ctx, "failed to handle function")
		return nil, fmt.Errorf("failed to handle function %s: %w", fnCall.Name, err)
	}

	fnPart := genai.NewPartFromFunctionResponse(string(fn.name()), results)

	full := append(win, &genai.Content{
		Role:  genai.RoleUser,
		Parts: []*genai.Part{fnPart, prompt},
	})

	for resp, err := range a.gc.client.Models.GenerateContentStream(
		ctx,
		string(prof.model),
		full,
		prof.config,
	) {
		if err != nil {
			log.With(
				"name", fnCall.Name,
				"args", fnCall.Args,
			).ErrorContext(ctx, "failed to handle function")
			return nil, fmt.Errorf("failed to handle function %s: %w", fnCall.Name, err)
		}

		if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
			continue
		}

		for _, part := range resp.Candidates[0].Content.Parts {
			if !yieldOk(ctx, yield, part.Text) {
				return fnPart, nil
			}
		}
	}

	return fnPart, nil
}

func yieldOk(ctx context.Context, yield func(string, error) bool, s string) bool {
	if ctx.Err() != nil {
		yield("", ctx.Err())
		return false
	}
	if !yield(s, nil) {
		return false
	}
	return true
}

const (
	contextCacheTTL6Hrs  time.Duration = 6 * time.Hour
	contextCacheTTL12Hrs time.Duration = 12 * time.Hour
	memories100          int           = 100
	sessions5            int           = 5
	tokenLimit           int           = 200_000
)

type LLMFunctionResponseDTO struct {
	Name    string         `json:"name"`
	Content map[string]any `json:"content"`
}

type InputPrompt struct {
	Text             string                  `json:"text"`
	FunctionResponse *LLMFunctionResponseDTO `json:"function_response"`
}

type LLMFunctionCallDTO struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args"`
}

type ModelOutput struct {
	Text         string              `json:"text"`
	FunctionCall *LLMFunctionCallDTO `json:"function_call"`
}

type Interaction struct {
	Input      InputPrompt `json:"input_prompt"`
	Output     ModelOutput `json:"model_output"`
	TurnNumber int32       `json:"turn_number"`
}

type SyncPayload struct {
	UserID      uuid.UUID    `json:"user_id"`
	SessionID   uuid.UUID    `json:"session_id"`
	Interaction *Interaction `json:"interaction"`
}

type ContextWindow struct {
	UserMemories     []dom.Memory  `json:"user_memories"`
	PreviousSessions []dom.Session `json:"previous_sessions"`
	History          []Interaction `json:"history"`
	Turns            int           `json:"turns"`
}

type ContextCache struct {
	UserID    uuid.UUID      `json:"user_id"`
	SessionID uuid.UUID      `json:"session_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	Window    *ContextWindow `json:"context_window"`
}

func (a *Agent) getContext(ctx context.Context) (*contextCache, []*genai.Content, error) {
	const method = "getContext"
	userID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	sessionID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	log := a.logger.With(
		slog.String("method", method),
	)

	now := time.Now()
	sc, err := a.getContextCache(ctx, userID, sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get context: %w", err)
	}

	if sc == nil {
		window, err := a.getDbContext(ctx, userID, sessionID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get context: %w", err)
		}

		sc = &contextCache{
			UserID:    userID,
			SessionID: sessionID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Window:    window,
		}
	}

	cw, err := a.buildContextWindow(ctx, sc.Window, searcherAgent)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get context: %w", err)
	}

	log.With(
		slog.String("duration", time.Since(now).String()),
	).DebugContext(ctx, "retrieved context successfully")

	return sc, cw, nil
}

func (a *Agent) getContextCache(ctx context.Context, userID, sessionID uuid.UUID) (*contextCache, error) {
	const method = "getContextCache"
	key := createContextCacheKey(userID, sessionID)
	log := a.logger.With(
		slog.String("method", method),
	)

	now := time.Now()
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

	var sc contextCache
	if err = json.Unmarshal(bytes, &sc); err != nil {
		log.With(
			"err", err,
		).ErrorContext(ctx, "failed to get context cache")
		return nil, fmt.Errorf("failed to get context cache: %w", err)
	}

	log.With(
		slog.String("duration", time.Since(now).String()),
	).DebugContext(ctx, "retrieved context cache successfully")

	return &sc, nil
}

func (a *Agent) getDbContext(
	ctx context.Context,
	userID, sessionID uuid.UUID,
) (*contextWindow, error) {
	const method = "getDbContext"
	userUUID := pgtype.UUID{
		Bytes: userID,
		Valid: true,
	}

	var (
		memories     []*userMemory
		sessions     []*previousSession
		interactions []*Interaction
	)

	log := a.logger.With(
		slog.String("method", method),
	)

	now := time.Now()
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		mem, err := a.store.Pg.GetMemoriesByUserID(ctx, database.GetMemoriesByUserIDParams{
			NumberOfMemories: int32(memories100),
			UserID:           userUUID,
		})
		if err != nil {
			log.With(
				"err", err,
			).ErrorContext(ctx, "failed to get context from db")
			return fmt.Errorf("failed to get context from db: %w", err)
		}

		local := make([]*userMemory, 0, len(mem))
		for _, m := range mem {
			local = append(local, &userMemory{
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
			log.With(
				"err", err,
			).ErrorContext(ctx, "failed to get context from db")
			return fmt.Errorf("failed to get context from db: %w", err)
		}

		local := make([]*previousSession, 0, len(prev))
		for _, p := range prev {
			local = append(local, &previousSession{
				lastAccessed: p.UpdatedAt.Time,
				summary:      p.Summary.String,
			})
		}
		sessions = local
		return nil
	})

	g.Go(func() error {
		messages, err := a.store.Pg.GetMessagesBySessionIDOrdered(ctx, pgtype.UUID{
			Bytes: sessionID,
			Valid: true,
		})
		if err != nil {
			log.With(
				"err", err,
			).ErrorContext(ctx, "failed to get context from db")
			return fmt.Errorf("failed to get context from db: %w", err)
		}

		local := make([]*Interaction, 0)
		var current inputPrompt

		for _, m := range messages {
			switch m.Role {
			case "user":
				current.UserInput = genai.NewPartFromText(m.Content)

			case "function":
				var responseMap map[string]any
				if err := json.Unmarshal([]byte(m.Content), &responseMap); err != nil {
					log.With(
						"err", err,
					).ErrorContext(ctx, "failed to get context from db")
					return fmt.Errorf("failed to get context from db: %w", err)
				}
				current.FunctionResponse = genai.NewPartFromFunctionResponse(
					m.FunctionName.String,
					responseMap,
				)

			case "model":
				local = append(local, &Interaction{
					Input:       current,
					ModelOutput: genai.NewPartFromText(m.Content),
					TurnNumber:  int(m.Turn),
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

	turns := 0
	if len(interactions) > 0 {
		last := interactions[len(interactions)-1]
		turns = last.TurnNumber
	}

	log.With(
		slog.String("duration", time.Since(now).String()),
	).DebugContext(ctx, "retrieved db context successfully")

	return &contextWindow{
		userMemories:     memories,
		previousSessions: sessions,
		history:          interactions,
		turns:            turns,
	}, nil
}

func (a *Agent) setContext(
	ctx context.Context,
	cc *contextCache,
	lastInteraction *Interaction,
) error {
	if err := a.setContextCache(ctx, cc.UserID, cc.SessionID, cc.Window); err != nil {
		return fmt.Errorf("failed to set context: %w", err)
	}

	if err := a.sendContextUpdate(ctx, cc.UserID, cc.SessionID, lastInteraction); err != nil {
		return fmt.Errorf("failed to set context: %w", err)
	}
	return nil
}

func (a *Agent) setContextCache(
	ctx context.Context,
	userID, sessionID uuid.UUID,
	win *contextWindow,
) error {
	const method = "setContextCache"
	log := a.logger.With(
		slog.String("method", method),
	)

	now := time.Now()
	win.turns += 1
	bytes, err := json.Marshal(&contextCache{
		UserID:    userID,
		SessionID: sessionID,
		CreatedAt: now,
		UpdatedAt: now,
		Window:    win,
	})
	if err != nil {
		log.With(
			"err", err,
		).ErrorContext(ctx, "failed to set context cache")
		return fmt.Errorf("failed to set context cache: %w", err)
	}

	key := createContextCacheKey(userID, sessionID)

	if err = a.store.Fly.Set(ctx, key, bytes, contextCacheTTL6Hrs); err != nil {
		log.With(
			"err", err,
		).ErrorContext(ctx, "failed to set context cache")
		return fmt.Errorf("failed to set context cache: %w", err)
	}

	log.With(
		slog.String("duration", time.Since(now).String()),
	).DebugContext(ctx, "set context cache successfully")

	return nil
}

func (a *Agent) sendContextUpdate(
	ctx context.Context,
	userID, sessionID uuid.UUID,
	interaction *Interaction,
) error {
	const method = "sendContextUpdate"
	log := a.logger.With(
		slog.String("method", method),
	)

	load := &SyncPayload{
		UserID:      userID,
		SessionID:   sessionID,
		Interaction: interaction,
	}

	start := time.Now()
	data, err := json.Marshal(load)
	if err != nil {
		log.With(
			"err", err,
		).ErrorContext(ctx, "failed to send context update")
		return fmt.Errorf("failed to send context update: %w", err)
	}

	msg := &nats.Msg{
		Subject: SyncerSubject,
		Data:    data,
	}

	msg.Header = nats.Header{}
	msg.Header.Set("Nats-Msg-Id", fmt.Sprintf("%s-%d", sessionID.String(), interaction.TurnNumber))

	ack, err := a.js.PublishMsg(ctx, msg)
	if err != nil {
		log.With(
			"err", err,
		).ErrorContext(ctx, "failed to send context update")
		return fmt.Errorf("failed to send context update: %w", err)
	}

	if ack == nil {
		log.With(
			"ack", ack,
		).ErrorContext(ctx, "unexpected publish ack")
		return fmt.Errorf("unexpected publish ack: %+v", ack)
	}

	log.With(
		slog.String("duration", time.Since(start).String()),
	).DebugContext(ctx, "sent context update successfully")

	return nil
}

func (a *Agent) buildContextWindow(
	ctx context.Context,
	cw *contextWindow,
	agent agentName,
) ([]*genai.Content, error) {
	const method = "buildContextWindow"
	log := a.logger.With(
		"method", method,
	)

	prof, err := a.getProfile(agent)
	if err != nil {
		log.With(
			"err", err,
		).ErrorContext(ctx, "failed to build context window")
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
		if inter.Input.FunctionResponse != nil {
			userParts = append(userParts, inter.Input.FunctionResponse)
		}
		if inter.Input.UserInput != nil {
			userParts = append(userParts, inter.Input.UserInput)
		}
		if len(userParts) > 0 {
			historyContents = append(historyContents, &genai.Content{
				Role:  genai.RoleUser,
				Parts: userParts,
			})
		}
		if inter.ModelOutput != nil {
			historyContents = append(historyContents, &genai.Content{
				Role:  genai.RoleModel,
				Parts: []*genai.Part{inter.ModelOutput},
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
			log.With(
				"err", err,
			).ErrorContext(ctx, "failed to build context window")
			return nil, fmt.Errorf("failed to build context window: %w", err)
		}

		if tokResp.TotalTokens < int32(tokenLimit) {
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

	log.DebugContext(ctx, "built context window successfully")
	return contents, nil
}

func createContextCacheKey(userID uuid.UUID, sessionID uuid.UUID) string {
	return fmt.Sprintf("user:%s:session:%s:context", userID.String(), sessionID.String())
}
