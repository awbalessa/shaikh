package dom

import (
	"encoding/json"
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
	FnName            *string
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
