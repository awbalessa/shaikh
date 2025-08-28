package dom

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID          int32
	Source      Source
	Content     string
	SurahNumber SurahNumber
	AyahNumber  AyahNumber
}

type Chunk struct {
	Document
	ParentID int32
}

type User struct {
	ID    uuid.UUID
	Email string
}

type Session struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	LastAccessed time.Time
	Summary      *string
}

type MsgMeta struct {
	ID                int32
	SessionID         uuid.UUID
	UserID            uuid.UUID
	Model             LargeLanguageModel
	Turn              int32
	TotalInputTokens  *int32
	TotalOutputTokens *int32
	Content           *string
	FunctionName      *string
	FunctionCall      json.RawMessage
	FunctionResponse  json.RawMessage
}

type Message interface {
	Role() MessageRole
	Meta() *MsgMeta
}

type UserMessage struct {
	MsgMeta
	MsgContent string
}

func (m *UserMessage) Role() MessageRole { return UserRole }
func (m *UserMessage) Meta() *MsgMeta    { return &m.MsgMeta }

type ModelMessage struct {
	MsgMeta
	MsgContent string
}

func (m *ModelMessage) Role() MessageRole { return ModelRole }
func (m *ModelMessage) Meta() *MsgMeta    { return &m.MsgMeta }

type FunctionMessage struct {
	MsgMeta
	FunctionName     string
	FunctionCall     json.RawMessage
	FunctionResponse json.RawMessage
}

func (m *FunctionMessage) Role() MessageRole { return FunctionRole }
func (m *FunctionMessage) Meta() *MsgMeta    { return &m.MsgMeta }

type Memory struct {
	ID        int32
	UserID    uuid.UUID
	UpdatedAt time.Time
	Content   string
}

type Embedder interface {
	EmbedQueries(ctx context.Context, queries []string) ([]Vector, error)
}

type Rank struct {
	Index     int32
	Relevance float64
}

type Reranker interface {
	RerankDocuments(
		ctx context.Context,
		query string,
		documents []string,
		topk TopK,
	) ([]Rank, error)
}

type SemanticSearcher interface {
	ParallelSemanticSearch(
		ctx context.Context,
		queries []FullQueryContext,
		topk int,
	) ([][]Chunk, error)
	SemanticSearch(
		ctx context.Context,
		vector VectorWithLabel,
		topk int,
	) ([]Chunk, error)
}

type LexicalSearcher interface {
	ParallelLexicalSearch(
		ctx context.Context,
		queries []FullQueryContext,
		topk int,
	) ([][]Chunk, error)
	LexicalSearch(
		ctx context.Context,
		query QueryWithFilter,
		topk int,
	) ([]Chunk, error)
}

type Searcher interface {
	SemanticSearcher
	LexicalSearcher
}

type Cache interface {
	Set(ctx context.Context, key string, value []byte, expr time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
}

type LLMRole string

const (
	LLMUserRole  LLMRole = "user"
	LLMModelRole LLMRole = "model"
)

type LLMFunctionCall struct {
	Name string
	Args map[string]any
}

type LLMFunctionResponse struct {
	Name    string
	Content map[string]any
}

type LLMPart struct {
	Text             string
	FunctionCall     *LLMFunctionCall
	FunctionResponse *LLMFunctionResponse
}

type LLMContent struct {
	Role  LLMRole
	Parts []*LLMPart
}

type LLMSchemaType string

const (
	SchemaString  LLMSchemaType = "STRING"
	SchemaInteger LLMSchemaType = "INTEGER"
	SchemaNumber  LLMSchemaType = "NUMBER"
	SchemaBoolean LLMSchemaType = "BOOLEAN"
	SchemaArray   LLMSchemaType = "ARRAY"
	SchemaObject  LLMSchemaType = "OBJECT"
)

type LLMSchema struct {
	Title       string        `json:"title,omitempty"`
	Description string        `json:"description,omitempty"`
	Type        LLMSchemaType `json:"type,omitempty"`

	Format string   `json:"format,omitempty"`
	Enum   []string `json:"enum,omitempty"`

	Required   []string              `json:"required,omitempty"`
	Properties map[string]*LLMSchema `json:"properties,omitempty"`

	Items    *LLMSchema `json:"items,omitempty"`
	MinItems *int64     `json:"minItems,omitempty"`
	MaxItems *int64     `json:"maxItems,omitempty"`

	Minimum *float64 `json:"minimum,omitempty"`
	Maximum *float64 `json:"maximum,omitempty"`

	Example any `json:"example,omitempty"`
}

type LLMFunctionDecl struct {
	Name        string
	Description string
	Parameters  *LLMSchema
}

type LLMGenConfig struct {
	SystemInstructions *LLMContent
	Temperature        float32
	CandidateCount     int32
	Tools              []*LLMFunctionDecl
	ResponseMimeType   string
	ResponseSchema     *LLMSchema
}

type LLMCountConfig struct {
	System *LLMContent
	Tools  []*LLMFunctionDecl
}

func Ptr[T any](v T) *T { return &v }

func StringEnum(options ...string) *LLMSchema {
	return &LLMSchema{
		Type: SchemaString,
		Enum: options,
	}
}

func ArrayOf(item *LLMSchema, min, max *int64) *LLMSchema {
	return &LLMSchema{
		Type:     SchemaArray,
		Items:    item,
		MinItems: min,
		MaxItems: max,
	}
}

func ObjectWith(props map[string]*LLMSchema, required ...string) *LLMSchema {
	return &LLMSchema{
		Type:       SchemaObject,
		Properties: props,
		Required:   required,
	}
}

func IntegerRange(min, max *float64) *LLMSchema {
	return &LLMSchema{
		Type:    SchemaInteger,
		Minimum: min,
		Maximum: max,
	}
}

func WithDocs(title *string, description *string, s *LLMSchema) *LLMSchema {
	if title != nil {
		s.Title = *title
	}

	if description != nil {
		s.Description = *description
	}

	return s
}

type LLM interface {
	Stream(
		ctx context.Context,
		model string,
		window []*LLMContent,
		cfg *LLMGenConfig,
		yield func(*LLMPart, error) bool,
	) *LLMGenResult

	CountTokens(
		ctx context.Context,
		model string,
		window []*LLMContent,
		cfg *LLMCountConfig,
	) (int32, error)
}

type AgentName string

const (
	Caller    AgentName = "Caller"
	Generator AgentName = "Generator"
)

type AgentProfile struct {
	Model  string
	Config *LLMGenConfig
}

type LLMFunctionName string

const (
	FunctionSearch LLMFunctionName = "Search()"
)

type LLMFunction interface {
	Call(ctx context.Context, args map[string]any) (map[string]any, error)
}

type TokenUsage struct {
	InputTokens  int32
	OutputTokens int32
}

type FinishReason string

const (
	FinishReasonUnspecified           FinishReason = "FINISH_REASON_UNSPECIFIED"
	FinishReasonStop                  FinishReason = "STOP"
	FinishReasonMaxTokens             FinishReason = "MAX_TOKENS"
	FinishReasonSafety                FinishReason = "SAFETY"
	FinishReasonRecitation            FinishReason = "RECITATION"
	FinishReasonLanguage              FinishReason = "LANGUAGE"
	FinishReasonOther                 FinishReason = "OTHER"
	FinishReasonBlocklist             FinishReason = "BLOCKLIST"
	FinishReasonProhibitedContent     FinishReason = "PROHIBITED_CONTENT"
	FinishReasonSPII                  FinishReason = "SPII"
	FinishReasonMalformedFunctionCall FinishReason = "MALFORMED_FUNCTION_CALL"
	FinishReasonImageSafety           FinishReason = "IMAGE_SAFETY"
	FinishReasonUnexpectedToolCall    FinishReason = "UNEXPECTED_TOOL_CALL"
)

type LLMGenResult struct {
	Output        *ModelOutput
	Usage         *TokenUsage
	FinishReason  FinishReason
	FinishMessage string
}

type Agent interface {
	Generate(
		ctx context.Context,
		name AgentName,
		win []*LLMContent,
	) iter.Seq2[*LLMPart, error]
	GenerateWithYield(
		ctx context.Context,
		name AgentName,
		win []*LLMContent,
		yield func(*LLMPart, error) bool,
	) *LLMGenResult
	BuildContextWindow(
		ctx context.Context,
		name AgentName,
		cw *ContextWindow,
	) ([]*LLMContent, error)
}

var (
	ErrAgentDoesNotExist = errors.New("agent does not exist")
)

type AgentStruct struct {
	Agents map[AgentName]AgentProfile
	LLM    LLM
}

func NewAgent()

func (a *AgentStruct) Generate(
	ctx context.Context,
	name AgentName,
	win []*LLMContent,
) iter.Seq2[*LLMPart, error] {
	return iter.Seq2[*LLMPart, error](func(yield func(*LLMPart, error) bool) {
		prof, ok := a.Agents[name]
		if !ok {
			yield(nil, ErrAgentDoesNotExist)
			return
		}

		a.LLM.Stream(ctx, prof.Model, win, prof.Config, yield)
	})
}

func (a *AgentStruct) GenerateWithYield(
	ctx context.Context,
	name AgentName,
	win []*LLMContent,
	yield func(*LLMPart, error) bool,
) *LLMGenResult {
	prof, ok := a.Agents[name]
	if !ok {
		yield(nil, ErrAgentDoesNotExist)
		return nil
	}

	return a.LLM.Stream(ctx, prof.Model, win, prof.Config, yield)
}

func BuildCaller() *AgentProfile {
	fnSearch := BuildFunctionSearch()

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

func BuildFunctionSearch() *LLMFunctionDecl {
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

const (
	TokenLimit int32 = 200_000
)

func (a *AgentStruct) BuildContextWindow(
	ctx context.Context,
	name AgentName,
	cw *ContextWindow,
	now time.Time,
) ([]*LLMContent, error) {
	prof, ok := a.Agents[name]
	if !ok {
		return nil, ErrAgentDoesNotExist
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
			partText := fmt.Sprintf("Last Accessed: %s\nSummary: %s",
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

		if inter.Input.Text != "" {
			t = append(t, &LLMContent{
				Role:  LLMUserRole,
				Parts: []*LLMPart{{Text: inter.Input.Text}},
			})
		}

		if inter.Output.FunctionCall != nil {
			t = append(t, &LLMContent{
				Role:  LLMModelRole,
				Parts: []*LLMPart{{FunctionCall: inter.Output.FunctionCall}},
			})
		}

		if inter.Input.FunctionResponse != nil {
			t = append(t, &LLMContent{
				Role:  LLMUserRole,
				Parts: []*LLMPart{{FunctionResponse: inter.Input.FunctionResponse}},
			})
		}

		if inter.Output.Text != "" {
			t = append(t, &LLMContent{
				Role:  LLMModelRole,
				Parts: []*LLMPart{{Text: inter.Output.Text}},
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
			return nil, fmt.Errorf("failed to build context window: %w", err)
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

type PubOptions struct {
	MsgID string
}

type PubAck struct {
	Stream string
	Seq    uint64
}

type PubMsgMetadata struct {
	Stream       string
	Consumer     string
	NumDelivered uint64
	Timestamp    time.Time
}

type PubMsg interface {
	Data() []byte
	Subject() string
	Ack() error
	Nak() error
	Term() error
	InProgress() error
	Metadata() (PubMsgMetadata, error)
}

type PubSubRetentionPolicy int
type PubSubStorageType int

const (
	WorkQueue PubSubRetentionPolicy = iota
	LimitsBased
)

const (
	FileStorage PubSubStorageType = iota
)

type PubSubStreamConfig struct {
	Name       string
	Subjects   []string
	Retention  PubSubRetentionPolicy
	MaxMsgs    int64
	MaxAge     time.Duration
	Storage    PubSubStorageType
	Replicas   int
	Duplicates time.Duration
}

type PubSub interface {
	CreateStream(ctx context.Context, cfg PubSubStreamConfig) error
	CreateConsumer(ctx context.Context, stream string, cfg PubSubConsumerConfig) (PubSubConsumer, error)
}

type Publisher interface {
	Publish(ctx context.Context, subject string, data []byte, opts *PubOptions) (*PubAck, error)
}

type PubSubDeliverPolicy int
type PubSubAckPolicy int
type PubSubReplayPolicy int

const (
	DeliverAll    PubSubDeliverPolicy = iota
	AckExplicit   PubSubAckPolicy     = iota
	ReplayInstant PubSubReplayPolicy  = iota
)

type PubSubConsumerConfig struct {
	Name              string
	Durable           bool
	InactiveThreshold time.Duration
	DeliverPolicy     PubSubDeliverPolicy
	AckPolicy         PubSubAckPolicy
	AckWait           time.Duration
	MaxDeliver        int
	BackOff           []time.Duration
	FilterSubjects    []string
	ReplayPolicy      PubSubReplayPolicy
	MaxRequestBatch   int
	MaxRequestExpires time.Duration
}

type PubSubConsumer interface {
	Fetch(batch int) ([]PubMsg, error)
	Messages(ctx context.Context) (<-chan PubMsg, error)
}

type MemoryRepo interface {
	GetMemoriesByUserID(
		ctx context.Context,
		userID uuid.UUID,
		numberOfMemories int32,
	) ([]Memory, error)
}

type SessionRepo interface {
	GetSessionsByUserID(
		ctx context.Context,
		userID uuid.UUID,
		numberOfSessions int32,
	) ([]Session, error)
}

type MessageRepo interface {
	CreateMessage(
		ctx context.Context,
		msg Message,
	) (Message, error)
	GetMessagesBySessionIDOrdered(
		ctx context.Context,
		sessionID uuid.UUID,
	) ([]Message, error)
}

type Tx interface {
	Get(repo any) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type UnitOfWork interface {
	Begin(ctx context.Context) (Tx, error)
}
