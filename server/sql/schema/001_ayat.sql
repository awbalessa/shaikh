-- +goose Up
CREATE TABLE IF NOT EXISTS ayat (
    surah INTEGER NOT NULL,
    ayah INTEGER NOT NULL,
    ar TEXT NOT NULL,
    ar_uthmani TEXT NOT NULL,
    en TEXT NOT NULL,
    PRIMARY KEY (surah, ayah)
);


-- +goose Down
DROP TABLE IF EXISTS ayat;
