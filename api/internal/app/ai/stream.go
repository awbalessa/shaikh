package ai

import "io"

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
