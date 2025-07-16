-- +goose Up
CREATE EXTENSION IF NOT EXISTS vector;

CREATE EXTENSION IF NOT EXISTS vectorscale;

CREATE EXTENSION IF NOT EXISTS pg_search;

CREATE TABLE IF NOT EXISTS chunks (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    sequence_id INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW (),
    granularity granularity NOT NULL,
    content_type content_type NOT NULL,
    source source NOT NULL,

    raw_chunk TEXT NOT NULL,
    tokenized_chunk TEXT NOT NULL,
    chunk_title TEXT NOT NULL,
    tokenized_chunk_title TEXT NOT NULL,
    context_header TEXT NOT NULL UNIQUE,
    embedded_chunk TEXT NOT NULL,
    embedding VECTOR (1024) NOT NULL,

    labels SMALLINT[] NOT NULL,
    has_parent BOOL NOT NULL,

    parent_id INTEGER REFERENCES documents(id),

    surah INTEGER,
    ayah INTEGER,
    FOREIGN KEY (surah, ayah) REFERENCES ayat(surah, ayah)
    );

-- Create B-Tree indices for frequent filters
CREATE INDEX IF NOT EXISTS btree_chunks_surah_ayah ON chunks (surah, ayah);
CREATE INDEX IF NOT EXISTS btree_chunks_source ON chunks (source);
CREATE INDEX IF NOT EXISTS btree_chunks_content_type ON chunks (content_type);

-- Create StreamingDiskAnn index on embedding and labels
CREATE INDEX IF NOT EXISTS diskann_chunks_embedding_labels ON chunks
USING diskann (embedding vector_cosine_ops, labels);

-- Create BM25 index for tokenized chunks
CREATE INDEX IF NOT EXISTS bm25_chunks_tokenized_chunk ON chunks
USING bm25 (id, tokenized_chunk, tokenized_chunk_title, content_type, source, surah, ayah)
WITH (
    key_field = 'id',
    text_fields = '{
        "tokenized_chunk": {
            "tokenizer": {"type": "whitespace"}
        },
        "tokenized_chunk_title": {
            "tokenizer": {"type": "whitespace"}
        }
    }',
    numeric_fields = '{
        "surah": {"fast": true},
        "ayah": {"fast": true},
        "content_type": {"fast": true},
        "source": {"fast": true}
    }'
);

-- +goose Down

-- Drop indices
DROP INDEX IF EXISTS btree_chunks_surah_ayah;
DROP INDEX IF EXISTS btree_chunks_source;
DROP INDEX IF EXISTS btree_chunks_content_type;
DROP INDEX IF EXISTS diskann_chunks_embedding_labels;
DROP INDEX IF EXISTS bm25_chunks_tokenized_chunk;

-- Drop table chunks
DROP TABLE IF EXISTS chunks;

-- Drop extensions
DROP EXTENSION IF EXISTS pg_search;

DROP EXTENSION IF EXISTS vectorscale;

DROP EXTENSION IF EXISTS vector;
