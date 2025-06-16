-- +goose Up
CREATE TABLE IF NOT EXISTS ayat (
    id SERIAL PRIMARY KEY,
    surah INT NOT NULL,
    ayah INT NOT NULL,
    key VARCHAR(7) GENERATED ALWAYS AS (surah || ':' || ayah) STORED,
    ar TEXT NOT NULL,
    ar_uthmani TEXT NOT NULL,
    UNIQUE (surah, ayah)
);

CREATE UNIQUE INDEX idx_ayat_key ON ayat (key);

-- +goose Down
DROP INDEX IF EXISTS idx_ayat_key;

DROP TABLE ayat;
