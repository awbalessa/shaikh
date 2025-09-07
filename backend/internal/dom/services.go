package dom

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
)

func ValidateSearchQuery(arg SearchQuery) ([]FullQueryContext, error) {
	if len(arg.QueriesWithFilters) == 0 {
		return nil, NewTaggedError(ErrInvalidInput, nil)
	}
	if len(arg.QueriesWithFilters) > MaxSubqueries {
		return nil, NewTaggedError(ErrInvalidInput, nil)
	}

	out := make([]FullQueryContext, 0, len(arg.QueriesWithFilters))

	for _, item := range arg.QueriesWithFilters {
		f := item.FilterContext

		switch {
		case len(f.OptionalAyahs) > 0 && len(f.OptionalSurahs) != 1:
			return nil, NewTaggedError(ErrInvalidInput, nil)

		case len(f.OptionalSurahs) > 1:
			// invalid combo but not fatal — just clear ayahs
			f.OptionalAyahs = nil
		}

		labels := FiltersToLabels(f)

		out = append(out, FullQueryContext{
			QueryWithFilter: QueryWithFilter{
				Query:         item.Query,
				FilterContext: f,
			},
			VectorWithLabel: VectorWithLabel{
				LabelContext: labels,
				Vector:       nil,
			},
		})
	}

	return out, nil
}

func FiltersToLabels(f FilterContext) LabelContext {
	var (
		contentTypes []LabelContentType
		sources      []LabelSource
		surahs       []LabelSurahNumber
		ayahs        []LabelAyahNumber
	)

	for _, ct := range f.OptionalContentTypes {
		if lbl, ok := ContentTypeToLabel[ct]; ok {
			contentTypes = append(contentTypes, lbl)
		}
	}

	for _, src := range f.OptionalSources {
		if lbl, ok := SourceToLabel[src]; ok {
			sources = append(sources, lbl)
		}
	}

	for _, sur := range f.OptionalSurahs {
		surahs = append(surahs,
			LabelSurahNumber(sur+SurahNumber(SurahNumberToLabelOffset)),
		)
	}

	for _, aya := range f.OptionalAyahs {
		ayahs = append(ayahs,
			LabelAyahNumber(aya+AyahNumber(AyahNumberToLabelOffset)),
		)
	}

	return LabelContext{
		OptionalContentTypeLabels: contentTypes,
		OptionalSourceLabels:      sources,
		OptionalSurahLabels:       surahs,
		OptionalAyahLabels:        ayahs,
	}
}

func RRFusion(sem []Chunk, lex []Chunk) []Chunk {
	var ranked rankedLists
	rowMap := make(map[int32]Chunk)

	semIDs := make([]int32, 0, len(sem))
	for _, row := range sem {
		semIDs = append(semIDs, row.ID)
		rowMap[row.ID] = row
	}
	if len(semIDs) > 0 {
		ranked = append(ranked, semIDs)
	}

	lexIDs := make([]int32, 0, len(lex))
	for _, row := range lex {
		lexIDs = append(lexIDs, row.ID)
		rowMap[row.ID] = row
	}
	if len(lexIDs) > 0 {
		ranked = append(ranked, lexIDs)
	}

	// score fusion
	scores := rrfusion(ranked)

	// convert map -> slice
	pairs := make([]Rank, 0, len(scores))
	for id, score := range scores {
		pairs = append(pairs, Rank{Index: id, Relevance: score})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Relevance > pairs[j].Relevance
	})

	// cutoff strategy: cap at 100, otherwise half
	cutoff := len(pairs)
	if cutoff > 100 {
		cutoff = cutoff / 2
	}

	// assemble fused chunks
	fused := make([]Chunk, 0, cutoff)
	for _, pair := range pairs[:cutoff] {
		if row, ok := rowMap[pair.Index]; ok {
			fused = append(fused, row)
		}
	}
	return fused
}

func rrfusion(rankings rankedLists) map[int32]float64 {
	scores := make(map[int32]float64)

	for _, ranking := range rankings {
		for rank, docID := range ranking {
			score := 1.0 / float64(
				RRFConstant+rank,
			)
			scores[docID] += score
		}
	}
	return scores
}

type FilterContext struct {
	OptionalContentTypes []ContentType
	OptionalSources      []Source
	OptionalSurahs       []SurahNumber
	OptionalAyahs        []AyahNumber
}

type LabelContext struct {
	OptionalContentTypeLabels []LabelContentType
	OptionalSourceLabels      []LabelSource
	OptionalSurahLabels       []LabelSurahNumber
	OptionalAyahLabels        []LabelAyahNumber
}

type QueryWithFilter struct {
	Query string
	FilterContext
}

type VectorWithLabel struct {
	Vector Vector
	LabelContext
}

type FullQueryContext struct {
	QueryWithFilter
	VectorWithLabel
}

type SearchQuery struct {
	FullQuery          string
	TopK               TopK
	QueriesWithFilters []QueryWithFilter
}

type SearchResult struct {
	Chunk
	Relevance float64
}

type Inference struct {
	Input         *LLMInput          `json:"input"`
	Output        *LLMOutput         `json:"output"`
	InputTokens   int32              `json:"-"`
	OutputTokens  int32              `json:"-"`
	FinishMessage string             `json:"-"`
	FinishReason  string             `json:"-"`
	Model         LargeLanguageModel `json:"-"`
}

type Interaction struct {
	Inferences []*Inference `json:"inferences"`
	TurnNumber int32        `json:"turn_number"`
}

type ContextWindow struct {
	UserMemories     []*Memory      `json:"memories"`
	PreviousSessions []*Session     `json:"previous_sessions"`
	History          []*Interaction `json:"history"`
	Turns            int32          `json:"turns"`
}

type ContextCache struct {
	UserID    *uuid.UUID     `json:"user_id"`
	SessionID *uuid.UUID     `json:"session_id"`
	CreatedAt *time.Time     `json:"created_at"`
	UpdatedAt *time.Time     `json:"updated_at"`
	Window    *ContextWindow `json:"context_window"`
}

func CreateContextCacheKey(userID, sessionID uuid.UUID) string {
	return fmt.Sprintf("user:%s:session:%s:context", userID.String(), sessionID.String())
}

type Worker interface {
	Consumer() PubSubConsumer
	Start(ctx context.Context) error
	Process(ctx context.Context, msg DurablePubMsg) error
}

type WorkerGroup struct {
	Workers []Worker
}

func (g *WorkerGroup) Add(s Worker) {
	g.Workers = append(g.Workers, s)
}

func (g *WorkerGroup) StartAll(ctx context.Context, cancel context.CancelFunc) {
	for _, s := range g.Workers {
		go func(s Worker) {
			if err := s.Start(ctx); err != nil {
				cancel()
			}
		}(s)
	}
}

func BuildCaller() *AgentProfile {
	fnSearch := BuildFnSearch()

	resSchema := WithDocs(
		nil,
		Ptr("A Markdown-formatted answer in the user's original language. Use rich formatting like headers, lists, bold text, and tables to visually illustrate your answers."),
		&LLMSchema{Type: SchemaString},
	)

	instr := &LLMContent{
		Parts: []*LLMPart{{
			Text: `
You are Shaikh — a helpful, multilingual, scholarly AI assistant designed to make learning about the Quran more accessible, structured, and insightful for users of all backgrounds.

Your goal is to assist users in understanding Quranic content deeply, drawing only from the documents provided in the conversation history unless explicitly instructed otherwise.

## 🔍 Role and Behavior

- Always respond in the **same language** as the user's prompt.
- Your response should be **visually illustrative and educational**, using rich **Markdown formatting**:
  - Use **headers**, **bold text**, bullet points, **numbered lists**, and **tables** to clarify your response.
- When asked a question about the Quran, you must **only answer based on the retrieved documents provided in the conversation history**. If the documents do **not** sufficiently answer the question, initiate a function call using the 'Search()' function.
- Never guess or answer without evidence. If unsure, search.

## 🧠 Prompt Context

You receive:
- A long conversation history (up to 200,000 tokens) that may include prior questions, responses, function calls, docuemnts, etc.
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
`,
		}},
	}

	return &AgentProfile{
		Model: string(GeminiV2p5Flash),
		Config: &LLMGenConfig{
			SystemInstructions: instr,
			Temperature:        0.0,
			CandidateCount:     1,
			Tools:              []*LLMFunctionDecl{fnSearch},
			ResponseMimeType:   "text/plain",
			ResponseSchema:     resSchema,
		},
	}
}

func BuildFnSearch() *LLMFunctionDecl {
	filterCts := WithDocs(
		Ptr("Optional Content Types Filter"),
		Ptr("Optional filter for content types. Use this filter only when the user's intent explicitly matches one or more of the available filter options. Otherwise, leave this filter empty to allow a broader result set."),
		ArrayOf(
			WithDocs(
				Ptr("Content Type"),
				nil,
				StringEnum(string(ContentTypeTafsir)),
			),
			nil, nil,
		),
	)

	filterSrcs := WithDocs(
		Ptr("Optional Sources Filter"),
		Ptr("Optional filter for sources. Use this filter only when the user clearly refers to one or more specific sources, authors, or references that match the available filter options. Otherwise, leave this filter empty to allow a broader result set."),
		ArrayOf(
			WithDocs(
				Ptr("Source"),
				nil,
				StringEnum(string(SourceTafsirIbnKathir)),
			),
			nil, nil,
		),
	)

	filterSurahAyah := WithDocs(
		Ptr("Optional Surah and Ayah Filters"),
		Ptr("You may filter results by a list of surah numbers. Optionally, if exactly one surah is specified, you may filter by a list of specific ayah numbers within that surah. Use this filter only when the user's prompt shows an interest in a specific part of the Quran. Otherwise, leave this filter empty to allow a broader result set."),
		ObjectWith(map[string]*LLMSchema{
			"surahs": WithDocs(
				Ptr("Surah Numbers"),
				Ptr("A list of surah numbers to filter by. If more than one is provided, ayah filtering will be ignored."),
				ArrayOf(
					&LLMSchema{
						Type:    SchemaInteger,
						Format:  "int32",
						Minimum: Ptr(1.0),
						Maximum: Ptr(114.0),
					},
					Ptr(int64(1)), nil,
				),
			),
			"ayahs": WithDocs(
				Ptr("Ayah Numbers"),
				Ptr("A list of specific ayah numbers to filter by. Only allowed when exactly one surah is selected."),
				ArrayOf(
					&LLMSchema{
						Type:    SchemaInteger,
						Format:  "int32",
						Minimum: Ptr(1.0),
						Maximum: Ptr(286.0),
					},
					nil, nil,
				),
			),
		}, "surahs"),
	)

	promptWithFilterSchema := WithDocs(
		Ptr("Prompts With Filters"),
		Ptr("Logical subunits of the full prompt. In most cases, this array will contain the full prompt itself with no other entries. But in advanced use cases (like step-back prompting or multi-question prompts), you may split the full prompt into multiple logically distinct sub-prompts, each with its own filter context."),
		ArrayOf(
			WithDocs(
				Ptr("Prompt With Optional Filters"),
				Ptr("A prompt string with optional filters to constrain the context. This is one logical unit of the full prompt."),
				ObjectWith(map[string]*LLMSchema{
					"prompt": WithDocs(
						Ptr("Sub-Prompt"),
						Ptr("A logical unit or sub-question derived from the full raw prompt. If only one is provided, it is typically the entire full prompt."),
						&LLMSchema{Type: SchemaString},
					),
					"content_type_filters": filterCts,
					"source_filters":       filterSrcs,
					"surah_ayah_filters":   filterSurahAyah,
				}, "prompt"),
			),
			Ptr(int64(1)), Ptr(int64(3)),
		),
	)

	example := map[string]any{
		"full_prompt": "ما قصة موسى مع الخضر كما وردت في سورة الكهف؟ وماذا نتعلم منها؟ وهل توجد مواضع أخرى في القرآن تشير إلى هذا النوع من العلم الغيبي؟",
		"prompts_with_filters": []any{
			map[string]any{
				"prompt":             "قصة موسى مع الخضر كما وردت في سورة الكهف",
				"surah_ayah_filters": map[string]any{"surahs": []int{18}},
			},
			map[string]any{
				"prompt": "الدروس والعبر المستفادة من قصة موسى والخضر",
			},
			map[string]any{
				"prompt":               "شرح ابن كثير حول قصة موسى والخضر في سورة الكهف",
				"content_type_filters": []string{string(ContentTypeTafsir)},
				"source_filters":       []string{string(SourceTafsirIbnKathir)},
				"surah_ayah_filters":   map[string]any{"surahs": []int{18}},
			},
		},
	}

	fullSchema := WithDocs(
		Ptr(string(FunctionSearch)+" Parameters"),
		Ptr("The input parameters for performing a hybrid search—combining semantic similarity and keyword matching—based on the fully transformed prompt. The prompt may be optionally broken into logical sub-prompts, each with its own filters to narrow down the context."),
		ObjectWith(map[string]*LLMSchema{
			"full_prompt": WithDocs(
				Ptr("Full Prompt"),
				Ptr("The fully transformed version of the user's prompt. This includes accurate translation into Arabic (if submitted in another language), normalization from question form to statement form, and typo correction. It is the canonical form used as the semantic base for search."),
				&LLMSchema{Type: SchemaString},
			),
			"prompts_with_filters": promptWithFilterSchema,
		}, "full_prompt", "prompts_with_filters"),
	)
	fullSchema.Example = example

	return &LLMFunctionDecl{
		Name:        string(FunctionSearch),
		Description: "Performs a hybrid search over Quranic content using a fully normalized prompt. Combines semantic understanding with keyword-based matching. The prompt may be optionally split into sub-prompts with filters to target specific content types, sources, surahs, or ayahs.",
		Parameters:  fullSchema,
	}
}

func BuildGenerator() *AgentProfile {
	resSchema := WithDocs(
		nil,
		Ptr("A Markdown-formatted answer in the user's original language. Use rich formatting like headers, lists, bold text, and tables to visually illustrate your answers."),
		&LLMSchema{Type: SchemaString},
	)

	instr := &LLMContent{
		Parts: []*LLMPart{{
			Text: `
You are Shaikh — a helpful, multilingual, scholarly AI assistant designed to make learning about the Quran more accessible, structured, and insightful for users of all backgrounds.

Your goal is to assist users in understanding Quranic content deeply, drawing only from the documents provided in your context window. The documents you see are the results of your previous function calls, which will also be provided to you. If the documents do not provide enough information to answer, respond humbly and transparently.

## 🔍 Role and Behavior

- Always respond in the **same language** as the user's prompt.
- Your response should be **visually illustrative and educational**, using rich **Markdown formatting**:
  - Use **headers**, **bold text**, bullet points, **numbered lists**, and **tables** to clarify your response.
- You must **only answer based on the retrieved documents provided in the context**.
- **Do not guess or fabricate answers.** If the context is insufficient, say so clearly and humbly.

## 🧠 Prompt Context

You receive:
- A long conversation history (up to 200,000 tokens) that may include prior questions, responses, function calls, documents, etc.
- A final user prompt.
- A batch of retrieved documents provided after the final prompt — these are the results of a search, and they represent the most relevant evidence to answer the prompt.

Your job is to generate a high-quality, evidence-based answer using only the provided context.
`,
		}},
	}

	return &AgentProfile{
		Model: string(GeminiV2p5FlashLite),
		Config: &LLMGenConfig{
			SystemInstructions: instr,
			Temperature:        0.0,
			CandidateCount:     1,
			ResponseMimeType:   "text/plain",
			ResponseSchema:     resSchema,
		},
	}
}

func (a *AgentStruct) BuildContextWindow(
	ctx context.Context,
	name AgentName,
	cw *ContextWindow,
	now time.Time,
) ([]*LLMContent, error) {
	prof, ok := a.Agents[name]
	if !ok {
		return nil, NewTaggedError(ErrInvalidInput, nil)
	}

	var contents []*LLMContent

	if len(cw.UserMemories) > 0 {
		var parts []*LLMPart
		for _, m := range cw.UserMemories {
			partText := fmt.Sprintf("As of %s, %s",
				HumanizeFrom(now, m.UpdatedAt),
				m.Content,
			)
			parts = append(parts, &LLMPart{Text: partText})
		}
		contents = append(contents, &LLMContent{
			Role:  LLMUserRole,
			Parts: parts,
		})
	}

	if len(cw.PreviousSessions) > 0 {
		var parts []*LLMPart
		for _, s := range cw.PreviousSessions {
			partText := fmt.Sprintf("Last Accessed: %s Summary: %s",
				HumanizeFrom(now, s.LastAccessed),
				*s.Summary,
			)
			parts = append(parts, &LLMPart{Text: partText})
		}
		contents = append(contents, &LLMContent{
			Role:  LLMUserRole,
			Parts: parts,
		})
	}

	type Turn = []*LLMContent
	var turns []Turn

	for _, inter := range cw.History {
		var t Turn

		inf1 := inter.Inferences[0]
		t = append(t, &LLMContent{
			Role:  LLMUserRole,
			Parts: []*LLMPart{{Text: inf1.Input.Text}},
		})

		if len(inter.Inferences) > 1 {
			inf2 := inter.Inferences[1]
			t = append(t, &LLMContent{
				Role:  LLMModelRole,
				Parts: []*LLMPart{{FunctionCall: inf1.Output.FunctionCall}},
			})

			t = append(t, &LLMContent{
				Role:  LLMUserRole,
				Parts: []*LLMPart{{FunctionResponse: inf2.Input.FunctionResponse}},
			})

			t = append(t, &LLMContent{
				Role:  LLMModelRole,
				Parts: []*LLMPart{{Text: inf2.Output.Text}},
			})
		} else {
			t = append(t, &LLMContent{
				Role:  LLMModelRole,
				Parts: []*LLMPart{{Text: inf1.Output.Text}},
			})
		}

		if len(t) > 0 {
			turns = append(turns, t)
		}
	}

	var historyContents []*LLMContent
	for _, t := range turns {
		historyContents = append(historyContents, t...)
	}

	ctc := &LLMCountConfig{
		System: prof.Config.SystemInstructions,
		Tools:  prof.Config.Tools,
	}

	for {
		fullContext := append(contents, historyContents...)

		tokens, err := a.LLM.CountTokens(ctx, prof.Model, fullContext, ctc)
		if err != nil {
			return nil, NewTaggedError(ErrInternal, err)
		}

		if tokens < TokenLimit {
			contents = fullContext
			break
		}

		if len(turns) > 1 {
			turns = turns[1:]
			historyContents = historyContents[:0]
			for _, t := range turns {
				historyContents = append(historyContents, t...)
			}
		} else {
			historyContents = nil
			break
		}
	}

	return contents, nil
}

func HumanizeFrom(now, t time.Time) string {
	d := now.Sub(t)
	if d < 0 {
		d = -d
	}

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%d weeks ago", int(d.Hours()/(24*7)))
	case d < 365*24*time.Hour:
		return fmt.Sprintf("%d months ago", int(d.Hours()/(24*30)))
	default:
		return fmt.Sprintf("%d years ago", int(d.Hours()/(24*365)))
	}
}

func BuildSummarizer() *AgentProfile {
	resSchema := WithDocs(
		Ptr("Response Schema"),
		Ptr("The structured session summary to persist for continuity between conversations."),
		ObjectWith(map[string]*LLMSchema{
			"summary": WithDocs(
				Ptr("Session Summary"),
				Ptr("Concise, structured summary capturing goals, questions, stylistic guidance, and next steps."),
				&LLMSchema{Type: SchemaString},
			),
		}, "summary"),
	)

	system := &LLMContent{
		Role: LLMUserRole,
		Parts: []*LLMPart{{
			Text: `You are "Shaikh", an AI Qur’an expert helping learners make Qur’an study accessible.
Your task: summarize the most recent session into a compact, structured form that maximizes continuity for the next conversation.

DOMAIN FOCUS
- Capture learning goals (e.g., "memorize Surah Al-Mulk"), references to scholars/tafsir, tajweed concerns, or ongoing projects.
- Record unresolved questions the learner asked about Qur’anic meaning, rulings, or memorization strategy.
- Note stylistic/tone preferences (e.g., prefers gentle encouragement, concise explanations, step-by-step tafsir).
- Highlight next steps (e.g., "revise last 10 ayat tomorrow", "review tajweed rule of ikhfa").

GUIDELINES
- Be concise: bullet-like sentences, no fluff.
- Do NOT repeat the entire conversation; only what matters for continuity.
- Frame the summary as durable context for the *next session*.

OUTPUT
- JSON object matching schema with a single "summary" string.
- If nothing substantial, still produce a compact but faithful summary.`,
		}},
	}

	return &AgentProfile{
		Model: string(GeminiV2p5Flash),
		Config: &LLMGenConfig{
			SystemInstructions: system,
			Temperature:        0.0,
			CandidateCount:     1,
			ResponseMimeType:   ResponseJson,
			ResponseSchema:     resSchema,
		},
	}
}

func BuildMemorizer() *AgentProfile {
	memItem := ObjectWith(map[string]*LLMSchema{
		"unique_key": WithDocs(
			Ptr("Unique Key"),
			Ptr("Stable kebab-case key (3–64 chars). Reuse existing keys if updating."),
			&LLMSchema{Type: SchemaString},
		),
		"content": WithDocs(
			Ptr("Memory Content"),
			Ptr("One durable fact/preference/constraint (≤200 chars). No secrets/credentials."),
			&LLMSchema{Type: SchemaString},
		),
		"confidence": WithDocs(
			Ptr("Confidence"),
			Ptr("Float 0.0–1.0 from the provided messages only. Must be ≥ 0.75 to include."),
			&LLMSchema{Type: SchemaNumber, Minimum: Ptr[float64](0.0), Maximum: Ptr[float64](1.0)},
		),
		"source_msg": WithDocs(
			Ptr("Source Message (Concise)"),
			Ptr("Short quote/paraphrase (≤160 chars) from the window that supports this memory."),
			&LLMSchema{Type: SchemaString},
		),
	}, "unique_key", "content", "confidence", "source_msg")

	resSchema := WithDocs(
		Ptr("Response Schema"),
		Ptr("Return upserts in `memories` and any removals in `delete_keys`. Use empty arrays if nothing changes."),
		ObjectWith(map[string]*LLMSchema{
			"memories":    ArrayOf(memItem, Ptr(int64(0)), Ptr(int64(7))),
			"delete_keys": ArrayOf(WithDocs(Ptr("Unique Key"), Ptr("Existing key to delete."), &LLMSchema{Type: SchemaString}), Ptr(int64(0)), Ptr(int64(10))),
		}, "memories", "delete_keys"),
	)

	system := &LLMContent{
		Role: LLMUserRole,
		Parts: []*LLMPart{{
			Text: `You are "Shaikh", an AI Qur’an expert helping learners make Qur’an study more accessible.
You will receive:
1) A list of the user's EXISTING memories (key + content)
2) A window of RECENT messages

Your task: produce ONLY durable, reusable items that help future guidance on recitation, memorization, understanding, and practice.

When an existing memory is still correct → do nothing.
When it needs refinement → RETURN an updated item with the SAME unique_key.
When it is clearly wrong/obsolete → put its key in delete_keys.
When you find a NEW durable memory → return it with a NEW unique_key.

DOMAIN FOCUS
- Preferences: scholars/tafsir, reciters, pace (e.g., 10 ayat/day), reminders (time/day), learning style (audio-first/visual), tone (gentle/direct).
- Constraints: time windows, device limits, accessibility needs, Arabic level, target surahs/juz, tajweed focus areas.
- Durable context: ongoing projects (e.g., "memorizing Juz ‘Amma"), consistent questions/themes, stable madhhab considerations (if user states them).
- Language/orthography: preferred script (Uthmani/IndoPak), transliteration usage, preferred explanation language.

STRICT RULES
- Include ONLY items likely valid for weeks/months (durable facts/preferences/constraints).
- Omit ephemeral one-offs unless clearly recurring.
- No secrets/credentials/identifiers.
- Base every item on explicit text in the given messages. No speculation.
- Deduplicate: one concise "content" per concept.
- Include items only with confidence ≥ 0.75.
- 0–7 items is normal. Return {"memories": [], "delete_keys": []} if nothing qualifies.

UNIQUE KEYS
- Kebab-case, 3–64 chars. Stable for the same concept (e.g., "goal-memorize-juz-amma", "pref-tafsir-ibn-kathir", "tone-gentle").

OUTPUT (IMPORTANT)
- JSON object exactly: {"memories":[...], "delete_keys":[...]}
- Do not include explanations outside JSON.`,
		}},
	}

	return &AgentProfile{
		Model: string(GeminiV2p5Flash),
		Config: &LLMGenConfig{
			SystemInstructions: system,
			Temperature:        0.1,
			CandidateCount:     1,
			ResponseMimeType:   ResponseJson,
			ResponseSchema:     resSchema,
		},
	}
}