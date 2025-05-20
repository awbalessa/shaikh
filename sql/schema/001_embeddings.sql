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
    created_at TIMESTAMP NOT NULL,
    granularity granularity NOT NULL,
    content_type content_type NOT NULL,
    content TEXT NOT NULL UNIQUE,
    lang lang NOT NULL DEFAULT 'ar',
    literature_source literature_source NOT NULL,
    embedding_title TEXT NOT NULL,
    embedding VECTOR (768) NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb
);

-- +goose Down
DROP TABLE embeddings;

DROP TYPE literature_source;
DROP TYPE lang;
DROP TYPE content_type;
DROP TYPE granularity;

DROP EXTENSION IF EXISTS vector;
