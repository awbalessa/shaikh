-- +goose Up

-- Documents table types
CREATE TYPE granularity AS enum ('phrase', 'ayah', 'surah', 'quran');

CREATE TYPE content_type AS enum ('tafsir');

CREATE TYPE source AS enum (
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

CREATE TABLE documents (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW (),
    granularity granularity NOT NULL,
    content_type content_type NOT NULL,
    source source NOT NULL,
    context_header TEXT NOT NULL,
    document TEXT NOT NULL,
    -- If document is Surah or Ayah specific
    surah INTEGER,
    ayah INTEGER,
    FOREIGN KEY (surah, ayah) REFERENCES ayat(surah, ayah)
);

-- Create B-Tree indices for frequent filters
CREATE INDEX idx_documents_surah_ayah ON documents (surah, ayah);
CREATE INDEX idx_documents_source ON documents (source);
CREATE INDEX idx_documents_content_type ON documents (content_type);

-- +goose Down

-- Drop indices
DROP INDEX idx_documents_surah_ayah;
DROP INDEX idx_documents_source;
DROP INDEX idx_documents_content_type;

-- Drop types
DROP TYPE source;

DROP TYPE content_type;

DROP TYPE granularity;
