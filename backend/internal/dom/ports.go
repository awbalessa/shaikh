package dom

import (
	"context"
	"fmt"
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
	Name string         `json:"name"`
	Args map[string]any `json:"args"`
}

type LLMFunctionResponse struct {
	Name     string         `json:"name"`
	Content  map[string]any `json:"content"`
	Metadata map[string]any `json:"-"`
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

const (
	SchemaString  string = "STRING"
	SchemaInteger string = "INTEGER"
	SchemaNumber  string = "NUMBER"
	SchemaBoolean string = "BOOLEAN"
	SchemaArray   string = "ARRAY"
	SchemaObject  string = "OBJECT"
)

type LLMSchema struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`

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
	ResponseMimeType   LLMResponseSchema
	ResponseSchema     *LLMSchema
}

type LLMCountConfig struct {
	System *LLMContent
	Tools  []*LLMFunctionDecl
}

type LLMResponseSchema string

const (
	ResponseJson LLMResponseSchema = "application/json"
	ResponseText LLMResponseSchema = "text/plain"
)

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

const (
	TokenLimit int32 = 200_000
)

var AgentToModel = map[AgentName]LargeLanguageModel{
	Caller:    GeminiV2p5Flash,
	Generator: GeminiV2p5FlashLite,
}

type TokenUsage struct {
	InputTokens  int32
	OutputTokens int32
}

const (
	FinishReasonUnspecified           string = "FINISH_REASON_UNSPECIFIED"
	FinishReasonStop                  string = "STOP"
	FinishReasonMaxTokens             string = "MAX_TOKENS"
	FinishReasonSafety                string = "SAFETY"
	FinishReasonRecitation            string = "RECITATION"
	FinishReasonLanguage              string = "LANGUAGE"
	FinishReasonOther                 string = "OTHER"
	FinishReasonBlocklist             string = "BLOCKLIST"
	FinishReasonProhibitedContent     string = "PROHIBITED_CONTENT"
	FinishReasonSPII                  string = "SPII"
	FinishReasonMalformedFunctionCall string = "MALFORMED_FUNCTION_CALL"
	FinishReasonImageSafety           string = "IMAGE_SAFETY"
	FinishReasonUnexpectedToolCall    string = "UNEXPECTED_TOOL_CALL"
)

type LLMInput struct {
	Text             string               `json:"text,omitempty"`
	FunctionResponse *LLMFunctionResponse `json:"function_response,omitempty"`
}

type LLMOutput struct {
	Text         string           `json:"text,omitempty"`
	FunctionCall *LLMFunctionCall `json:"function_call,omitempty"`
	Json         []byte           `json:"json,omitempty"`
}

type LLMOut interface {
	Text() string
	FunctionCall() *LLMFunctionCall
	MarshalJSON() ([]byte, error)
	TokenUsage() (int32, int32)
	Finish() (string, string)
}

type LLM interface {
	Generate(
		ctx context.Context,
		model string,
		window []*LLMContent,
		cfg *LLMGenConfig,
	) (LLMOut, error)
	Stream(
		ctx context.Context,
		model string,
		window []*LLMContent,
		cfg *LLMGenConfig,
		yield func(LLMOut, error) bool,
	) *Inference
	CountTokens(
		ctx context.Context,
		model string,
		window []*LLMContent,
		cfg *LLMCountConfig,
	) (int32, error)
}

type CallResult struct {
	Response map[string]any
	Metadata map[string]any
}

type AgentFn interface {
	Call(ctx context.Context, args map[string]any) (*CallResult, error)
}

type Agent interface {
	Generate(
		ctx context.Context,
		name AgentName,
		win []*LLMContent,
	) (LLMOut, error)
	Stream(
		ctx context.Context,
		name AgentName,
		win []*LLMContent,
		yield func(LLMOut, error) bool,
	) *Inference
	BuildContextWindow(
		ctx context.Context,
		name AgentName,
		cw *ContextWindow,
		now time.Time,
	) ([]*LLMContent, error)
}

type AgentName string

const (
	Caller     AgentName = "Caller"
	Generator  AgentName = "Generator"
	Summarizer AgentName = "Summarizer"
	Memorizer  AgentName = "Memorizer"
)

type AgentProfile struct {
	Model  string
	Config *LLMGenConfig
}

type AgentFnName string

const (
	FunctionSearch AgentFnName = "Search()"
)

type AgentFns map[AgentFnName]AgentFn

type AgentStruct struct {
	Agents map[AgentName]*AgentProfile
	LLM    LLM
}

func BuildAgent(llm LLM) *AgentStruct {
	agents := map[AgentName]*AgentProfile{
		Caller:     BuildCaller(),
		Generator:  BuildGenerator(),
		Summarizer: BuildSummarizer(),
		Memorizer:  BuildMemorizer(),
	}

	return &AgentStruct{
		Agents: agents,
		LLM:    llm,
	}
}

func (a *AgentStruct) Generate(
	ctx context.Context,
	name AgentName,
	win []*LLMContent,
) (LLMOut, error) {
	prof, ok := a.Agents[name]
	if !ok {
		return nil, fmt.Errorf("generate: %w", ErrInvalidInput)
	}

	resp, err := a.LLM.Generate(ctx, prof.Model, win, prof.Config)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (a *AgentStruct) Stream(
	ctx context.Context,
	name AgentName,
	win []*LLMContent,
	yield func(LLMOut, error) bool,
) *Inference {
	prof, ok := a.Agents[name]
	if !ok {
		yield(nil, fmt.Errorf("stream: %w", ErrInvalidInput))
		return nil
	}

	inf := a.LLM.Stream(ctx, prof.Model, win, prof.Config, yield)
	inf.Model = AgentToModel[name]
	return inf
}

type PubMsg struct {
	Subject string
	Reply   string
	Data    []byte
}

type DurablePubOptions struct {
	MsgID string
}

type DurablePubAck struct {
	Stream string
	Seq    uint64
}

type DurablePubMsgMetadata struct {
	Stream       string
	Consumer     string
	NumDelivered uint64
	Timestamp    time.Time
}

type DurablePubMsg interface {
	Data() []byte
	Subject() string
	Ack() error
	Nak() error
	Term() error
	InProgress() error
	Metadata() (DurablePubMsgMetadata, error)
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
	Publisher() Publisher
	Subscriber() Subscriber
	CreateStream(ctx context.Context, cfg PubSubStreamConfig) error
	CreateConsumer(ctx context.Context, stream string, cfg PubSubConsumerConfig) (PubSubConsumer, error)
}

type Subscriber interface {
	Subscribe(subject string, handler func(msg *PubMsg)) error
}

type Publisher interface {
	Publish(subject string, data []byte) error
	Request(ctx context.Context, subject string, data []byte) (*PubMsg, error)
	DurablePublish(ctx context.Context, subject string, data []byte, opts *DurablePubOptions) (*DurablePubAck, error)
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
	Fetch(batch int) ([]DurablePubMsg, error)
	Messages(ctx context.Context) (<-chan DurablePubMsg, <-chan error, error)
}

type MemoryRepo interface {
	CreateMemory(
		ctx context.Context,
		userID uuid.UUID,
		sourceMsg string,
		confidence float32,
		unique_key string,
		content string,
	) (*Memory, error)
	UpsertMemory(
		ctx context.Context,
		userID uuid.UUID,
		sourceMsg string,
		confidence float32,
		unique_key string,
		content string,
	) (*Memory, error)
	GetMemoriesByUserID(
		ctx context.Context,
		userID uuid.UUID,
		numberOfMemories int32,
	) ([]*Memory, error)
	DeleteMemoryByUserIDKey(
		ctx context.Context,
		userID uuid.UUID,
		key string,
	) error
}

type SessionRepo interface {
	CreateSession(
		ctx context.Context,
		id, userID uuid.UUID,
	) (*Session, error)
	GetSessionByID(
		ctx context.Context,
		id uuid.UUID,
	) (*Session, error)
	GetSessionsByUserID(
		ctx context.Context,
		userID uuid.UUID,
		numberOfSessions int32,
	) ([]*Session, error)
	UpdateSessionByID(
		ctx context.Context,
		id uuid.UUID,
		maxTurn *int32,
		maxTurnSummarized *int32,
		summary *string,
		archived_at *time.Time,
	) (*Session, error)
	GetMaxTurnByID(
		ctx context.Context,
		id uuid.UUID,
	) (int32, error)
	ListSessionsWithBacklog(
		ctx context.Context,
	) ([]*Session, error)
	DeleteSessionByID(
		ctx context.Context,
		id uuid.UUID,
	) error
}

type MessageRepo interface {
	CreateMessage(
		ctx context.Context,
		msg Message,
	) (Message, error)
	GetMessagesBySessionID(
		ctx context.Context,
		sessionID uuid.UUID,
	) ([]Message, error)
	GetUserMessagesByUserID(
		ctx context.Context,
		userID uuid.UUID,
		numberOfMessages int32,
	) ([]Message, error)
}

type UserRepo interface {
	CreateUser(
		ctx context.Context,
		id uuid.UUID,
		email, hash string,
	) (*User, error)
	GetUserByID(
		ctx context.Context,
		id uuid.UUID,
	) (*User, error)
	GetUserByEmail(
		ctx context.Context,
		email string,
	) (*User, error)
	IncrementUserMessagesByID(
		ctx context.Context,
		id uuid.UUID,
		delta int32,
		deltaMemorized int32,
	) (*User, error)
	ListUsersWithBacklog(
		ctx context.Context,
	) ([]*User, error)
	DeleteUserByID(
		ctx context.Context,
		id uuid.UUID,
	) error
}

type Tx interface {
	Get(repo any) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type UnitOfWork interface {
	Begin(ctx context.Context) (Tx, error)
}

type Probe interface {
	Name() string
	Ping(ctx context.Context) error
}

type RefreshTokenRepo interface {
	CreateRefreshToken(
		ctx context.Context,
		userID uuid.UUID,
		ttl time.Duration,
	) (string, error)
	ValidateAndRotate(
		ctx context.Context,
		rawToken string,
	) (uuid.UUID, error)
	Revoke(
		ctx context.Context,
		rawToken string,
	) error
	RevokeAll(
		ctx context.Context,
		userID uuid.UUID,
	) error
}
