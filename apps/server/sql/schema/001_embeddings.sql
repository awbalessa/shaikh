-- +goose Up
CREATE EXTENSION IF NOT EXISTS vector;

-- Embeddings table types
CREATE TYPE granularity AS enum ('word', 'ayah', 'surah', 'quran');

CREATE TYPE content_type AS enum ('quran', 'tafseer');

CREATE TYPE literature_source AS enum (
    'Al Quran',
    'Ibn Kathir',
    'Al Tabari',
    'Al Qurtubi',
    'Al Baghawi',
    'Al Saadi',
    'Al Muyassar',
    'Al Wasit',
    'Al Jalalayn'
    -- Add more
);

-- You'll probably have to rethink your schema. Some documents belong in a pure relational table.
-- Ayat, Tafsir, Hadith, Asbab Nozool... etc. Look into strcture of all the literature.
CREATE TABLE embeddings ( -- Figure out a better name
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW (),
    granularity granularity NOT NULL,
    content_type content_type NOT NULL,
    literature_source literature_source NOT NULL,
    raw_content TEXT NOT NULL,
    context_header TEXT NOT NULL,
    -- context_header + raw_content
    embedded_content TEXT NOT NULL,
    embedding VECTOR (1536) NOT NULL,
    is_chunk BOOL NOT NULL,
    -- NULLable columns that depend on is_chunk
    parent_document TEXT,
    -- NULLable columns that depend on granularity
    surah_number INT,
    surah TEXT,
    ayah_number INT,
    ayah TEXT,
    phrase TEXT,
);

-- Add composite unique constraint
--Figure this out ALTER TABLE embeddings ADD CONSTRAINT unique_content_metadata UNIQUE (raw_content, metadata);
-- Add a GIN index on document & chunk metadata
CREATE INDEX idx_embeddings_document_metadata ON embeddings USING GIN (document_metadata);

-- +goose Down
-- Drop the GIN index
DROP INDEX IF EXISTS idx_embeddings_metadata;

-- Drop composite unique constraint
ALTER TABLE documents
DROP CONSTRAINT IF EXISTS unique_content_metadata;

-- Drop table embeddings
DROP TABLE embeddings;

-- Drop types
DROP TYPE literature_source;

DROP TYPE content_type;

DROP TYPE granularity;

-- Drop pgvector
DROP EXTENSION IF EXISTS vector;
