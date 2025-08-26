package dom

import (
	"context"
	"errors"
	"iter"
	"time"

	"github.com/google/uuid"
)

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
	Parts []LLMPart
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
	Tools              []LLMFunctionDecl
	ResponseMimeType   string
	ResponseSchema     *LLMSchema
}

type LLMCountConfig struct {
	System *LLMContent
	Tools  []LLMFunctionDecl
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
	Name() LLMFunctionName
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
	FinalOutput   *ModelOutput
	Usage         *TokenUsage
	FinishReason  FinishReason
	FinishMessage string
}

type Agent interface {
	Generate(
		ctx context.Context,
		name AgentName,
		cw []*LLMContent,
	) iter.Seq2[*LLMPart, error]
	GenerateWithYield(
		ctx context.Context,
		name AgentName,
		cw []*LLMContent,
		yield func(*LLMPart, error) bool,
	) *LLMGenResult
}

var (
	ErrAgentDoesNotExist = errors.New("agent does not exist")
)

type AgentStruct struct {
	Agents map[AgentName]AgentProfile
	LLM    LLM
}

func (a *AgentStruct) Generate(
	ctx context.Context,
	name AgentName,
	cw []*LLMContent,
) iter.Seq2[*LLMPart, error] {
	return iter.Seq2[*LLMPart, error](func(yield func(*LLMPart, error) bool) {
		prof, ok := a.Agents[name]
		if !ok {
			yield(nil, ErrAgentDoesNotExist)
			return
		}

		a.LLM.Stream(ctx, prof.Model, cw, prof.Config, yield)
	})
}

func (a *AgentStruct) GenerateWithYield(
	ctx context.Context,
	name AgentName,
	cw []*LLMContent,
	yield func(*LLMPart, error) bool,
) *LLMGenResult {
	prof, ok := a.Agents[name]
	if !ok {
		yield(nil, ErrAgentDoesNotExist)
		return nil
	}

	return a.LLM.Stream(ctx, prof.Model, cw, prof.Config, yield)
}

type PubOptions struct {
	MsgID string
}

type PubAck struct {
	Stream string
	Seq    uint64
}

type Publisher interface {
	Publish(ctx context.Context, subject string, data []byte, opts *PubOptions) (*PubAck, error)
}

type ContextRepo interface {
	GetMemoriesByUserID(
		ctx context.Context,
		userID uuid.UUID,
		numberOfMemories int32,
	) ([]Memory, error)
	GetSessionsByUserID(
		ctx context.Context,
		userID uuid.UUID,
		numberOfSessions int32,
	) ([]Session, error)
	GetMessagesBySessionIDOrdered(
		ctx context.Context,
		sessionID uuid.UUID,
	) ([]Message, error)
}
