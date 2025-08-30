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
	Functions  dom.LLMFunctions
	CtxManager *ContextManager
	SearchSvc  *SearchSvc
	Logger     *slog.Logger
}

func (a *AskSvc) Ask(
	ctx context.Context,
	prompt string,
) iter.Seq2[string, error] {
	return iter.Seq2[string, error](func(yield func(string, error) bool) {
		cc, err := a.CtxManager.GetContext(ctx)
		if err != nil {
			yield("", err)
			return
		}

		cw := toDomainContextWindow(cc.Window)

		win, err := a.Agent.BuildContextWindow(ctx, dom.Caller, cw, time.Now())
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
) [2]*InferenceDTO {
	var infs [2]*InferenceDTO

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

		infs = [2]*InferenceDTO{
			&InferenceDTO{
				Input: &InputPromptDTO{
					Text: prompt,
				},
				Output: &ModelOutputDTO{
					FunctionCall: &LLMFunctionCallDTO{
						Name: results[0].Output.FunctionCall.Name,
						Args: results[0].Output.FunctionCall.Args,
					},
				},
				InputTokens:  results[0].Usage.InputTokens,
				OutputTokens: results[0].Usage.OutputTokens,
				Model:        dom.AgentToModel[dom.Caller],
			},
			&InferenceDTO{
				Input: &InputPromptDTO{
					FunctionResponse: &LLMFunctionResponseDTO{
						Name:    fnResp.Name,
						Content: fnResp.Content,
					},
				},
				Output: &ModelOutputDTO{
					Text: results[1].Output.Text,
				},
				InputTokens:  results[1].Usage.InputTokens,
				OutputTokens: results[1].Usage.OutputTokens,
				Model:        dom.AgentToModel[dom.Generator],
			},
		}

		return infs
	}

	infs = [2]*InferenceDTO{
		&InferenceDTO{
			Input: &InputPromptDTO{
				Text: prompt,
			},
			Output: &ModelOutputDTO{
				Text: results[0].Output.Text,
			},
			InputTokens:  results[0].Usage.InputTokens,
			OutputTokens: results[0].Usage.OutputTokens,
			Model:        dom.AgentToModel[dom.Caller],
		},
		nil,
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
	FunctionResponse *LLMFunctionResponseDTO `json:"function_response,omitempty"`
}

type LLMFunctionCallDTO struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args"`
}

type ModelOutputDTO struct {
	Text         string              `json:"text,omitempty"`
	FunctionCall *LLMFunctionCallDTO `json:"function_call,omitempty"`
}

type InferenceDTO struct {
	Input        *InputPromptDTO `json:"input"`
	Output       *ModelOutputDTO `json:"output"`
	InputTokens  int32
	OutputTokens int32
	Model        dom.LargeLanguageModel
}

type InteractionDTO struct {
	Inferences [2]*InferenceDTO `json:"inferences"`
	TurnNumber int32            `json:"turn_number"`
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
		var inter dom.Interaction
		if len(i.Inferences) > 1 {
			inter = dom.Interaction{
				Inferences: [2]*dom.Inference{
					&dom.Inference{
						Input: &dom.InputPrompt{
							Text: i.Inferences[0].Input.Text,
						},
						Output: &dom.ModelOutput{
							FunctionCall: (*dom.LLMFunctionCall)(i.Inferences[0].Output.FunctionCall),
						},
						InputTokens:  i.Inferences[0].InputTokens,
						OutputTokens: i.Inferences[0].OutputTokens,
						Model:        i.Inferences[0].Model,
					},
					&dom.Inference{
						Input: &dom.InputPrompt{
							FunctionResponse: (*dom.LLMFunctionResponse)(i.Inferences[1].Input.FunctionResponse),
						},
						Output: &dom.ModelOutput{
							Text: i.Inferences[1].Output.Text,
						},
						InputTokens:  i.Inferences[1].InputTokens,
						OutputTokens: i.Inferences[1].OutputTokens,
						Model:        i.Inferences[1].Model,
					},
				},
			}
		} else {
			inter = dom.Interaction{
				Inferences: [2]*dom.Inference{
					&dom.Inference{
						Input: &dom.InputPrompt{
							Text: i.Inferences[0].Input.Text,
						},
						Output: &dom.ModelOutput{
							Text: i.Inferences[0].Output.Text,
						},
						InputTokens:  i.Inferences[0].InputTokens,
						OutputTokens: i.Inferences[0].OutputTokens,
						Model:        i.Inferences[0].Model,
					},
					nil,
				},
			}
		}

		history = append(history, inter)
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

func (c *ContextManager) GetContext(ctx context.Context) (*ContextCacheDTO, error) {
	const method = "GetContext"
	userID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	sessionID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	log := c.Logger.With(
		slog.String("method", method),
	)

	now := time.Now()
	cc, err := c.getContextCache(ctx, userID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get context: %w", err)
	}

	if cc == nil {
		window, err := c.getDbContext(ctx, userID, sessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to get context: %w", err)
		}

		cc = &ContextCacheDTO{
			UserID:    userID,
			SessionID: sessionID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Window:    window,
		}
	}

	log.With(
		slog.String("duration", time.Since(now).String()),
	).DebugContext(ctx, "retrieved context successfully")

	return cc, nil
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
		mem, err := c.MemoryRepo.GetMemoriesByUserID(ctx, userID, 50)
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
		msgs, err := c.MessageRepo.GetMessagesBySessionIDOrdered(ctx, sessionID)
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
		inf1         InferenceDTO = InferenceDTO{}
		inf2         InferenceDTO = InferenceDTO{}
		has2infs     bool         = false
	)
	for _, m := range messages {
		meta := m.Meta()
		role := m.Role()
		switch role {
		case dom.UserRole:
			inf1 = InferenceDTO{
				Input: &InputPromptDTO{
					Text: *meta.Content,
				},
				InputTokens: *meta.TotalInputTokens,
			}
		case dom.FunctionRole:
			has2infs = true
			call, err := fromJsonRawMessage(meta.FunctionCall)
			if err != nil {
				return nil, err
			}
			resp, err := fromJsonRawMessage(meta.FunctionResponse)
			if err != nil {
				return nil, err
			}
			inf1.Output.FunctionCall = &LLMFunctionCallDTO{
				Name: *meta.FunctionName,
				Args: call,
			}
			inf1.OutputTokens = *meta.TotalOutputTokens
			inf2 = InferenceDTO{
				Input: &InputPromptDTO{
					FunctionResponse: &LLMFunctionResponseDTO{
						Name:    *meta.FunctionName,
						Content: resp,
					},
				},
				InputTokens: *meta.TotalInputTokens,
			}

		case dom.ModelRole:
			if !has2infs {
				inf1.Output.Text = *meta.Content
				inf1.OutputTokens = *meta.TotalOutputTokens
				inf1.Model = meta.Model
			} else {
				inf2.Output.Text = *meta.Content
				inf2.OutputTokens = *meta.TotalOutputTokens
				inf2.Model = meta.Model
			}

			has2infs = false
			interactions = append(interactions, InteractionDTO{
				Inferences: [2]*InferenceDTO{&inf1, &inf2},
				TurnNumber: meta.Turn,
			})
			inf1 = InferenceDTO{}
			inf2 = InferenceDTO{}
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
	infs [2]*InferenceDTO,
	turn int32,
) *InteractionDTO {
	dto := &InteractionDTO{
		TurnNumber: turn,
	}

	if len(infs) > 1 {
		dto.Inferences[0] = &InferenceDTO{
			Input: &InputPromptDTO{
				Text: infs[0].Input.Text,
			},
			Output: &ModelOutputDTO{
				FunctionCall: infs[0].Output.FunctionCall,
			},
			InputTokens:  infs[0].InputTokens,
			OutputTokens: infs[0].OutputTokens,
			Model:        infs[0].Model,
		}

		dto.Inferences[1] = &InferenceDTO{
			Input: &InputPromptDTO{
				FunctionResponse: infs[1].Input.FunctionResponse,
			},
			Output: &ModelOutputDTO{
				Text: infs[1].Output.Text,
			},
			InputTokens:  infs[1].InputTokens,
			OutputTokens: infs[1].OutputTokens,
			Model:        infs[1].Model,
		}

		return dto
	}

	dto.Inferences = [2]*InferenceDTO{
		&InferenceDTO{
			Input: &InputPromptDTO{
				Text: infs[0].Input.Text,
			},
			Output: &ModelOutputDTO{
				Text: infs[0].Output.Text,
			},
			InputTokens:  infs[0].InputTokens,
			OutputTokens: infs[0].OutputTokens,
			Model:        infs[0].Model,
		},
		nil,
	}
	return dto
}

func toJsonRawMessage(m map[string]any) (json.RawMessage, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

func fromJsonRawMessage(m json.RawMessage) (map[string]any, error) {
	final := make(map[string]any)
	if err := json.Unmarshal(m, &final); err != nil {
		return nil, err
	}

	return final, nil
}
