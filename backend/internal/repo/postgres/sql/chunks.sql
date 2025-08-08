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
  FROM rag.chunks
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
      WHEN cardinality(sqlc.narg('content_types')::content_type[]) > 0 THEN
        ARRAY[
          paradedb.term_set(terms => (
            SELECT ARRAY_AGG(paradedb.term('content_type', ct))
            FROM UNNEST(sqlc.narg('content_types')::content_type[]) AS ct
          ))
        ]
      ELSE ARRAY[]::paradedb.searchqueryinput[]
    END
    ||
    CASE
      WHEN cardinality(sqlc.narg('sources')::source[]) > 0 THEN
        ARRAY[
          paradedb.term_set(terms => (
            SELECT ARRAY_AGG(paradedb.term('source', s))
            FROM UNNEST(sqlc.narg('sources')::source[]) AS s
          ))
        ]
      ELSE ARRAY[]::paradedb.searchqueryinput[]
    END
    ||
    CASE
      WHEN cardinality(sqlc.narg('surahs')::int[]) > 1 THEN
        ARRAY[
          paradedb.term_set(terms => (
            SELECT ARRAY_AGG(paradedb.term('surah', s))
            FROM UNNEST(sqlc.narg('surahs')::int[]) AS s
          ))
        ]

        WHEN cardinality(sqlc.narg('surahs')::int[]) = 1 AND
             cardinality(sqlc.narg('ayahs')::int[]) > 0 THEN
        ARRAY[
          paradedb.boolean(
            should => (
              SELECT ARRAY_AGG(
                paradedb.boolean(must => ARRAY[
                  paradedb.term('surah', s),
                  paradedb.term('ayah', a)
                ])
              )
              FROM UNNEST(sqlc.narg('ayahs')::int[]) AS a,
                   (SELECT UNNEST(sqlc.narg('surahs')::int[]) LIMIT 1) AS t(s)
            )
          )
        ]

      ELSE ARRAY[]::paradedb.searchqueryinput[]
    END
  )
),
deduped_chunks AS (
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
    (1 - (embedding <=> sqlc.arg('vector')::vector))::float8 AS score,
    embedded_chunk,
    raw_chunk,
    source,
    surah,
    ayah
  FROM rag.chunks
  WHERE
    (
      cardinality(sqlc.narg('content_type_labels')::smallint[]) = 0
      OR labels && sqlc.narg('content_type_labels')::smallint[]
    )
    AND
    (
      cardinality(sqlc.narg('source_labels')::smallint[]) = 0
      OR labels && sqlc.narg('source_labels')::smallint[]
    )
    AND (
      (
        cardinality(sqlc.narg('surah_labels')::smallint[]) = 0
        AND cardinality(sqlc.narg('ayah_labels')::smallint[]) = 0
      )
      OR (
        cardinality(sqlc.narg('surah_labels')::smallint[]) > 1
        AND labels && sqlc.narg('surah_labels')::smallint[]
      )
      OR (
        cardinality(sqlc.narg('surah_labels')::smallint[]) = 1
        AND cardinality(sqlc.narg('ayah_labels')::smallint[]) > 0
        AND labels && sqlc.narg('surah_labels')::smallint[]
        AND labels && sqlc.narg('ayah_labels')::smallint[]
      )
      OR (
        cardinality(sqlc.narg('surah_labels')::smallint[]) = 1
        AND cardinality(sqlc.narg('ayah_labels')::smallint[]) = 0
        AND labels && sqlc.narg('surah_labels')::smallint[]
      )
    )
),
deduped_chunks AS (
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
