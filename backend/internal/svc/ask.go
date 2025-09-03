package svc

import (
	"context"
	"encoding/json"
	"errors"
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

func BuildAskSvc(
	ag dom.Agent,
	fns dom.LLMFunctions,
	ctx *ContextManager,
	se *SearchSvc,
) *AskSvc {
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
	userID, sessionID uuid.UUID,
) iter.Seq2[string, error] {
	return iter.Seq2[string, error](func(yield func(string, error) bool) {
		cc, err := a.CtxManager.GetContext(ctx, userID, sessionID)
		if err != nil {
			a.Logger.With(
				"err", err,
			).ErrorContext(ctx, "failed to get context")
			yield("", fmt.Errorf("failed to get context: %w", err))
			return
		}
		if cc.Window == nil {
			cc.Window = &dom.ContextWindow{}
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
			yield("", fmt.Errorf("failed to build context window: %w", err))
			return
		}

		infs := a.ask(ctx, prompt, win, yield)
		turn := cc.Window.Turns + 1
		inter := &dom.Interaction{
			Inferences: infs,
			TurnNumber: turn,
		}

		cc.Window.History = append(cc.Window.History, inter)

		if err = a.CtxManager.SetContext(ctx, cc, inter); err != nil {
			log.With(
				"err", err,
			).ErrorContext(ctx, "failed to set context")
			yield("", fmt.Errorf("failed to set context: %w", err))
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
		a.Agent.StreamWithYield(ctx, dom.Caller, win, syield),
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
			a.Agent.StreamWithYield(ctx, dom.Generator, win, syield),
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
	dom.MemoryRepo
	dom.SessionRepo
	dom.MessageRepo
	dom.Publisher
	Logger *slog.Logger
}

const (
	ContextStream            string        = "CONTEXT"
	ContextStreamSubject     string        = "context"
	ContextStreamSubjectStar string        = "context.*"
	ContextStreamMaxAge      time.Duration = 24 * time.Hour
)

func BuildContextManager(
	ctx context.Context,
	ca dom.Cache,
	mr dom.MessageRepo,
	memr dom.MemoryRepo,
	sr dom.SessionRepo,
	pub dom.Publisher,
	ps dom.PubSub,
	log *slog.Logger,
) (*ContextManager, error) {
	log = log.With(
		"component", "ContextManager",
	)

	cfg := dom.PubSubStreamConfig{
		Name:      ContextStream,
		Subjects:  []string{ContextStreamSubjectStar},
		Retention: dom.WorkQueue,
		Storage:   dom.FileStorage,
		MaxAge:    ContextStreamMaxAge,
	}

	if err := ps.CreateStream(ctx, cfg); err != nil {
		return nil, err
	}

	return &ContextManager{
		Cache:       ca,
		SessionRepo: sr,
		MemoryRepo:  memr,
		MessageRepo: mr,
		Publisher:   pub,
		Logger:      log,
	}, nil
}

const (
	ContextCacheTTL6Hrs time.Duration = 6 * time.Hour
)

func (c *ContextManager) GetContext(ctx context.Context, userID, sessionID uuid.UUID) (*dom.ContextCache, error) {
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
		memories []*dom.Memory
		sessions []*dom.Session
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
		interactions []*dom.Interaction
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
			call, err := dom.FromJsonRawMessage(meta.FunctionCall)
			if err != nil {
				return nil, err
			}
			resp, err := dom.FromJsonRawMessage(meta.FunctionResponse)
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
			interactions = append(interactions, &dom.Interaction{
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

	if err = c.Cache.Set(ctx, key, bytes, ContextCacheTTL6Hrs); err != nil {
		return err
	}

	return nil
}

func (c *ContextManager) sendContextUpdate(
	ctx context.Context,
	userID, sessionID uuid.UUID,
	interaction *dom.Interaction,
) error {
	load := &SyncPayload{
		UserID:      userID,
		SessionID:   sessionID,
		Interaction: interaction,
	}

	data, err := json.Marshal(load)
	if err != nil {
		return err
	}

	ack, err := c.Publisher.DurablePublish(ctx, SyncerSubject, data, &dom.DurablePubOptions{
		MsgID: fmt.Sprintf("sync:%s:%s:%d", userID, sessionID, interaction.TurnNumber),
	})
	if err != nil {
		return err
	}

	if ack.Stream != ContextStream {
		return fmt.Errorf("published to unexpected stream: %s", ack.Stream)
	}

	return nil
}

type SurahAyahFilters struct {
	Surahs []int32 `json:"surahs,omitempty"`
	Ayahs  []int32 `json:"ayahs,omitempty"`
}

type PWFFnSearch struct {
	Prompt             string            `json:"prompt"`
	ContentTypeFilters []string          `json:"content_type_filters,omitempty"`
	SourceFilters      []string          `json:"source_filters,omitempty"`
	SurahAyahFilters   *SurahAyahFilters `json:"surah_ayah_filters,omitempty"`
}

type FnSearchSchema struct {
	FullPrompt         string        `json:"full_prompt"`
	PromptsWithFilters []PWFFnSearch `json:"prompts_with_filters"`
}

type FnSearch struct {
	SearchSvc *SearchSvc
	Logger    *slog.Logger
}

func BuildFnSearch(se *SearchSvc, log *slog.Logger) *FnSearch {
	log = log.With(
		"component", "FnSearch",
	)

	return &FnSearch{
		SearchSvc: se,
		Logger:    log,
	}
}

func (f *FnSearch) Call(
	ctx context.Context,
	args map[string]any,
) (map[string]any, error) {
	log := f.Logger.With(
		"method", "Call",
	)

	fullPrompt, ok := args["full_prompt"].(string)
	if !ok {
		return nil, errors.New("missing or invalid 'full_prompt'")
	}

	argPrompts, ok := args["prompts_with_filter"].([]any)
	if !ok {
		return nil, errors.New("missing or invalid 'prompt_with_filter'")
	}

	prompts := make([]dom.QueryWithFilter, 0, len(argPrompts))
	for _, raw := range argPrompts {
		pmap, ok := raw.(map[string]any)
		if !ok {
			return nil, errors.New("invalid prompts_with_filter entry")
		}

		prompt, _ := pmap["prompt"].(string)

		contentTypes := dom.RawToContentTypes(pmap["content_type_filters"].([]string))
		sources := dom.RawToSources(pmap["source_filters"].([]string))

		var surahs []dom.SurahNumber
		var ayahs []dom.AyahNumber
		if surahAyah, ok := pmap["surah_ayah_filters"].(map[string]any); ok {
			surahs = dom.RawToSurahNumbers(surahAyah["surahs"].([]int))
			ayahs = dom.RawToAyahNumbers(surahAyah["ayahs"].([]int))
		}

		prompts = append(prompts, dom.QueryWithFilter{
			Query: prompt,
			FilterContext: dom.FilterContext{
				OptionalContentTypes: contentTypes,
				OptionalSources:      sources,
				OptionalSurahs:       surahs,
				OptionalAyahs:        ayahs,
			},
		})
	}

	log.With(
		"full_prompt", fullPrompt,
		"prompts_with_filter_count", len(prompts),
		"raw", args,
	).DebugContext(ctx, "agent called Search() function")

	params := dom.SearchQuery{
		FullQuery:          fullPrompt,
		QueriesWithFilters: prompts,
		TopK:               dom.Top20Documents,
	}

	results, err := f.SearchSvc.Search(ctx, params)
	if err != nil {
		log.With(
			"err", err,
		).ErrorContext(ctx, "agent failed to call Search() function")
		return nil, fmt.Errorf("agent failed to call Search() function: %w", err)
	}

	serialized := make([]map[string]any, 0, len(results))
	for _, r := range results {
		serialized = append(serialized, map[string]any{
			"relevance": r.Relevance,
			"source":    r.Source,
			"document":  r.Content,
			"surah":     r.SurahNumber,
			"ayah":      r.AyahNumber,
			"parent_id": r.ParentID,
		})
	}

	return map[string]any{
		"results": serialized,
	}, nil
}

type CtxKey string

const (
	CtxUserIDKey    CtxKey = "userID"
	CtxSessionIDKey CtxKey = "sessionID"
)

func UserIDFromCtx(ctx context.Context) (uuid.UUID, error) {
	v := ctx.Value(CtxUserIDKey)
	id, ok := v.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("missing or invalid userID in context")
	}
	return id, nil
}

func SessionIDFromCtx(ctx context.Context) (uuid.UUID, error) {
	v := ctx.Value(CtxSessionIDKey)
	id, ok := v.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("missing or invalid sessionID in context")
	}
	return id, nil
}
