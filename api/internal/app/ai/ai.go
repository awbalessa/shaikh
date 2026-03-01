package ai

import (
	"context"
	"encoding/json"
	"io"
)

type Model interface {
	ID() string
	Provider() string
	Generate(ctx context.Context, call CallOptions) (GenerateResult, error)
	Stream(ctx context.Context, call CallOptions) (StreamResult, error)
}

type CallOptions struct {
	Prompt Prompt
	Tools []*Tool
	ToolChoice *ToolChoice
	MaxOutputTokens *int32
	Temperature *float32
	PresencePenalty *float32
	FrequencyPenalty *float32
}


type GenerateResult struct {
	Contents []Content
	Usage    *Usage
}

type Prompt []Message


type Message struct {
	Role Role
	Content []Part
}

type Role string

const (
	RoleSystem Role = "system"
	RoleUser Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool Role = "tool"
)
type Part interface {
	Type() PartType
}

type PartType string

const (
	PartText PartType = "text"
	PartReasoning PartType = "reasoning"
	PartFile PartType = "file"
	PartToolCall PartType = "tool-call"
	PartToolResult PartType = "tool-result"
)

type TextPart struct {
	Text string
}

func (TextPart) Type() PartType { return PartText }

type ReasoningPart struct {
	Text string
}

func (ReasoningPart) Type() PartType { return PartReasoning }

type ToolCallPart struct {
	ToolCallID string
	ToolName string
	Input json.RawMessage
}

func (ToolCallPart) Type() PartType { return PartToolCall }

type ToolResultPart struct {
	ToolCallID string
	ToolName string
	Result json.RawMessage
	IsError bool
}

func (ToolResultPart) Type() PartType { return PartToolResult }

type FilePart struct {
	Filename string
	Data []byte
	MediaType string
}

func (FilePart) Type() PartType { return PartFile }

type Tool struct {
	Name string
	Description string
	InputSchema json.RawMessage
	InputExamples []json.RawMessage
}

type ToolChoiceType string

const (
	ToolChoiceAuto ToolChoiceType = "auto"
	ToolChoiceNone ToolChoiceType = "none"
	ToolChoiceRequired ToolChoiceType = "required"
	ToolChoiceTool ToolChoiceType = "tool"
)

type ToolChoice struct {
	Type ToolChoiceType
	ToolName string
}

type Content interface {
	Type() ContentType
}

type ContentType string

const (
	ContentText ContentType = "text"
	ContentReasoning ContentType = "reasoning"
	ContentFile ContentType = "file"
	ContentToolCall ContentType = "tool-call"
	ContentToolResult ContentType = "tool-result"
	ContentSource ContentType = "source"
)

type TextContent struct {
	Text string
}

func (TextContent) Type() ContentType { return ContentText }

type ReasoningContent struct {
	Text string
}

func (ReasoningContent) Type() ContentType { return ContentReasoning }

type FileContent struct {
	Data []byte
	MediaType string
}

func (FileContent) Type() ContentType { return ContentFile }

type ToolCallContent struct {
	ToolCallID string
	ToolName string
	Input json.RawMessage
}

func (ToolCallContent) Type() ContentType { return ContentToolCall }

type ToolResultContent struct {
	ToolCallID string
	ToolName string
	Result json.RawMessage
	IsError bool
}

func (ToolResultContent) Type() ContentType { return ContentToolResult }

type SourceType string

const (
	SourceDocument SourceType = "document"
)

type SourceContent struct {
	ID string
	SourceType SourceType
	Title string
	MediaType string
}

func (SourceContent) Type() ContentType { return ContentSource }


type StreamResult interface {
	Recv() (Event, error)
	Close() error
}

var _ = io.EOF

type EventType string

const (
	EventStreamStart EventType = "stream-start"
	EventTextDelta   EventType = "text-delta"
	EventTextStart   EventType = "text-start"
	EventTextEnd     EventType = "text-end"

	EventResponseMetadata EventType = "response-metadata"

	EventToolCall        EventType = "tool-call"
	EventToolCallDelta   EventType = "tool-call-delta"
	EventToolResult      EventType = "tool-result"

	EventFinish EventType = "finish"
	EventError  EventType = "error"
)

type FinishReason string

const (
	FinishReasonStop FinishReason = "stop"
	FinishReasonLength FinishReason = "length"
	FinishReasonToolCalls FinishReason = "tool-calls"
	FinishReasonContentFilter FinishReason = "content-filter"
	FinishReasonError FinishReason = "error"
	FinishReasonOther FinishReason = "other"
)

type Usage struct {
	InputTokens int
	OutputTokens int
	TotalTokens int
	ReasoningTokens int
}

type Event struct {
	Type EventType
	ID string
	Delta string
	Text string
	ToolName string
	ToolInput any
	ToolOutput any
	ToolError bool
	Reason FinishReason
	Usage *Usage
	Err error
}
