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

func (a *AskSvc) Ask(
	ctx context.Context,
	prompt string,
	userID, sessionID uuid.UUID,
	reqID string,
) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		log := slog.Default().With(
			"service", "ask",
			"reqID", reqID,
			"userID", userID,
			"sessionID", sessionID,
			"prompt", prompt,
		)

		// 1. Get context. If it fails, log and yield immediately.
		ccRes, err := a.CtxManager.GetContext(ctx, userID, sessionID)
		if err != nil {
			err = dom.ToDomainError(err)
			log.ErrorContext(ctx, "ask service error", "err", err)
			yield("", err)
			return
		}
		if ccRes.Result.Window == nil {
			ccRes.Result.Window = &dom.ContextWindow{}
		}

		// 2. Build window. If it fails, log and yield immediately.
		win, err := a.Agent.BuildContextWindow(ctx, dom.Caller, ccRes.Result.Window, time.Now())
		if err != nil {
			err = dom.ToDomainError(err)
			log.ErrorContext(ctx, "ask service error", "err", err)
			yield("", err)
			return
		}

		// 3. Perform the streaming call.
		var streamingErr error
		ar := a.ask(ctx, prompt, win, func(str string, e error) bool {
			if e != nil {
				streamingErr = dom.ToDomainError(e) // Capture the error
				return yield("", streamingErr)      // and yield it.
			}
			return yield(str, nil)
		})

		// 4. Handle post-stream logic and final logging.
		if streamingErr != nil {
			// The error was already sent to the client, but we log it here at the boundary.
			log.ErrorContext(ctx, "ask service stream error", "err", streamingErr)
			return
		}

		if ar != nil {
			// Persist context. If this fails, we can only log it.
			if err := a.CtxManager.SetContext(ctx, ccRes.Result, &dom.Interaction{
				Inferences: ar.infs,
				TurnNumber: ccRes.Result.Window.Turns + 1,
			}); err != nil {
				log.ErrorContext(ctx, "failed to persist context after stream", "err", err)
			}

			// Finally, log the successful completion with all metadata.
			log.DebugContext(ctx, "ask service completed",
				"userID", userID,
				"sessionID", sessionID,
				"prompt", prompt,
				"context", ccRes.Metadata,
				"pipeline", ar.metadata,
			)
		}
	}
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

	firstStageMS := time.Since(start).Milliseconds()
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
				"duration_ms":    firstStageMS,
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
		return nil, dom.NewTaggedError(dom.ErrInvalidInput, nil)
	}

	res, err := function.Call(ctx, fn.Args)
	if err != nil {
		return nil, err
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
		return nil, dom.NewTaggedError(dom.ErrInvalidInput, nil)
	}

	rawPrompts, ok := args["prompts_with_filters"].([]any)
	if !ok {
		return nil, dom.NewTaggedError(dom.ErrInvalidInput, nil)
	}
	if len(rawPrompts) == 0 {
		return nil, dom.NewTaggedError(dom.ErrInvalidInput, nil)
	}

	prompts := make([]dom.QueryWithFilter, 0, len(rawPrompts))
	subQueryMeta := make([]map[string]any, 0, len(rawPrompts))

	for _, raw := range rawPrompts {
		pmap, ok := raw.(map[string]any)
		if !ok {
			return nil, dom.NewTaggedError(dom.ErrInvalidInput, nil)
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
		return nil, err
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
			UserID:    &userID,
			SessionID: &sessionID,
			CreatedAt: dom.Ptr(time.Now()),
			UpdatedAt: dom.Ptr(time.Now()),
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
			i1 := inf1
			i2 := inf2
			interactions = append(interactions, &dom.Interaction{
				Inferences: []*dom.Inference{&i1, &i2},
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
	if err := c.setContextCache(ctx, cc); err != nil {
		return err
	}

	if err := c.sendContextUpdate(ctx, cc, interaction); err != nil {
		return err
	}
	return nil
}

func (c *ContextManager) setContextCache(
	ctx context.Context,
	cc *dom.ContextCache,
) error {
	now := time.Now()
	var createdAt time.Time = now
	if cc.CreatedAt != nil {
		createdAt = *cc.CreatedAt
	}

	cc.Window.Turns += 1
	bytes, err := json.Marshal(&dom.ContextCache{
		UserID:    cc.UserID,
		SessionID: cc.SessionID,
		CreatedAt: &createdAt,
		UpdatedAt: &now,
		Window:    cc.Window,
	})
	if err != nil {
		return err
	}

	key := dom.CreateContextCacheKey(*cc.UserID, *cc.SessionID)

	if err = c.Cache.Set(ctx, key, bytes, ContextCacheTTL6Hrs); err != nil {
		return err
	}

	return nil
}

func (c *ContextManager) sendContextUpdate(
	ctx context.Context,
	cc *dom.ContextCache,
	interaction *dom.Interaction,
) error {
	load := &SyncPayload{
		UserID:      *cc.UserID,
		SessionID:   *cc.SessionID,
		Interaction: interaction,
	}

	data, err := json.Marshal(load)
	if err != nil {
		return err
	}

	ack, err := c.Publisher.DurablePublish(ctx, SyncerSubject, data, &dom.DurablePubOptions{
		MsgID: fmt.Sprintf("sync:%s:%s:%d", *cc.UserID, *cc.SessionID, interaction.TurnNumber),
	})
	if err != nil {
		return err
	}

	if ack.Stream != ContextStream {
		return dom.NewTaggedError(dom.ErrInvalidState, nil)
	}

	return nil
}
