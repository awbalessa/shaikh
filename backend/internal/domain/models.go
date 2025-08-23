package domain

import (
	"github.com/jackc/pgx/v5/pgtype"
	pgvector_go "github.com/pgvector/pgvector-go"
)

type Memory struct {
	ID        int32
	UserID    pgtype.UUID
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	Memory    string
}

type Message struct {
	ID           int32
	SessionID    pgtype.UUID
	UserID       pgtype.UUID
	CreatedAt    pgtype.Timestamptz
	Role         MessagesRole
	Content      string
	Turn         int32
	FunctionName pgtype.Text
}

type RagAyah struct {
	Surah     RagSurahNumber
	Ayah      RagAyahNumber
	Ar        string
	ArUthmani string
	En        string
}

type RagChunk struct {
	ID                  int64
	SequenceID          int32
	CreatedAt           pgtype.Timestamp
	UpdatedAt           pgtype.Timestamp
	Granularity         RagGranularity
	ContentType         RagContentType
	Source              RagSource
	RawChunk            string
	TokenizedChunk      string
	ChunkTitle          string
	TokenizedChunkTitle string
	ContextHeader       string
	EmbeddedChunk       string
	Labels              []int16
	Embedding           pgvector_go.Vector
	ParentID            pgtype.Int4
	Surah               NullRagSurah
	Ayah                NullRagAyah
}

type RagDocument struct {
	ID            int32
	CreatedAt     pgtype.Timestamp
	UpdatedAt     pgtype.Timestamp
	Granularity   RagGranularity
	ContentType   RagContentType
	Source        RagSource
	ContextHeader string
	Document      string
	Surah         NullRagSurah
	Ayah          NullRagAyah
}

type Session struct {
	ID        pgtype.UUID
	UserID    pgtype.UUID
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	EndedAt   pgtype.Timestamptz
	Summary   pgtype.Text
}

type User struct {
	ID        pgtype.UUID
	Email     string
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}
