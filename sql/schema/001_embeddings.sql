-- +goose Up
CREATE EXTENSION IF NOT EXISTS vector;

-- Embeddings table types
CREATE TYPE granularity AS enum ('word', 'ayah', 'surah', 'quran');
CREATE TYPE content_type AS enum ('quran', 'tafseer');
CREATE TYPE lang AS enum ('ar', 'en');
CREATE TYPE literature_source AS enum (
    'Quran',
    'Ibn Kathir',
    'Al Tabari',
    'Al Qurtubi',
    'Al Baghawi',
    'Al Saadi',
    'Al Muyassar',
    'Al Wasit',
    'Al Jalalayn'
);

CREATE TABLE embeddings (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    granularity granularity NOT NULL,
    content_type content_type NOT NULL,
    raw_content TEXT NOT NULL,
    embedded_content TEXT NOT NULL,
    lang lang NOT NULL DEFAULT 'ar',
    literature_source literature_source NOT NULL,
    embedding_title TEXT NOT NULL,
    embedding VECTOR (768) NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb
);

-- Add composite unique constraint
ALTER TABLE embeddings ADD CONSTRAINT unique_content_metadata UNIQUE (content, metadata);

-- Add a GIN index on metadata
CREATE INDEX idx_embeddings_metadata ON embeddings USING GIN (metadata);

-- +goose Down
DROP TABLE embeddings;

-- Drop the GIN index
DROP INDEX IF EXISTS idx_embeddings_metadata;

-- Drop composite unique constraint
ALTER TABLE embeddings
DROP CONSTRAINT IF EXISTS unique_content_metadata;

DROP TYPE literature_source;
DROP TYPE lang;
DROP TYPE content_type;
DROP TYPE granularity;

DROP EXTENSION IF EXISTS vector;
