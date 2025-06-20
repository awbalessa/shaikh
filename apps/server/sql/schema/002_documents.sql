-- +goose Up

-- Documents table types
CREATE TYPE IF NOT EXISTS granularity AS enum ('phrase', 'ayah', 'surah', 'quran');

CREATE TYPE IF NOT EXISTS content_type AS enum ('tafsir');

CREATE TYPE IF NOT EXISTS source AS enum (
    'Tafsir Ibn Kathir',
    'Tafsir Al Tabari',
    'Tafsir Al Qurtubi',
    'Tafsir Al Baghawi',
    'Tafsir Al Saadi',
    'Tafsir Al Muyassar',
    'Tafsir Al Wasit',
    'Tafsir Al Jalalayn',
    'Tafsir Tanwir Al Miqbas'
);

CREATE TABLE IF NOT EXISTS documents (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW (),
    granularity granularity NOT NULL,
    content_type content_type NOT NULL,
    source source NOT NULL,
    context_header TEXT NOT NULL UNIQUE,
    document TEXT NOT NULL,
    -- If document is Surah or Ayah specific
    surah INTEGER,
    ayah INTEGER,
    FOREIGN KEY (surah, ayah) REFERENCES ayat(surah, ayah)
    );

-- Create B-Tree indices for frequent filters
CREATE INDEX IF NOT EXISTS idx_documents_surah_ayah ON documents (surah, ayah);
CREATE INDEX IF NOT EXISTS idx_documents_source ON documents (source);
CREATE INDEX IF NOT EXISTS idx_documents_content_type ON documents (content_type);

-- +goose Down

-- Drop indices
DROP INDEX IF EXISTS idx_documents_surah_ayah;
DROP INDEX IF EXISTS idx_documents_source;
DROP INDEX IF EXISTS idx_documents_content_type;

-- Drop table
DROP TABLE IF EXISTS documents;

-- Drop types
DROP TYPE IF EXISTS source;

DROP TYPE IF EXISTS content_type;

DROP TYPE IF EXISTS granularity;
