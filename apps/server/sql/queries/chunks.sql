-- name: LexicalSearch :many
WITH ranked_chunks AS (
    SELECT
    id,
    paradedb.score(id)::float8 AS score,
    embedded_chunk,
    raw_chunk,
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
), deduped_chunks AS (
  SELECT DISTINCT ON (raw_chunk)
    id,
    score,
    embedded_chunk,
    source,
    surah,
    ayah
  FROM ranked_chunks
  ORDER BY raw_chunk, score DESC
)
SELECT * FROM deduped_chunks
ORDER BY score DESC
LIMIT sqlc.arg('number_of_chunks');


-- name: SemanticSearch :many
WITH ranked_chunks AS (
    SELECT
        id,
        (1 - (embedding <=> sqlc.arg('vector')::vector))::float8 as score,
        embedded_chunk,
        raw_chunk,
        source,
        surah,
        ayah
    FROM chunks
    WHERE (
    cardinality(sqlc.arg('label_filters')::smallint[]) = 0
    OR labels && sqlc.arg('label_filters')::smallint[]
    )
), deduped_chunks AS (
    SELECT DISTINCT ON (raw_chunk)
        id,
        score,
        embedded_chunk,
        source,
        surah,
        ayah
    FROM ranked_chunks
    ORDER BY raw_chunk, score DESC
    )
SELECT * FROM deduped_chunks
ORDER BY score DESC
LIMIT sqlc.arg('number_of_chunks');
