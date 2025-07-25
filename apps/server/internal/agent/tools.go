package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/awbalessa/shaikh/apps/server/internal/database"
	"github.com/awbalessa/shaikh/apps/server/internal/models"
	"github.com/awbalessa/shaikh/apps/server/internal/rag"
	"google.golang.org/genai"
)

type toolRAG struct {
	name      toolName
	functions map[functionName]function
	pipeline  *rag.Pipeline
	logger    *slog.Logger
}

type functionSearch struct {
	name        functionName
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

func (t *toolRAG) getName() toolName {
	return t.name
}

func (t *toolRAG) getFunction(fn functionName) (function, error) {
	function, ok := t.functions[fn]
	if !ok {
		return nil, fmt.Errorf("function %s does not exist", string(fn))
	}

	return function, nil
}

func buildToolRAG(p *rag.Pipeline, log *slog.Logger) *toolRAG {
	search := buildFunctionSearch(log)
	return &toolRAG{
		name: RAG,
		functions: map[functionName]function{
			Search: search,
		},
		pipeline: p,
		logger:   log,
	}
}

func (t *functionSearch) getName() functionName {
	return t.name
}

func (t *functionSearch) call(ctx context.Context, bytes []byte) (any, error) {
	var inp functionSearchSchema
	if err := json.Unmarshal(bytes, &inp); err != nil {
		t.logger.With(
			slog.String("function", string(t.name)),
		).ErrorContext(ctx, "failed to unmarshal agent output")
		return nil, fmt.Errorf("failed to unmarshal agent output: %w", err)
	}

	pwf := make([]rag.PromptWithFilters, len(inp.PromptsWithFilters))
	for i, p := range inp.PromptsWithFilters {
		pwf[i] = rag.PromptWithFilters{
			Prompt:               p.Prompt,
			NullableContentTypes: toContentTypes(p.ContentTypeFilters),
			NullableSources:      toSources(p.SourceFilters),
			NullableSurahs:       toSurahNumbers(p.SurahAyahFilters.Surahs),
			NullableAyahs:        toAyahNumbers(p.SurahAyahFilters.Ayahs),
		}
	}

	arg := rag.SearchParameters{
		RawPrompt:          inp.FullPrompt,
		ChunkLimit:         rag.Top20Documents,
		PromptsWithFilters: pwf,
	}
	return t.pipeline.Search(ctx, arg)
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
			Enum:        []string{string(contentTypeTafsir)},
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
			Enum:        []string{string(sourceTafsirIbnKathir)},
		},
	}

	filterSurahAyah := &genai.Schema{
		Title:       "Surah and Ayah Filters",
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
		Title:       "Search Parameters",
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
		name: Search,
		declaration: &genai.FunctionDeclaration{
			Name:        string(Search),
			Description: "Performs a hybrid search over Quranic content using a fully normalized prompt. Combines semantic understanding with keyword-based matching. The prompt may be optionally split into sub-prompts with filters to target specific content types, sources, surahs, or ayahs.",
			Parameters:  fullSchema,
		},
		logger: log,
	}
}

func toContentTypes(in []string) []database.ContentType {
	out := make([]database.ContentType, len(in))
	for i, s := range in {
		out[i] = database.ContentType(s)
	}
	return out
}

func toSources(in []string) []database.Source {
	out := make([]database.Source, len(in))
	for i, s := range in {
		out[i] = database.Source(s)
	}
	return out
}

func toSurahNumbers(in []int) []models.SurahNumber {
	out := make([]models.SurahNumber, len(in))
	for i, n := range in {
		out[i] = models.SurahNumber(n)
	}
	return out
}

func toAyahNumbers(in []int) []models.AyahNumber {
	out := make([]models.AyahNumber, len(in))
	for i, n := range in {
		out[i] = models.AyahNumber(n)
	}
	return out
}

const (
	contentTypeTafsir     enumContentType = "TAFSIR"
	sourceTafsirIbnKathir enumSource      = "TAFSIR IBN KATHIR"
)

var (
	surahNumMin float64 = 1
	surahNumMax float64 = 114
	ayahNumMin  float64 = 1
	ayahNumMax  float64 = 286
)

type enumContentType string
type enumSource string

func ptr[T any](t T) *T {
	return &t
}
