package agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/awbalessa/shaikh/backend/internal/database"
	"github.com/awbalessa/shaikh/backend/internal/models"
	"github.com/awbalessa/shaikh/backend/internal/rag"
	"google.golang.org/genai"
)

const (
	search functionName = "Search()"
)

var (
	surahNumMin float64 = 1
	surahNumMax float64 = 114
	ayahNumMin  float64 = 1
	ayahNumMax  float64 = 286
)

type functionName string
type enumContentType string
type enumSource string

func ptr[T any](t T) *T {
	return &t
}

type functionSearch struct {
	fname       functionName
	declaration *genai.FunctionDeclaration
	pipeline    *rag.Pipeline
	logger      *slog.Logger
}

type surahAyahFilters struct {
	Surahs []int `json:"surahs,omitempty"`
	Ayahs  []int `json:"ayahs,omitempty"`
}

type promptWithFilter struct {
	Prompt             string           `json:"prompt,omitempty"`
	ContentTypeFilters []string         `json:"content_type_filters,omitempty"`
	SourceFilters      []string         `json:"source_filters,omitempty"`
	SurahAyahFilters   surahAyahFilters `json:"surah_ayah_filters"`
}

type functionSearchSchema struct {
	FullPrompt         string             `json:"full_prompt"`
	PromptsWithFilters []promptWithFilter `json:"prompts_with_filters"`
}

func (f *functionSearch) name() functionName {
	return f.fname
}

func (f *functionSearch) call(
	ctx context.Context,
	args map[string]any,
) (map[string]any, error) {
	// Parse 'full_prompt'
	fullPrompt, ok := args["full_prompt"].(string)
	if !ok {
		return nil, errors.New("missing or invalid 'full_prompt'")
	}

	// Parse 'prompt_with_filter'
	argPrompts, ok := args["prompts_with_filter"].([]any)
	if !ok {
		return nil, errors.New("missing or invalid 'prompt_with_filter'")
	}

	prompts := make([]rag.PromptWithFilters, 0, len(argPrompts))
	for _, raw := range argPrompts {
		pmap, ok := raw.(map[string]any)
		if !ok {
			return nil, errors.New("invalid prompts_with_filter entry")
		}

		prompt, _ := pmap["prompt"].(string)

		// Optional filters
		contentTypes := toContentTypes(pmap["content_type_filters"].([]string))
		sources := toSources(pmap["source_filters"].([]string))

		var surahs []models.SurahNumber
		var ayahs []models.AyahNumber
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

func buildFunctionSearch(log *slog.Logger) *functionSearch {
	filterCts := &genai.Schema{
		Title:       "Optional Content Types Filter",
		Type:        genai.TypeArray,
		Description: "Optional filter for content types. Use this filter only when the user's intent explicitly matches one or more of the available filter options. Otherwise, leave this filter empty to allow a broader result set.",
		Items: &genai.Schema{
			Title:       "Content Type",
			Type:        genai.TypeString,
			Format:      "enum",
			Description: "The content type to filter by.",
			Enum:        []string{string(database.ContentTypeTafsir)},
		},
	}

	filterSrcs := &genai.Schema{
		Title:       "Optional Sources Filter",
		Type:        genai.TypeArray,
		Description: "Optional filter for sources. Use this filter only when the user clearly refers to one or more specific sources, authors, or references that match the available filter options. Otherwise, leave this filter empty to allow a broader result set.",
		Items: &genai.Schema{
			Title:       "Source",
			Type:        genai.TypeString,
			Format:      "enum",
			Description: "The source to filter by.",
			Enum:        []string{string(database.SourceTafsirIbnKathir)},
		},
	}

	filterSurahAyah := &genai.Schema{
		Title:       "Optional Surah and Ayah Filters",
		Type:        genai.TypeObject,
		Description: "You may filter results by a list of surah numbers. Optionally, if exactly one surah is specified, you may filter by a list of specific ayah numbers within that surah. Use this filter only when the user's prompt shows an interest in a specific part of the Quran. Otherwise, leave this filter empty to allow a broader result set.",
		Required:    []string{"surahs"},
		Properties: map[string]*genai.Schema{
			"surahs": {
				Title:       "Surah Numbers",
				Description: "A list of surah numbers to filter by. If more than one is provided, ayah filtering will be ignored.",
				Type:        genai.TypeArray,
				Items: &genai.Schema{
					Type:    genai.TypeInteger,
					Format:  "int32",
					Minimum: ptr(surahNumMin),
					Maximum: ptr(surahNumMax),
				},
				MinItems: ptr(int64(1)),
			},
			"ayahs": {
				Title:       "Ayah Numbers",
				Description: "A list of specific ayah numbers to filter by. Only allowed when exactly one surah is selected.",
				Type:        genai.TypeArray,
				Items: &genai.Schema{
					Type:    genai.TypeInteger,
					Format:  "int32",
					Minimum: &ayahNumMin,
					Maximum: &ayahNumMax,
				},
			},
		},
	}

	promptWithFilterSchema := &genai.Schema{
		Title:       "Prompts With Filters",
		Type:        genai.TypeArray,
		Description: "Logical subunits of the full prompt. In most cases, this array will contain the full prompt itself with no other entries. But in advanced use cases (like step-back prompting or multi-question prompts), you may split the full prompt into multiple logically distinct sub-prompts, each with its own filter context.",
		MinItems:    ptr(int64(1)),
		MaxItems:    ptr(int64(3)),
		Items: &genai.Schema{
			Title:       "Prompt With Optional Filters",
			Type:        genai.TypeObject,
			Description: "A prompt string with optional filters to constrain the context. This is one logical unit of the full prompt.",
			Required:    []string{"prompt"},
			Properties: map[string]*genai.Schema{
				"prompt": {
					Title:       "Sub-Prompt",
					Type:        genai.TypeString,
					Description: "A logical unit or sub-question derived from the full raw prompt. If only one is provided, it is typically the entire full prompt.",
				},
				"content_type_filters": filterCts,
				"source_filters":       filterSrcs,
				"surah_ayah_filters":   filterSurahAyah,
			},
		},
	}

	example := map[string]any{
		"full_prompt": "ما قصة موسى مع الخضر كما وردت في سورة الكهف؟ وماذا نتعلم منها؟ وهل توجد مواضع أخرى في القرآن تشير إلى هذا النوع من العلم الغيبي؟",
		"prompts_with_filters": []any{
			map[string]any{
				"prompt": "قصة موسى مع الخضر كما وردت في سورة الكهف",
				"surah_ayah_filters": map[string]any{
					"surahs": []int{18},
				},
			},
			map[string]any{
				"prompt": "الدروس والعبر المستفادة من قصة موسى والخضر",
			},
			map[string]any{
				"prompt":               "شرح ابن كثير حول قصة موسى والخضر في سورة الكهف",
				"content_type_filters": []string{"tafsir"},
				"source_filters":       []string{"ibn_kathir"},
				"surah_ayah_filters": map[string]any{
					"surahs": []int{18},
				},
			},
		},
	}

	fullSchema := &genai.Schema{
		Title:       "Search() Parameters",
		Type:        genai.TypeObject,
		Description: "The input parameters for performing a hybrid search—combining semantic similarity and keyword matching—based on the fully transformed prompt. The prompt may be optionally broken into logical sub-prompts, each with its own filters to narrow down the context.",
		Required:    []string{"full_prompt", "prompts_with_filters"},
		Example:     example,
		Properties: map[string]*genai.Schema{
			"full_prompt": {
				Title:       "Full Prompt",
				Type:        genai.TypeString,
				Description: "The fully transformed version of the user's prompt. This includes accurate translation into Arabic (if submitted in another language), normalization from question form to statement form, and typo correction. It is the canonical form used as the semantic base for search.",
			},
			"prompts_with_filters": promptWithFilterSchema,
		},
	}

	return &functionSearch{
		declaration: &genai.FunctionDeclaration{
			Name:        string(search),
			Description: "Performs a hybrid search over Quranic content using a fully normalized prompt. Combines semantic understanding with keyword-based matching. The prompt may be optionally split into sub-prompts with filters to target specific content types, sources, surahs, or ayahs.",
			Parameters:  fullSchema,
		},
		logger: log,
	}
}

func toContentTypes(v any) []database.ContentType {
	raw, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []database.ContentType
	for _, item := range raw {
		if s, ok := item.(string); ok {
			out = append(out, database.ContentType(s))
		}
	}
	return out
}

func toSources(v any) []database.Source {
	raw, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []database.Source
	for _, item := range raw {
		if s, ok := item.(string); ok {
			out = append(out, database.Source(s))
		}
	}
	return out
}

func toSurahNumbers(v any) []models.SurahNumber {
	raw, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []models.SurahNumber
	for _, item := range raw {
		if f, ok := item.(float64); ok { // JSON numbers -> float64
			out = append(out, models.SurahNumber(int(f)))
		}
	}
	return out
}

func toAyahNumbers(v any) []models.AyahNumber {
	raw, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []models.AyahNumber
	for _, item := range raw {
		if f, ok := item.(float64); ok {
			out = append(out, models.AyahNumber(int(f)))
		}
	}
	return out
}
