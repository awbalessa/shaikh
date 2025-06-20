-- +goose Up
CREATE EXTENSION vector;

CREATE EXTENSION vectorscale;

CREATE EXTENSION pg_search;

CREATE TABLE chunks (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW (),
    granularity granularity NOT NULL,
    content_type content_type NOT NULL,
    source source NOT NULL,
    raw_chunk TEXT NOT NULL,
    tokenized_chunk TEXT[] NOT NULL,
    -- Context header will be a piece of text prepended with every chunk that provides context around the chunk
    context_header TEXT NOT NULL,
    -- context_header + raw_chunk
    -- You will potentially want a chunk title to situate within parent documents.
    embedded_chunk TEXT NOT NULL,
    embedding VECTOR (1536) NOT NULL,
    -- Integer mappings of labels for pre-vector search filtering
    labels SMALLINT[] NOT NULL,
    has_parent BOOL NOT NULL,
    -- If chunk has a Parent document
    parent_id INTEGER REFERENCES documents(id),
    -- If chunk is Surah or Ayah specific
    surah INTEGER,
    ayah INTEGER,
    FOREIGN KEY (surah, ayah) REFERENCES ayat(surah, ayah)
    );

-- Add composite unique constraint
-- Create B-Tree indices for frequent filters
CREATE INDEX btree_chunks_surah_ayah ON chunks (surah, ayah);
CREATE INDEX btree_chunks_source ON chunks (source);
CREATE INDEX btree_chunks_content_type ON chunks (content_type);

-- Create StreamingDiskAnn index on embedding and labels
CREATE INDEX diskann_chunks_embedding_labels ON chunks
USING diskann (embedding vector_cosine_ops, labels);
-- +goose Down

-- Drop composite unique constraint
ALTER TABLE chunks

-- Drop table chunks
DROP TABLE chunks;

-- Drop indicies

DROP INDEX btree_chunks_surah_ayah;
DROP INDEX btree_chunks_source;
DROP INDEX btree_chunks_content_type;

-- Drop extensions
DROP EXTENSION pg_search;

DROP EXTENSION vectorscale;

DROP EXTENSION vector;
