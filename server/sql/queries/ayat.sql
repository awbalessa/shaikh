-- name: GetAyatByKeys :many
SELECT * FROM ayat
WHERE surah = $1 AND ayah = ANY($2::int[]);
