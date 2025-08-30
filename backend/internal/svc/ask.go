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

func BuildAskSvc(ag dom.Agent, fns dom.LLMFunctions, ctx *ContextManager, se *SearchSvc) *AskSvc {
	log := slog.Default().With(
		"service", "ask",
	)

	return &AskSvc{
		Agent:      ag,
		Functions:  fns,
		CtxManager: ctx,
		SearchSvc:  se,
		Logger:     log,
	}
}

func (a *AskSvc) Ask(
	ctx context.Context,
	prompt string,
) iter.Seq2[string, error] {
	return iter.Seq2[string, error](func(yield func(string, error) bool) {
		cc, err := a.CtxManager.GetContext(ctx)
		if err != nil {
			a.Logger.With(
				"err", err,
			).ErrorContext(ctx, "failed to get context")
			yield("", fmt.Errorf("failed to get context: %w", err))
			return
		}

		log := a.Logger.With(
			"user_id", cc.UserID,
			"session_id", cc.SessionID,
		)

		log.With(
			"prompt", prompt,
		).InfoContext(ctx, "asking agent...")

		win, err := a.Agent.BuildContextWindow(ctx, dom.Caller, cc.Window, time.Now())
		if err != nil {
			yield("", err)
			return
		}

		turn := cc.Window.Turns + 1
		infs := a.ask(ctx, prompt, win, yield)
		inter := dom.Interaction{
			Inferences: infs,
			TurnNumber: turn,
		}

		if cc.Window == nil {
			cc.Window = &dom.ContextWindow{}
		}
		cc.Window.History = append(cc.Window.History, inter)

		if err = a.CtxManager.SetContext(ctx, cc, &inter); err != nil {
			log.With(
				"err", err,
			).ErrorContext(ctx, "failed to set context")
			yield("", err)
			return
		}

		log.With(
			"number_of_inferences", len(infs),
			"total_input_tokens", infs[0].InputTokens+infs[1].InputTokens,
			"total_output_tokens", infs[0].OutputTokens+infs[0].OutputTokens,
			"turn", turn,
		).InfoContext(ctx, "agent answered succesfully")
	})
}

func (a *AskSvc) ask(
	ctx context.Context,
	prompt string,
	win []*dom.LLMContent,
	yield func(string, error) bool,
) [2]*dom.Inference {
	var infs [2]*dom.Inference

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

		infs = [2]*dom.Inference{
			&dom.Inference{
				Input: &dom.InputPrompt{
					Text: prompt,
				},
				Output: &dom.ModelOutput{
					FunctionCall: &dom.LLMFunctionCall{
						Name: results[0].Output.FunctionCall.Name,
						Args: results[0].Output.FunctionCall.Args,
					},
				},
				InputTokens:  results[0].Usage.InputTokens,
				OutputTokens: results[0].Usage.OutputTokens,
				Model:        dom.AgentToModel[dom.Caller],
			},
			&dom.Inference{
				Input: &dom.InputPrompt{
					FunctionResponse: &dom.LLMFunctionResponse{
						Name:    fnResp.Name,
						Content: fnResp.Content,
					},
				},
				Output: &dom.ModelOutput{
					Text: results[1].Output.Text,
				},
				InputTokens:  results[1].Usage.InputTokens,
				OutputTokens: results[1].Usage.OutputTokens,
				Model:        dom.AgentToModel[dom.Generator],
			},
		}

		return infs
	}

	infs = [2]*dom.Inference{
		&dom.Inference{
			Input: &dom.InputPrompt{
				Text: prompt,
			},
			Output: &dom.ModelOutput{
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

type ContextManager struct {
	dom.Cache
	dom.ContextRepo
	dom.Publisher
	Logger *slog.Logger
}

func BuildContextManager(ca dom.Cache, ctx dom.ContextRepo, pub dom.Publisher, log *slog.Logger) *ContextManager {
	log = log.With(
		"component", "ContextManager",
	)

	return &ContextManager{
		Cache:       ca,
		ContextRepo: ctx,
		Publisher:   pub,
		Logger:      log,
	}
}

func (c *ContextManager) GetContext(ctx context.Context) (*dom.ContextCache, error) {
	userID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	sessionID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	log := c.Logger.With(
		"method", "GetContext",
	)

	cc, err := c.getContextCache(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	if cc == nil {
		log.WarnContext(ctx, "cache miss, pulling from db...")
		window, err := c.getDbContext(ctx, userID, sessionID)
		if err != nil {
			return nil, err
		}

		cc = &dom.ContextCache{
			UserID:    userID,
			SessionID: sessionID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Window:    window,
		}
	}

	return cc, nil
}

func (c *ContextManager) getContextCache(ctx context.Context, userID, sessionID uuid.UUID) (*dom.ContextCache, error) {
	key := dom.CreateContextCacheKey(userID, sessionID)
	bytes, err := c.Cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if bytes == nil {
		return nil, nil
	}

	var cc dom.ContextCache
	if err := json.Unmarshal(bytes, &cc); err != nil {
		return nil, err
	}

	return &cc, nil
}

func (c *ContextManager) getDbContext(
	ctx context.Context,
	userID, sessionID uuid.UUID,
) (*dom.ContextWindow, error) {
	g, ctx := errgroup.WithContext(ctx)

	var (
		memories []dom.Memory
		sessions []dom.Session
		messages []dom.Message
	)

	g.Go(func() error {
		mem, err := c.MemoryRepo.GetMemoriesByUserID(ctx, userID, 50)
		if err != nil {
			return err
		}
		memories = mem
		return nil
	})

	g.Go(func() error {
		prev, err := c.SessionRepo.GetSessionsByUserID(ctx, userID, 5)
		if err != nil {
			return err
		}
		sessions = prev
		return nil
	})

	g.Go(func() error {
		msgs, err := c.MessageRepo.GetMessagesBySessionIDOrdered(ctx, sessionID)
		if err != nil {
			return err
		}
		messages = msgs
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	var (
		interactions []dom.Interaction
		inf1         dom.Inference = dom.Inference{}
		inf2         dom.Inference = dom.Inference{}
		has2infs     bool          = false
	)
	for _, m := range messages {
		meta := m.Meta()
		role := m.Role()
		switch role {
		case dom.UserRole:
			inf1 = dom.Inference{
				Input: &dom.InputPrompt{
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
			inf1.Output.FunctionCall = &dom.LLMFunctionCall{
				Name: *meta.FunctionName,
				Args: call,
			}
			inf1.OutputTokens = *meta.TotalOutputTokens
			inf2 = dom.Inference{
				Input: &dom.InputPrompt{
					FunctionResponse: &dom.LLMFunctionResponse{
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
				inf1.Model = *meta.Model
			} else {
				inf2.Output.Text = *meta.Content
				inf2.OutputTokens = *meta.TotalOutputTokens
				inf2.Model = *meta.Model
			}

			has2infs = false
			interactions = append(interactions, dom.Interaction{
				Inferences: [2]*dom.Inference{&inf1, &inf2},
				TurnNumber: meta.Turn,
			})
			inf1 = dom.Inference{}
			inf2 = dom.Inference{}
		}
	}

	var turns int32 = 0
	if len(interactions) > 0 {
		turns = interactions[len(interactions)-1].TurnNumber
	}

	return &dom.ContextWindow{
		UserMemories:     memories,
		PreviousSessions: sessions,
		History:          interactions,
		Turns:            turns,
	}, nil
}

func (c *ContextManager) SetContext(
	ctx context.Context,
	cc *dom.ContextCache,
	interaction *dom.Interaction,
) error {
	if err := c.setContextCache(ctx, cc.UserID, cc.SessionID, cc.Window); err != nil {
		return err
	}

	if err := c.sendContextUpdate(ctx, cc.UserID, cc.SessionID, interaction); err != nil {
		return err
	}
	return nil
}

func (c *ContextManager) setContextCache(
	ctx context.Context,
	userID, sessionID uuid.UUID,
	win *dom.ContextWindow,
) error {
	now := time.Now()
	win.Turns += 1
	bytes, err := json.Marshal(&dom.ContextCache{
		UserID:    userID,
		SessionID: sessionID,
		CreatedAt: now,
		UpdatedAt: now,
		Window:    win,
	})
	if err != nil {
		return err
	}

	key := dom.CreateContextCacheKey(userID, sessionID)

	if err = c.Cache.Set(ctx, key, bytes, dom.ContextCacheTTL6Hrs); err != nil {
		return err
	}

	return nil
}

func (c *ContextManager) sendContextUpdate(
	ctx context.Context,
	userID, sessionID uuid.UUID,
	interaction *dom.Interaction,
) error {
	load := &dom.SyncPayload{
		UserID:      userID,
		SessionID:   sessionID,
		Interaction: interaction,
	}

	data, err := json.Marshal(load)
	if err != nil {
		return err
	}

	ack, err := c.Publisher.Publish(ctx, SyncerSubject, data, &dom.PubOptions{
		MsgID: fmt.Sprintf("%s:%s:%d", userID, sessionID, interaction.TurnNumber),
	})
	if err != nil {
		return err
	}

	if ack.Stream != dom.ContextStream {
		return fmt.Errorf("published to unexpected stream: %s", ack.Stream)
	}

	return nil
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
