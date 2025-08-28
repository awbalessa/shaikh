package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"log/slog"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/dom"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type AskSvc struct {
	Agent      dom.Agent
	Functions  map[dom.LLMFunctionName]dom.LLMFunction
	CtxManager *ContextManager
	SearchSvc  *SearchSvc
	Logger     *slog.Logger
}

type Inference struct {
	Input        *dom.InputPrompt
	Output       *dom.ModelOutput
	InputTokens  int32
	OutputTokens int32
}

func (a *AskSvc) Ask(
	ctx context.Context,
	prompt string,
) iter.Seq2[string, error] {
	return iter.Seq2[string, error](func(yield func(string, error) bool) {
		cc, win, err := a.GetContext(ctx, dom.Caller)
		if err != nil {
			yield("", err)
			return
		}

		infs := a.ask(ctx, prompt, win, yield)
		inter := toInteractionDTO(prompt, infs, cc.Window.Turns+1)

		if cc.Window == nil {
			cc.Window = &ContextWindowDTO{}
		}
		cc.Window.History = append(cc.Window.History, *inter)

		if err = a.CtxManager.SetContext(ctx, cc, inter); err != nil {
			yield("", err)
			return
		}
	})
}

func (a *AskSvc) ask(
	ctx context.Context,
	prompt string,
	win []*dom.LLMContent,
	yield func(string, error) bool,
) [2]*Inference {
	var infs [2]*Inference

	win = append(win, &dom.LLMContent{
		Role: dom.LLMUserRole,
		Parts: []*dom.LLMPart{
			&dom.LLMPart{Text: prompt},
		},
	})

	var fnResp *dom.LLMFunctionResponse

	results := make([]*dom.LLMGenResult, 0, 2)

	syield := func(p *dom.LLMPart, err error) bool {
		if err != nil {
			return yield("", fmt.Errorf("ask: %w", err))
		}
		if p == nil {
			return true
		}

		if p.Text != "" {
			return yield(p.Text, nil)
		}
		if p.FunctionCall != nil {
			fnResp, err = a.handleFn(ctx, *p.FunctionCall)
			if err != nil {
				return yield("", fmt.Errorf("ask: %w", err))
			}
			return false
		}

		return true
	}

	results = append(results,
		a.Agent.GenerateWithYield(ctx, dom.Caller, win, syield),
	)

	infs[0] = &Inference{
		Input: &dom.InputPrompt{
			Text: prompt,
		},
		Output:       results[0].Output,
		InputTokens:  results[0].Usage.InputTokens,
		OutputTokens: results[0].Usage.OutputTokens,
	}

	if fnResp != nil {
		win = append(win, &dom.LLMContent{
			Role: dom.LLMModelRole,
			Parts: []*dom.LLMPart{
				&dom.LLMPart{FunctionCall: results[0].Output.FunctionCall},
			},
		})
		win = append(win, &dom.LLMContent{
			Role: dom.LLMUserRole,
			Parts: []*dom.LLMPart{
				&dom.LLMPart{FunctionResponse: fnResp},
			},
		})

		results = append(results,
			a.Agent.GenerateWithYield(ctx, dom.Generator, win, syield),
		)

		infs[1] = &Inference{
			Input: &dom.InputPrompt{
				Text:             prompt,
				FunctionResponse: fnResp,
			},
			Output:       results[1].Output,
			InputTokens:  results[1].Usage.InputTokens,
			OutputTokens: results[1].Usage.OutputTokens,
		}
	}

	return infs
}

func (a *AskSvc) handleFn(
	ctx context.Context,
	fn dom.LLMFunctionCall,
) (*dom.LLMFunctionResponse, error) {
	function, ok := a.Functions[dom.LLMFunctionName(fn.Name)]
	if !ok {
		return nil, fmt.Errorf("function %s does not exist", fn.Name)
	}

	res, err := function.Call(ctx, fn.Args)
	if err != nil {
		return nil, fmt.Errorf("handleFn: %w", err)
	}

	return &dom.LLMFunctionResponse{
		Name:    fn.Name,
		Content: res,
	}, nil
}

type LLMFunctionResponseDTO struct {
	Name    string         `json:"name"`
	Content map[string]any `json:"content"`
}

type InputPromptDTO struct {
	Text             string                  `json:"text"`
	FunctionResponse *LLMFunctionResponseDTO `json:"function_response"`
}

type LLMFunctionCallDTO struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args"`
}

type ModelOutputDTO struct {
	Text         string              `json:"text"`
	FunctionCall *LLMFunctionCallDTO `json:"function_call"`
}

type InteractionDTO struct {
	Input      InputPromptDTO `json:"input_prompt"`
	Output     ModelOutputDTO `json:"model_output"`
	TurnNumber int32          `json:"turn_number"`
	Usage      []dom.TokenUsage
}

type SyncPayloadDTO struct {
	UserID         uuid.UUID       `json:"user_id"`
	SessionID      uuid.UUID       `json:"session_id"`
	InteractionDTO *InteractionDTO `json:"interaction"`
}

type ContextWindowDTO struct {
	UserMemories     []dom.Memory     `json:"user_memories"`
	PreviousSessions []dom.Session    `json:"previous_sessions"`
	History          []InteractionDTO `json:"history"`
	Turns            int32            `json:"turns"`
}

func toDomainContextWindow(cw *ContextWindowDTO) *dom.ContextWindow {
	if cw == nil {
		return nil
	}

	history := make([]dom.Interaction, 0, len(cw.History))
	for _, i := range cw.History {
		history = append(history, dom.Interaction{
			Input: dom.InputPrompt{
				Text: i.Input.Text,
				FunctionResponse: &dom.LLMFunctionResponse{
					Name:    i.Input.FunctionResponse.Name,
					Content: i.Input.FunctionResponse.Content,
				},
			},
			Output: dom.ModelOutput{
				Text: i.Output.Text,
				FunctionCall: &dom.LLMFunctionCall{
					Name: i.Output.FunctionCall.Name,
					Args: i.Output.FunctionCall.Args,
				},
			},
			TurnNumber: i.TurnNumber,
		})
	}

	return &dom.ContextWindow{
		UserMemories:     cw.UserMemories,
		PreviousSessions: cw.PreviousSessions,
		History:          history,
		Turns:            cw.Turns,
	}
}

type ContextCacheDTO struct {
	UserID    uuid.UUID         `json:"user_id"`
	SessionID uuid.UUID         `json:"session_id"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Window    *ContextWindowDTO `json:"context_window"`
}

type ContextManager struct {
	dom.Cache
	dom.ContextRepo
	dom.Publisher
	Logger *slog.Logger
}

func (a *AskSvc) GetContext(ctx context.Context, name dom.AgentName) (*ContextCacheDTO, []*dom.LLMContent, error) {
	const method = "GetContext"
	userID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	sessionID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	log := a.Logger.With(
		slog.String("method", method),
	)

	now := time.Now()
	cc, err := a.CtxManager.getContextCache(ctx, userID, sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get context: %w", err)
	}

	if cc == nil {
		window, err := a.CtxManager.getDbContext(ctx, userID, sessionID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get context: %w", err)
		}

		cc = &ContextCacheDTO{
			UserID:    userID,
			SessionID: sessionID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Window:    window,
		}
	}

	cw := toDomainContextWindow(cc.Window)

	win, err := a.Agent.BuildContextWindow(ctx, name, cw)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get context: %w", err)
	}

	log.With(
		slog.String("duration", time.Since(now).String()),
	).DebugContext(ctx, "retrieved context successfully")

	return cc, win, nil
}

func (c *ContextManager) getContextCache(ctx context.Context, userID, sessionID uuid.UUID) (*ContextCacheDTO, error) {
	const method = "getContextCache"
	log := c.Logger.With(slog.String("method", method))

	now := time.Now()
	key := dom.CreateContextCacheKey(userID, sessionID)

	bytes, err := c.Cache.Get(ctx, key)
	if err != nil {
		log.With("err", err).ErrorContext(ctx, "failed to get context cache")
		return nil, fmt.Errorf("getContextCache: %w", err)
	}
	if bytes == nil {
		log.WarnContext(ctx, "context cache miss: returning nil")
		return nil, nil
	}

	var cc ContextCacheDTO
	if err := json.Unmarshal(bytes, &cc); err != nil {
		log.With("err", err).ErrorContext(ctx, "failed to unmarshal context cache")
		return nil, fmt.Errorf("getContextCache: %w", err)
	}

	log.With(
		slog.String("duration", time.Since(now).String()),
	).DebugContext(ctx, "retrieved context cache successfully")

	return &cc, nil
}

func (c *ContextManager) getDbContext(
	ctx context.Context,
	userID, sessionID uuid.UUID,
) (*ContextWindowDTO, error) {
	const method = "getDbContext"
	log := c.Logger.With(slog.String("method", method))
	start := time.Now()

	g, ctx := errgroup.WithContext(ctx)

	var (
		memories []dom.Memory
		sessions []dom.Session
		messages []dom.Message
	)

	g.Go(func() error {
		mem, err := c.MemoryRepo.GetByUserID(ctx, userID, 50)
		if err != nil {
			log.With("err", err).ErrorContext(ctx, "failed to fetch memories")
			return fmt.Errorf("getDbContext: %w", err)
		}
		memories = mem
		return nil
	})

	g.Go(func() error {
		prev, err := c.SessionRepo.GetSessionsByUserID(ctx, userID, 5)
		if err != nil {
			log.With("err", err).ErrorContext(ctx, "failed to fetch sessions")
			return fmt.Errorf("getDbContext: %w", err)
		}
		sessions = prev
		return nil
	})

	g.Go(func() error {
		msgs, err := c.MessageRepo.GetBySessionIDOrdered(ctx, sessionID)
		if err != nil {
			log.With("err", err).ErrorContext(ctx, "failed to fetch messages")
			return fmt.Errorf("getDbContext: %w", err)
		}
		messages = msgs
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	var (
		interactions []InteractionDTO
		current      InteractionDTO = InteractionDTO{}
	)
	for _, m := range messages {
		switch m.Role() {
		case dom.UserRole:
			current.Input.Text = *m.Meta().Content
		case dom.FunctionRole:
			var call map[string]any
			if err := json.Unmarshal(m.Meta().FunctionCall, &call); err != nil {
				log.With("err", err).ErrorContext(ctx, "failed to decode function response")
				return nil, fmt.Errorf("getDbContext: %w", err)
			}
			current.Output.FunctionCall = &LLMFunctionCallDTO{
				Name: *m.Meta().FnName,
				Args: call,
			}

			var resp map[string]any
			if err := json.Unmarshal(m.Meta().FunctionResponse, &resp); err != nil {
				log.With("err", err).ErrorContext(ctx, "failed to decode function response")
				return nil, fmt.Errorf("getDbContext: %w", err)
			}
			current.Input.FunctionResponse = &LLMFunctionResponseDTO{
				Name:    *m.Meta().FnName,
				Content: resp,
			}
		case dom.ModelRole:
			current.Output.Text = *m.Meta().Content
			current.TurnNumber = m.Meta().Turn
			interactions = append(interactions, current)
			current = InteractionDTO{}
		}
	}

	var turns int32 = 0
	if len(interactions) > 0 {
		turns = interactions[len(interactions)-1].TurnNumber
	}

	log.With(slog.String("duration", time.Since(start).String())).
		DebugContext(ctx, "retrieved db context successfully")

	return &ContextWindowDTO{
		UserMemories:     memories,
		PreviousSessions: sessions,
		History:          interactions,
		Turns:            turns,
	}, nil
}

func (c *ContextManager) SetContext(
	ctx context.Context,
	cc *ContextCacheDTO,
	interaction *InteractionDTO,
) error {
	if err := c.setContextCache(ctx, cc.UserID, cc.SessionID, cc.Window); err != nil {
		return fmt.Errorf("failed to set context: %w", err)
	}

	if err := c.sendContextUpdate(ctx, cc.UserID, cc.SessionID, interaction); err != nil {
		return fmt.Errorf("failed to set context: %w", err)
	}
	return nil
}

func (c *ContextManager) setContextCache(
	ctx context.Context,
	userID, sessionID uuid.UUID,
	win *ContextWindowDTO,
) error {
	const method = "setContextCache"
	log := c.Logger.With(
		slog.String("method", method),
	)

	now := time.Now()
	win.Turns += 1
	bytes, err := json.Marshal(&ContextCacheDTO{
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

	key := dom.CreateContextCacheKey(userID, sessionID)

	if err = c.Cache.Set(ctx, key, bytes, dom.ContextCacheTTL6Hrs); err != nil {
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

func (c *ContextManager) sendContextUpdate(
	ctx context.Context,
	userID, sessionID uuid.UUID,
	interaction *InteractionDTO,
) error {
	const method = "sendContextUpdate"
	log := c.Logger.With(
		slog.String("method", method),
	)

	load := &SyncPayloadDTO{
		UserID:         userID,
		SessionID:      sessionID,
		InteractionDTO: interaction,
	}

	start := time.Now()
	data, err := json.Marshal(load)
	if err != nil {
		log.With(
			"err", err,
		).ErrorContext(ctx, "failed to send context update")
		return fmt.Errorf("failed to send context update: %w", err)
	}

	_, err = c.Publisher.Publish(ctx, SyncerSubject, data, &dom.PubOptions{
		MsgID: fmt.Sprintf("%s:%s:%d", userID, sessionID, interaction.TurnNumber),
	})
	if err != nil {
		log.With(
			"err", err,
		).ErrorContext(ctx, "failed to send context update")
		return fmt.Errorf("failed to send context update: %w", err)
	}

	// if ack.Stream != AskSvcStream {
	// 	log.With(
	// 		"stream", ack.Stream,
	// 	).ErrorContext(ctx, "published to unexpected stream")
	// 	return fmt.Errorf("published to unexpected stream: %s", ack.Stream)
	// }

	log.With(
		slog.String("duration", time.Since(start).String()),
	).DebugContext(ctx, "sent context update successfully")

	return nil
}

func toInteractionDTO(
	prompt string,
	infs [2]*Inference,
	turn int32,
) *InteractionDTO {
	dto := &InteractionDTO{
		TurnNumber: turn,
	}

	if infs[0] != nil {
		dto.Usage = append(dto.Usage, dom.TokenUsage{
			InputTokens:  infs[0].InputTokens,
			OutputTokens: infs[0].OutputTokens,
		})
	}

	if infs[1] != nil {
		dto.Input = InputPromptDTO{
			Text: prompt,
			FunctionResponse: &LLMFunctionResponseDTO{
				Name:    infs[1].Input.FunctionResponse.Name,
				Content: infs[1].Input.FunctionResponse.Content,
			},
		}
		dto.Output = ModelOutputDTO{
			Text: infs[1].Output.Text,
			FunctionCall: &LLMFunctionCallDTO{
				Name: infs[0].Output.FunctionCall.Name,
				Args: infs[0].Output.FunctionCall.Args,
			},
		}
		dto.Usage = append(dto.Usage, dom.TokenUsage{
			InputTokens:  infs[1].InputTokens,
			OutputTokens: infs[1].OutputTokens,
		})
	} else if infs[0] != nil {
		dto.Input = InputPromptDTO{
			Text: prompt,
		}
		dto.Output = ModelOutputDTO{
			Text: infs[0].Output.Text,
		}
	}

	return dto
}
