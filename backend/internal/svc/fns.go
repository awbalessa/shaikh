package svc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/awbalessa/shaikh/backend/internal/dom"
)

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
