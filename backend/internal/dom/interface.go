package dom

import (
	"context"
	"time"
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
	SemanticSearch(
		ctx context.Context,
		vector VectorWithLabel,
		topk int,
	) ([]Chunk, error)
}

type LexicalSearcher interface {
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
		onPart func(LLMPart) bool,
	) error

	CountTokens(
		ctx context.Context,
		model string,
		window []*LLMContent,
		cfg *LLMCountConfig,
	) (int32, error)
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
