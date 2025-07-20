-- name: LexicalSearch :many
SELECT
  id,
  paradedb.score(id)::float8 AS score,
  embedded_chunk,
  source,
  surah,
  ayah
FROM chunks
WHERE id @@@ paradedb.boolean(
  must => ARRAY[
    paradedb.boolean(
      should => ARRAY[
        paradedb.match('tokenized_chunk', sqlc.arg('query')::text),
        paradedb.match('tokenized_chunk_title', sqlc.arg('query')::text)
      ]
    )
  ]
  ||
  CASE
    WHEN sqlc.narg('content_type')::content_type IS NOT NULL
    THEN ARRAY[paradedb.term('content_type', sqlc.narg('content_type')::content_type)]
    ELSE ARRAY[]::paradedb.searchqueryinput[]
  END
  ||
  CASE
    WHEN sqlc.narg('source')::source IS NOT NULL
    THEN ARRAY[paradedb.term('source', sqlc.narg('source')::source)]
    ELSE ARRAY[]::paradedb.searchqueryinput[]
  END
  ||
  CASE
    WHEN sqlc.narg('surah')::int IS NOT NULL
    THEN ARRAY[paradedb.term('surah', sqlc.narg('surah')::int)]
    ELSE ARRAY[]::paradedb.searchqueryinput[]
  END
  ||
  CASE
    WHEN sqlc.narg('surah')::int IS NOT NULL
     AND sqlc.narg('ayah_start')::int IS NOT NULL
     AND sqlc.narg('ayah_end')::int IS NOT NULL
    THEN ARRAY[paradedb.range('ayah', int4range(sqlc.narg('ayah_start'), sqlc.narg('ayah_end'), '[]'))]
    ELSE ARRAY[]::paradedb.searchqueryinput[]
  END
)
ORDER BY score DESC
LIMIT sqlc.arg('number_of_chunks');


-- name: SemanticSearch :many
SELECT
    id,
    (1 - embedding <=> sqlc.arg('vector')::vector)::float8 as score,
    embedded_chunk,
    source,
    surah,
    ayah
FROM chunks
WHERE (
  cardinality(sqlc.arg('label_filters')::smallint[]) = 0
  OR labels && sqlc.arg('label_filters')::smallint[]
)
ORDER BY score DESC
LIMIT sqlc.arg('number_of_chunks');
