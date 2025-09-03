package dom

import (
	"context"
	"errors"
	"iter"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNoResults = errors.New("no results found")
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

type LLMContentResult struct {
	Text  *string
	Bytes []byte
}

type LLM interface {
	Generate(
		ctx context.Context,
		model string,
		window []*LLMContent,
		cfg *LLMGenConfig,
		format LLMResponseSchema,
	) (*LLMContentResult, error)
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

type AgentFn interface {
	Call(ctx context.Context, args map[string]any) (map[string]any, error)
}

type Agent interface {
	Generate(
		ctx context.Context,
		name AgentName,
		win []*LLMContent,
	) (*LLMContentResult, error)
	Stream(
		ctx context.Context,
		name AgentName,
		win []*LLMContent,
	) iter.Seq2[*LLMPart, error]
	StreamWithYield(
		ctx context.Context,
		name AgentName,
		win []*LLMContent,
		yield func(*LLMPart, error) bool,
	) *LLMGenResult
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
) (*LLMContentResult, error) {
	prof, ok := a.Agents[name]
	if !ok {
		return nil, ErrAgentDoesNotExist
	}

	resp, err := a.LLM.Generate(ctx, prof.Model, win, prof.Config, prof.Config.ResponseMimeType)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (a *AgentStruct) Stream(
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

func (a *AgentStruct) StreamWithYield(
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
	Messages(ctx context.Context) (<-chan DurablePubMsg, error)
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
	BelongsToUser(
		ctx context.Context,
		id, userID uuid.UUID,
	) (bool, error)
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
	GetMessagesBySessionIDOrdered(
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
}

type Tx interface {
	Get(repo any) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type UnitOfWork interface {
	Begin(ctx context.Context) (Tx, error)
}

var (
	ErrNotPingable = errors.New("provider not pingable")
)

type Probe interface {
	Name() string
	Ping(ctx context.Context) error
}
