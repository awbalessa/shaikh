package svc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/dom"
	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/genai"
)

type AgentConfig struct {
	Context context.Context
	Stream  jetstream.JetStream
}

type AskSvc struct {
	agents    map[agentName]*agentProfile
	functions map[functionName]function
	Log       *slog.Logger
	gc        *geminiClient
	js        jetstream.JetStream
}

func NewAgent(cfg AgentConfig) (*Agent, error) {
	log := slog.Default().With(
		"component", "agent",
	)

	gc, err := newGeminiClient(geminiClientConfig{
		context:    cfg.Context,
		maxRetries: geminiMaxRetriesThree,
		timeout:    geminiTimeoutFifteenSeconds,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build new agent: %w", err)
	}

	se := buildSearcher(searcherConfig{
		pipe:   cfg.Pipeline,
		logger: log,
	})

	g := buildGenerator()

	fmap := map[functionName]function{
		search: buildFunctionSearch(log),
	}

	amap := map[agentName]*agentProfile{
		searcherAgent: {
			name:   searcherAgent,
			model:  se.model,
			config: se.baseCfg,
		},
		generatorAgent: {
			name:   generatorAgent,
			model:  g.model,
			config: g.baseCfg,
		},
	}

	_, err = cfg.Stream.CreateStream(cfg.Context, jetstream.StreamConfig{
		Name:        AgentStream,
		Subjects:    []string{"agent.context.*"},
		Retention:   jetstream.WorkQueuePolicy,
		Storage:     jetstream.FileStorage,
		MaxAge:      jsMsgsMaxAge,
		MaxMsgSize:  1 * 1024 * 1024,
		DenyDelete:  true,
		DenyPurge:   false,
		AllowRollup: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build new agent: %w", err)
	}

	return &Agent{
		agents:    amap,
		gc:        gc,
		functions: fmap,
		logger:    log,
		store:     cfg.Store,
		js:        cfg.Stream,
	}, nil
}

const (
	searcherAgent  agentName     = "searcher"
	generatorAgent agentName     = "generator"
	jsMsgsMaxAge   time.Duration = 24 * time.Hour
)

type agentName string

type agentProfile struct {
	name   agentName
	model  geminiModel
	config *genai.GenerateContentConfig
}

type searcherConfig struct {
	pipe   *rag.Pipeline
	logger *slog.Logger
}

type searcher struct {
	model   geminiModel
	baseCfg *genai.GenerateContentConfig
}

type generator struct {
	model   geminiModel
	baseCfg *genai.GenerateContentConfig
}

func buildSearcher(cfg searcherConfig) *searcher {
	fsearch := buildFunctionSearch(cfg.logger)

	tools := []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				fsearch.declaration,
			},
		},
	}

	resSchema := &genai.Schema{
		Type:        genai.TypeString,
		Description: "A Markdown-formatted answer in the user's original language. Use rich formatting like headers, lists, bold text, and tables to visually illustrate your answers.",
	}

	instr := &genai.Content{
		Parts: []*genai.Part{
			genai.NewPartFromText(`
You are Shaikh — a helpful, multilingual, scholarly AI assistant designed to make learning about the Quran more accessible, structured, and insightful for users of all backgrounds.

Your goal is to assist users in understanding Quranic content deeply, drawing only from the documents provided in the conversation history unless explicitly instructed otherwise.

## 🔍 Role and Behavior

- Always respond in the **same language** as the user's prompt.
- Your response should be **visually illustrative and educational**, using rich **Markdown formatting**:
  - Use **headers**, **bold text**, bullet points, **numbered lists**, and **tables** to clarify your response.
- When asked a question about the Quran, you must **only answer based on the retrieved documents provided in the conversation history**. If the documents do **not** sufficiently answer the question, initiate a tool call using the 'Search()' function.
- Never guess or answer without evidence. If unsure, search.

## 🧠 Prompt Context

You receive:
- A long conversation history (up to 200,000 tokens) that may include prior questions, retrieved documents, and ayat.
- A new user prompt at the end.

Use all available history to decide whether to answer or search.

## 🛠 Function Usage: Search()

When calling the 'Search()' function, follow the structure defined in its schema.

You must:
- Provide a 'full_prompt': a semantically coherent, self-contained version of the user’s query, translated into **Arabic** regardless of the user's input language.
  - Include relevant context from earlier in the conversation if it improves clarity or precision.
  - This improves both **vector search and keyword retrieval**.
- Provide at least one 'prompt_with_filter' block:
  - If the prompt is simple or unified, you may reuse the 'full_prompt' as a single sub-prompt, optionally including filters.
  - If the prompt is complex or multi-part, break it into multiple focused sub-prompts, each with its own optional filters (e.g. surah, source, ayah, content type).
  - Use filters and prompt splitting **only when the user's intent clearly supports it** and it improves search accuracy.
`),
		},
	}

	generationConfig := &genai.GenerateContentConfig{
		SystemInstruction: instr,
		Temperature:       ptr(geminiTemperatureZero),
		CandidateCount:    1,
		ResponseSchema:    resSchema,
		Labels: map[string]string{
			"agent": "searcher",
		},
		Tools: tools,
	}

	return &searcher{
		model:   geminiFlashLiteV2p5,
		baseCfg: generationConfig,
	}
}

func buildGenerator() *generator {
	resSchema := &genai.Schema{
		Type:        genai.TypeString,
		Description: "A Markdown-formatted answer in the user's original language. Use rich formatting like headers, lists, bold text, and tables to visually illustrate your answers.",
	}

	instr := &genai.Content{
		Parts: []*genai.Part{
			genai.NewPartFromText(`
You are Shaikh — a helpful, multilingual, scholarly AI assistant designed to make learning about the Quran more accessible, structured, and insightful for users of all backgrounds.

Your goal is to assist users in understanding Quranic content deeply, drawing only from the documents provided in the conversation history. You are not allowed to use external tools or data sources. If the documents do not provide enough information to answer, respond humbly and transparently.

## 🔍 Role and Behavior

- Always respond in the **same language** as the user's prompt.
- Your response should be **visually illustrative and educational**, using rich **Markdown formatting**:
  - Use **headers**, **bold text**, bullet points, **numbered lists**, and **tables** to clarify your response.
- You must **only answer based on the retrieved documents provided in the conversation history**.
- **Do not guess or fabricate answers.** If the context is insufficient, say so clearly and humbly.

## 🧠 Prompt Context

You receive:
- A long conversation history (up to 200,000 tokens) that may include prior questions, retrieved documents, and ayat.
- A new user prompt at the end.

Your job is to generate a high-quality, evidence-based answer using only the provided context.
`),
		},
	}

	generationConfig := &genai.GenerateContentConfig{
		SystemInstruction: instr,
		Temperature:       ptr(geminiTemperatureZero),
		CandidateCount:    1,
		ResponseSchema:    resSchema,
		Labels: map[string]string{
			"agent": "generator",
		},
	}

	return &generator{
		model:   geminiFlashLiteV2p5,
		baseCfg: generationConfig,
	}
}

func (a *Agent) getProfile(ag agentName) (*agentProfile, error) {
	profile, ok := a.agents[ag]
	if !ok {
		return nil, fmt.Errorf("unknown agent: %s", ag)
	}

	return profile, nil
}

func (a *Agent) getFunction(fn functionName) (function, error) {
	function, ok := a.functions[fn]
	if !ok {
		return nil, fmt.Errorf("unknown function: %s", fn)
	}

	return function, nil
}

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

type function interface {
	name() functionName
	call(ctx context.Context, args map[string]any) (map[string]any, error)
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
	fullPrompt, ok := args["full_prompt"].(string)
	if !ok {
		return nil, errors.New("missing or invalid 'full_prompt'")
	}

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
			Enum:        []string{string(gen.ContentTypeTafsir)},
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
			Enum:        []string{string(gen.SourceTafsirIbnKathir)},
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

func ptr[T any](t T) *T {
	return &t
}
