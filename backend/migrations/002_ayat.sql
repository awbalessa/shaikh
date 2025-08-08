-- +goose Up
-- +goose StatementBegin
CREATE TABLE rag.ayat (
    surah rag.surah NOT NULL,
    ayah rag.ayah NOT NULL,
    ar text NOT NULL,
    ar_uthmani text NOT NULL,
    en text NOT NULL
);

ALTER TABLE ONLY rag.ayat
    ADD CONSTRAINT ayat_pkey PRIMARY KEY (surah, ayah);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS rag.ayat;
-- +goose StatementEnd