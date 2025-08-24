-- +goose Up
-- +goose StatementBegin
CREATE TABLE rag.chunks (
    id bigint NOT NULL,
    sequence_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    granularity rag.granularity NOT NULL,
    content_type rag.content_type NOT NULL,
    source rag.source NOT NULL,
    raw_chunk text NOT NULL,
    tokenized_chunk text NOT NULL,
    chunk_title text NOT NULL,
    tokenized_chunk_title text NOT NULL,
    context_header text NOT NULL,
    embedded_chunk text NOT NULL,
    labels smallint[] NOT NULL,
    embedding public.vector(1024) NOT NULL,
    parent_id integer,
    surah rag.surah,
    ayah rag.ayah
);

ALTER TABLE ONLY rag.chunks
    ADD CONSTRAINT chunks_pkey PRIMARY KEY (id);

ALTER TABLE rag.chunks ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME rag.chunks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);

ALTER TABLE ONLY rag.chunks
    ADD CONSTRAINT unique_context_header_sequence_id_key UNIQUE (context_header, sequence_id);

ALTER TABLE ONLY rag.chunks
    ADD CONSTRAINT chunks_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES rag.documents(id);

ALTER TABLE ONLY rag.chunks
    ADD CONSTRAINT chunks_surah_ayah_fkey FOREIGN KEY (surah, ayah) REFERENCES rag.ayat(surah, ayah);

CREATE INDEX bm25_chunks_tokenized_chunk ON rag.chunks USING bm25 (id, tokenized_chunk, tokenized_chunk_title, content_type, source, surah, ayah) WITH (key_field=id, text_fields='{
        "tokenized_chunk": {
            "tokenizer": {"type": "whitespace", "stemmer": "Arabic"}
        },
        "tokenized_chunk_title": {
            "tokenizer": {"type": "whitespace", "stemmer": "Arabic"}
        }
    }', numeric_fields='{
        "surah": {"fast": true},
        "ayah": {"fast": true},
        "content_type": {"fast": true},
        "source": {"fast": true}
    }');

CREATE INDEX btree_chunks_content_type ON rag.chunks USING btree (content_type);
CREATE INDEX btree_chunks_source ON rag.chunks USING btree (source);
CREATE INDEX btree_chunks_surah_ayah ON rag.chunks USING btree (surah, ayah);
CREATE INDEX diskann_chunks_embedding_labels ON rag.chunks USING diskann (embedding, labels);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS rag.diskann_chunks_embedding_labels;
DROP INDEX IF EXISTS rag.btree_chunks_surah_ayah;
DROP INDEX IF EXISTS rag.btree_chunks_source;
DROP INDEX IF EXISTS rag.btree_chunks_content_type;
DROP INDEX IF EXISTS rag.bm25_chunks_tokenized_chunk;

ALTER TABLE IF EXISTS rag.chunks DROP CONSTRAINT IF EXISTS chunks_surah_ayah_fkey;
ALTER TABLE IF EXISTS rag.chunks DROP CONSTRAINT IF EXISTS chunks_parent_id_fkey;

ALTER TABLE IF EXISTS rag.chunks DROP CONSTRAINT IF EXISTS unique_context_header_sequence_id_key;

ALTER TABLE IF EXISTS rag.chunks ALTER COLUMN id DROP IDENTITY;

DROP TABLE IF EXISTS rag.chunks;

DROP SEQUENCE IF EXISTS rag.chunks_id_seq;
-- +goose StatementEnd
