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

type PromptWithFilterDTO struct {
	Prompt             string            `json:"prompt"`
	ContentTypeFilters []string          `json:"content_type_filters,omitempty"`
	SourceFilters      []string          `json:"source_filters,omitempty"`
	SurahAyahFilters   *SurahAyahFilters `json:"surah_ayah_filters,omitempty"`
}

type FnSearchSchema struct {
	FullPrompt         string                `json:"full_prompt"`
	PromptsWithFilters []PromptWithFilterDTO `json:"prompts_with_filters"`
}

type FnSearch struct {
	SearchSvc *SearchSvc
}

func (f *FnSearch) Call(
	ctx context.Context,
	args map[string]any,
) (map[string]any, error) {
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

		contentTypes := toContentTypes(pmap["content_type_filters"].([]string))
		sources := toSources(pmap["source_filters"].([]string))

		var surahs []dom.SurahNumber
		var ayahs []dom.AyahNumber
		if surahAyah, ok := pmap["surah_ayah_filters"].(map[string]any); ok {
			surahs = toSurahNumbers(surahAyah["surahs"].([]int))
			ayahs = toAyahNumbers(surahAyah["ayahs"].([]int))
		}

		prompts = append(prompts, rag.PromptWithFilters{
			Prompt:               prompt,
			NullableContentTypes: contentTypes,
			NullableSources:      sources,
			NullableSurahs:       surahs,
			NullableAyahs:        ayahs,
		})
	}

	log := f.logger.With(
		slog.String("full_prompt", fullPrompt),
		slog.Group("prompts_with_filters",
			slog.Int("count", len(prompts)),
			slog.Any("items", prompts),
		),
	)

	log.DebugContext(ctx, "searcher agent called Search() function")

	params := rag.SearchParameters{
		RawPrompt:          fullPrompt,
		ChunkLimit:         rag.Top10Documents,
		PromptsWithFilters: prompts,
	}

	results, err := f.pipeline.Search(ctx, params)
	if err != nil {
		log.With(
			"err", err,
		).ErrorContext(ctx, "searcher agent failed to call Search() function")
		return nil, fmt.Errorf("searcher agent failed to call Search() function: %w", err)
	}

	serialized := make([]map[string]any, 0, len(results))
	for _, r := range results {
		serialized = append(serialized, map[string]any{
			"relevance": r.Relevance,
			"source":    r.Source,
			"document":  r.EmbeddedChunk,
			"surah":     r.Surah,
			"ayah":      r.Ayah,
		})
	}

	return map[string]any{
		"results": serialized,
	}, nil
}
