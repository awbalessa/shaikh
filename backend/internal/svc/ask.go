package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"maps"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/dom"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type AskSvc struct {
	Agent      dom.Agent
	Functions  dom.AgentFns
	CtxManager *ContextManager
	SearchSvc  *SearchSvc
}

func BuildAskSvc(
	ctx context.Context,
	ag dom.Agent,
	cache dom.Cache,
	memr dom.MemoryRepo,
	sr dom.SessionRepo,
	mr dom.MessageRepo,
	ps dom.PubSub,
	se *SearchSvc,
) (*AskSvc, error) {
	fns := map[dom.AgentFnName]dom.AgentFn{
		dom.FunctionSearch: BuildFnSearch(se),
	}

	ctxManager, err := BuildContextManager(
		ctx, cache,
		mr, memr,
		sr, ps,
	)
	if err != nil {
		return nil, err
	}

	return &AskSvc{
		Agent:      ag,
		Functions:  fns,
		CtxManager: ctxManager,
		SearchSvc:  se,
	}, nil
}

type AskResult struct {
	Stream   iter.Seq2[string, *dom.Err]
	Metadata func() map[string]any
}

func (a *AskSvc) Ask(
	ctx context.Context,
	prompt string,
	userID, sessionID uuid.UUID,
) (*AskResult, *dom.Err) {
	ccRes, err := a.CtxManager.GetContext(ctx, userID, sessionID)
	if err != nil {
		return nil, dom.ToDomErr(err)
	}
	if ccRes.Result.Window == nil {
		ccRes.Result.Window = &dom.ContextWindow{}
	}

	win, err := a.Agent.BuildContextWindow(ctx, dom.Caller, ccRes.Result.Window, time.Now())
	if err != nil {
		return nil, dom.ToDomErr(err)
	}

	var ar *askResult
	stream := iter.Seq2[string, *dom.Err](func(yield func(string, *dom.Err) bool) {
		if ctx.Err() != nil {
			_ = yield("", dom.ToDomErr(ctx.Err()))
			return
		}
		ar = a.ask(ctx, prompt, win, func(str string, err error) bool {
			var derr *dom.Err
			if err != nil {
				derr = dom.ToDomErr(err)
			}
			if !yield(str, derr) {
				return false
			}
			if ctx.Err() != nil {
				_ = yield("", dom.ToDomErr(ctx.Err()))
				return false
			}
			return true
		})
	})

	inter := &dom.Interaction{
		Inferences: [2]*dom.Inference(ar.infs),
		TurnNumber: ccRes.Result.Window.Turns + 1,
	}
	if err := a.CtxManager.SetContext(ctx, ccRes.Result, inter); err != nil {
		return nil, dom.ToDomErr(err)
	}

	return &AskResult{
		Stream: stream,
		Metadata: func() map[string]any {
			m := map[string]any{
				"userID":    userID,
				"sessionID": sessionID,
				"prompt":    prompt,
				"context":   ccRes.Metadata,
			}
			if ar.metadata != nil {
				maps.Copy(m, ar.metadata)
			}
			return m
		},
	}, nil
}

type askResult struct {
	infs     []*dom.Inference
	metadata map[string]any
}

func (a *AskSvc) ask(
	ctx context.Context,
	prompt string,
	win []*dom.LLMContent,
	yield func(string, error) bool,
) *askResult {
	start := time.Now()
	win = append(win, &dom.LLMContent{
		Role:  dom.LLMUserRole,
		Parts: []*dom.LLMPart{{Text: prompt}},
	})

	var infs []*dom.Inference
	syield := func(out dom.LLMOut, err error) bool {
		if err != nil {
			yield("", err)
			return false
		}
		if out.Text() != "" {
			if !yield(out.Text(), nil) {
				return false
			}
		}
		if out.FunctionCall() != nil {
			return false
		}

		return true
	}
	infs = append(infs, a.Agent.Stream(ctx, dom.Caller, win, syield))
	if infs[0].Output.FunctionCall == nil {
		infs[0].Input = &dom.LLMInput{
			Text:             prompt,
			FunctionResponse: nil,
		}
		return &askResult{
			infs: infs,
			metadata: map[string]any{
				"caller": map[string]any{
					"duration_ms":    time.Since(start).Milliseconds(),
					"input_tokens":   infs[0].InputTokens,
					"output_tokens":  infs[0].OutputTokens,
					"finish_message": infs[0].FinishMessage,
					"finish_reason":  infs[0].FinishReason,
				},
			},
		}
	}

	secondStart := time.Now()
	res, err := a.handleFn(ctx, infs[0].Output.FunctionCall)
	if err != nil {
		yield("", err)
		return nil
	}
	win = append(win, &dom.LLMContent{
		Role:  dom.LLMModelRole,
		Parts: []*dom.LLMPart{{FunctionCall: infs[0].Output.FunctionCall}},
	})
	win = append(win, &dom.LLMContent{
		Role:  dom.LLMUserRole,
		Parts: []*dom.LLMPart{{FunctionResponse: res}},
	})

	infs = append(infs, a.Agent.Stream(ctx, dom.Generator, win, syield))
	infs[1].Input = &dom.LLMInput{
		Text:             prompt,
		FunctionResponse: res,
	}
	return &askResult{
		infs: infs,
		metadata: map[string]any{
			"caller": map[string]any{
				"duration_ms":    secondStart.Sub(start).Milliseconds(),
				"input_tokens":   infs[0].InputTokens,
				"output_tokens":  infs[0].OutputTokens,
				"finish_message": infs[0].FinishMessage,
				"finish_reason":  infs[0].FinishReason,
			},
			"function_call": map[string]any{
				"name": res.Name,
				"meta": res.Metadata,
			},
			"generator": map[string]any{
				"duration_ms":    time.Since(secondStart).Milliseconds(),
				"input_tokens":   infs[1].InputTokens,
				"output_tokens":  infs[1].OutputTokens,
				"finish_message": infs[1].FinishMessage,
				"finish_reason":  infs[1].FinishReason,
			},
		},
	}

}

func (a *AskSvc) handleFn(
	ctx context.Context,
	fn *dom.LLMFunctionCall,
) (*dom.LLMFunctionResponse, error) {
	function, ok := a.Functions[dom.AgentFnName(fn.Name)]
	if !ok {
		return nil, fmt.Errorf("function %s does not exist: %w", fn.Name, dom.ErrInvalidInput)
	}

	res, err := function.Call(ctx, fn.Args)
	if err != nil {
		return nil, fmt.Errorf("handleFn: %w", err)
	}

	return &dom.LLMFunctionResponse{
		Name:     fn.Name,
		Content:  res.Response,
		Metadata: res.Metadata,
	}, nil
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
}

func BuildFnSearch(se *SearchSvc) *FnSearch {
	return &FnSearch{
		SearchSvc: se,
	}
}

func (f *FnSearch) Call(
	ctx context.Context,
	args map[string]any,
) (*dom.CallResult, error) {
	start := time.Now()

	fullPrompt, ok := args["full_prompt"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'full_prompt': %w", dom.ErrInvalidInput)
	}

	rawPrompts, ok := args["prompts_with_filter"].([]any)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'prompts_with_filter': %w", dom.ErrInvalidInput)
	}

	prompts := make([]dom.QueryWithFilter, 0, len(rawPrompts))
	subQueryMeta := make([]map[string]any, 0, len(rawPrompts))

	for _, raw := range rawPrompts {
		pmap, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid prompts_with_filter entry: %w", dom.ErrInvalidInput)
		}

		prompt, _ := pmap["prompt"].(string)

		rawCT := pmap["content_type_filters"]
		rawSO := pmap["source_filters"]

		var rawSurahAyah any = nil
		if sa, ok := pmap["surah_ayah_filters"].(map[string]any); ok {
			rawSurahAyah = sa
		}

		var surahs []dom.SurahNumber
		var ayahs []dom.AyahNumber
		if sa, ok := rawSurahAyah.(map[string]any); ok {
			surahs = dom.RawToSurahNumbers(sa["surahs"])
			ayahs = dom.RawToAyahNumbers(sa["ayahs"])
		}

		prompts = append(prompts, dom.QueryWithFilter{
			Query: prompt,
			FilterContext: dom.FilterContext{
				OptionalContentTypes: dom.RawToContentTypes(rawCT),
				OptionalSources:      dom.RawToSources(rawSO),
				OptionalSurahs:       surahs,
				OptionalAyahs:        ayahs,
			},
		})

		metaSA := map[string]any{"surahs": nil, "ayahs": nil}
		if sa, ok := rawSurahAyah.(map[string]any); ok {
			metaSA = map[string]any{
				"surahs": sa["surahs"],
				"ayahs":  sa["ayahs"],
			}
		}

		subQueryMeta = append(subQueryMeta, map[string]any{
			"prompt":               prompt,
			"content_type_filters": rawCT,
			"source_filters":       rawSO,
			"surah_ayah_filters":   metaSA,
		})
	}

	params := dom.SearchQuery{
		FullQuery:          fullPrompt,
		QueriesWithFilters: prompts,
		TopK:               dom.Top20Documents,
	}

	res, err := f.SearchSvc.Search(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("call: %w", err)
	}

	out := make([]map[string]any, 0, len(res.Results))
	for _, r := range res.Results {
		out = append(out, map[string]any{
			"relevance": r.Relevance,
			"source":    r.Source,
			"document":  r.Content,
			"surah":     r.SurahNumber,
			"ayah":      r.AyahNumber,
		})
	}

	metadata := map[string]any{
		"full_prompt": fullPrompt,
		"sub_queries": subQueryMeta,
		"duration_ms": time.Since(start).Milliseconds(),
		"search":      res.Metadata,
	}

	return &dom.CallResult{
		Response: map[string]any{"results": out},
		Metadata: metadata,
	}, nil
}

type ContextManager struct {
	dom.Cache
	dom.MemoryRepo
	dom.SessionRepo
	dom.MessageRepo
	dom.Publisher
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
	ps dom.PubSub,
) (*ContextManager, error) {
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

	pub := ps.Publisher()

	return &ContextManager{
		Cache:       ca,
		SessionRepo: sr,
		MemoryRepo:  memr,
		MessageRepo: mr,
		Publisher:   pub,
	}, nil
}

const (
	ContextCacheTTL6Hrs time.Duration = 6 * time.Hour
)

type GetContextResult struct {
	Result   *dom.ContextCache
	Metadata map[string]any
}

func (c *ContextManager) GetContext(ctx context.Context, userID, sessionID uuid.UUID) (*GetContextResult, error) {
	cc, err := c.getContextCache(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	var cacheHit bool = true
	if cc == nil {
		cacheHit = false
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

	return &GetContextResult{
		Result: cc,
		Metadata: map[string]any{
			"cache_hit":         cacheHit,
			"user_memories":     len(cc.Window.UserMemories),
			"session_summaries": len(cc.Window.PreviousSessions),
			"previous_turns":    cc.Window.Turns,
		},
	}, nil
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
		msgs, err := c.MessageRepo.GetMessagesBySessionID(ctx, sessionID)
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
		case dom.MessageRoleUser:
			inf1 = dom.Inference{
				Input: &dom.LLMInput{
					Text: *meta.Content,
				},
				InputTokens: *meta.TotalInputTokens,
			}
		case dom.MessageRoleFunction:
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
				Input: &dom.LLMInput{
					FunctionResponse: &dom.LLMFunctionResponse{
						Name:    *meta.FunctionName,
						Content: resp,
					},
				},
				InputTokens: *meta.TotalInputTokens,
			}

		case dom.MessageRoleModel:
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
		return fmt.Errorf("published to unexpected stream: %s, %w", ack.Stream, dom.ErrInvalidState)
	}

	return nil
}
