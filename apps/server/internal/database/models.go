// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package database

import (
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	pgvector_go "github.com/pgvector/pgvector-go"
)

type ContentType string

const (
	ContentTypeTafsir ContentType = "tafsir"
)

func (e *ContentType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ContentType(s)
	case string:
		*e = ContentType(s)
	default:
		return fmt.Errorf("unsupported scan type for ContentType: %T", src)
	}
	return nil
}

type NullContentType struct {
	ContentType ContentType
	Valid       bool // Valid is true if ContentType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullContentType) Scan(value interface{}) error {
	if value == nil {
		ns.ContentType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.ContentType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullContentType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.ContentType), nil
}

type Granularity string

const (
	GranularityPhrase Granularity = "phrase"
	GranularityAyah   Granularity = "ayah"
	GranularitySurah  Granularity = "surah"
	GranularityQuran  Granularity = "quran"
)

func (e *Granularity) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = Granularity(s)
	case string:
		*e = Granularity(s)
	default:
		return fmt.Errorf("unsupported scan type for Granularity: %T", src)
	}
	return nil
}

type NullGranularity struct {
	Granularity Granularity
	Valid       bool // Valid is true if Granularity is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullGranularity) Scan(value interface{}) error {
	if value == nil {
		ns.Granularity, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.Granularity.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullGranularity) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.Granularity), nil
}

type Source string

const (
	SourceTafsirIbnKathir      Source = "Tafsir Ibn Kathir"
	SourceTafsirAlTabari       Source = "Tafsir Al Tabari"
	SourceTafsirAlQurtubi      Source = "Tafsir Al Qurtubi"
	SourceTafsirAlBaghawi      Source = "Tafsir Al Baghawi"
	SourceTafsirAlSaadi        Source = "Tafsir Al Saadi"
	SourceTafsirAlMuyassar     Source = "Tafsir Al Muyassar"
	SourceTafsirAlWasit        Source = "Tafsir Al Wasit"
	SourceTafsirAlJalalayn     Source = "Tafsir Al Jalalayn"
	SourceTafsirTanwirAlMiqbas Source = "Tafsir Tanwir Al Miqbas"
)

func (e *Source) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = Source(s)
	case string:
		*e = Source(s)
	default:
		return fmt.Errorf("unsupported scan type for Source: %T", src)
	}
	return nil
}

type NullSource struct {
	Source Source
	Valid  bool // Valid is true if Source is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullSource) Scan(value interface{}) error {
	if value == nil {
		ns.Source, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.Source.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullSource) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.Source), nil
}

type Ayat struct {
	Surah     int32
	Ayah      int32
	Ar        string
	ArUthmani string
	En        string
}

type Chunk struct {
	ID                  int64
	SequenceID          int32
	CreatedAt           pgtype.Timestamp
	UpdatedAt           pgtype.Timestamp
	Granularity         Granularity
	ContentType         ContentType
	Source              Source
	RawChunk            string
	TokenizedChunk      string
	ChunkTitle          string
	TokenizedChunkTitle string
	ContextHeader       string
	EmbeddedChunk       string
	Labels              []int16
	Embedding           pgvector_go.Vector
	HasParent           bool
	ParentID            pgtype.Int4
	Surah               pgtype.Int4
	Ayah                pgtype.Int4
}

type Document struct {
	ID            int32
	CreatedAt     pgtype.Timestamp
	UpdatedAt     pgtype.Timestamp
	Granularity   Granularity
	ContentType   ContentType
	Source        Source
	ContextHeader string
	Document      string
	Surah         pgtype.Int4
	Ayah          pgtype.Int4
}
