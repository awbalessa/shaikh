-- +goose Up
-- +goose StatementBegin
CREATE TABLE rag.documents (
    id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    granularity rag.granularity NOT NULL,
    content_type rag.content_type NOT NULL,
    source rag.source NOT NULL,
    context_header text NOT NULL,
    document text NOT NULL,
    surah rag.surah,
    ayah rag.ayah
);

ALTER TABLE ONLY rag.documents
    ADD CONSTRAINT documents_context_header_key UNIQUE (context_header);

ALTER TABLE ONLY rag.documents
    ADD CONSTRAINT documents_pkey PRIMARY KEY (id);

ALTER TABLE rag.documents ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME rag.documents_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);

ALTER TABLE ONLY rag.documents
    ADD CONSTRAINT documents_surah_ayah_fkey FOREIGN KEY (surah, ayah) REFERENCES rag.ayat(surah, ayah);

CREATE INDEX idx_documents_content_type ON rag.documents USING btree (content_type);
CREATE INDEX idx_documents_source ON rag.documents USING btree (source);
CREATE INDEX idx_documents_surah_ayah ON rag.documents USING btree (surah, ayah);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS rag.idx_documents_surah_ayah;
DROP INDEX IF EXISTS rag.idx_documents_source;
DROP INDEX IF EXISTS rag.idx_documents_content_type;

ALTER TABLE IF EXISTS rag.documents DROP CONSTRAINT IF EXISTS documents_surah_ayah_fkey;

ALTER TABLE IF EXISTS rag.documents ALTER COLUMN id DROP IDENTITY;

DROP TABLE IF EXISTS rag.documents;

DROP SEQUENCE IF EXISTS rag.documents_id_seq;
-- +goose StatementEnd
