-- name: GetAyatByKeys :many
SELECT * FROM ayat
WHERE surah = sqlc.arg('surah')
    AND ayah = ANY(sqlc.arg('ayat')::int[]);
