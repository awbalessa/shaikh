-- name: GetAyatByKeys :many
SELECT * FROM rag.ayat
WHERE surah = sqlc.arg('surah')
    AND ayah = ANY(sqlc.arg('ayat')::int[]);
