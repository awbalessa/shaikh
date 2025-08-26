-- name: LexicalSearch :many
WITH ranked_chunks AS (
  SELECT
    id,
    paradedb.score(id)::float8 AS score,
    embedded_chunk,
    raw_chunk,
    source,
    surah,
    ayah,
    parent_id
  FROM rag.chunks
  WHERE id @@@ paradedb.boolean(
    must => ARRAY[
      paradedb.boolean(
        should => ARRAY[
          paradedb.match('tokenized_chunk',        sqlc.arg('query')::text),
          paradedb.match('tokenized_chunk_title',  sqlc.arg('query')::text)
        ]
      )
    ]
    -- OPTIONAL: content_types
    || CASE
         WHEN cardinality(COALESCE(sqlc.narg('content_types')::rag.content_type[], '{}')) >= 1 THEN
           ARRAY[
             paradedb.term_set(terms => (
               SELECT ARRAY_AGG(paradedb.term('content_type', ct))
               FROM UNNEST(COALESCE(sqlc.narg('content_types')::rag.content_type[], '{}')) AS ct
             ))
           ]
         ELSE ARRAY[]::paradedb.searchqueryinput[]
       END
    -- OPTIONAL: sources
    || CASE
         WHEN cardinality(COALESCE(sqlc.narg('sources')::rag.source[], '{}')) >= 1 THEN
           ARRAY[
             paradedb.term_set(terms => (
               SELECT ARRAY_AGG(paradedb.term('source', s))
               FROM UNNEST(COALESCE(sqlc.narg('sources')::rag.source[], '{}')) AS s
             ))
           ]
         ELSE ARRAY[]::paradedb.searchqueryinput[]
       END
    -- OPTIONAL: surahs / ayahs
    || CASE
         -- many surahs
         WHEN cardinality(COALESCE(sqlc.narg('surahs')::rag.surah[], '{}')) >= 2 THEN
           ARRAY[
             paradedb.term_set(terms => (
               SELECT ARRAY_AGG(paradedb.term('surah', s))
               FROM UNNEST(COALESCE(sqlc.narg('surahs')::rag.surah[], '{}')) AS s
             ))
           ]
         -- single surah + specific ayahs
         WHEN cardinality(COALESCE(sqlc.narg('surahs')::rag.surah[], '{}')) = 1
          AND cardinality(COALESCE(sqlc.narg('ayahs')::rag.ayah[],   '{}')) >= 1 THEN
           ARRAY[
             paradedb.boolean(
               should => (
                 SELECT ARRAY_AGG(
                   paradedb.boolean(must => ARRAY[
                     paradedb.term('surah', s),
                     paradedb.term('ayah',  a)
                   ])
                 )
                 FROM UNNEST(COALESCE(sqlc.narg('ayahs')::rag.ayah[], '{}')) AS a,
                      (SELECT UNNEST(sqlc.narg('surahs')::rag.surah[]) LIMIT 1) AS t(s)
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
    ayah,
    parent_id
  FROM ranked_chunks
  ORDER BY raw_chunk, score DESC, id  -- deterministic tiebreaker
)
SELECT id, score, embedded_chunk, source, surah, ayah, parent_id
FROM deduped_chunks
ORDER BY score DESC, id
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
    ayah,
    parent_id
  FROM rag.chunks
  WHERE
    (
      cardinality(COALESCE(sqlc.narg('content_type_labels')::smallint[], '{}')) = 0
      OR labels && COALESCE(sqlc.narg('content_type_labels')::smallint[], '{}')
    )
    AND
    (
      cardinality(COALESCE(sqlc.narg('source_labels')::smallint[], '{}')) = 0
      OR labels && COALESCE(sqlc.narg('source_labels')::smallint[], '{}')
    )
    AND (
      -- no surah nor ayah filter
      (
        cardinality(COALESCE(sqlc.narg('surah_labels')::smallint[], '{}')) = 0
        AND cardinality(COALESCE(sqlc.narg('ayah_labels')::smallint[],  '{}')) = 0
      )
      -- multiple surahs
      OR (
        cardinality(COALESCE(sqlc.narg('surah_labels')::smallint[], '{}')) >= 2
        AND labels && COALESCE(sqlc.narg('surah_labels')::smallint[], '{}')
      )
      -- single surah + specific ayahs
      OR (
        cardinality(COALESCE(sqlc.narg('surah_labels')::smallint[], '{}')) = 1
        AND cardinality(COALESCE(sqlc.narg('ayah_labels')::smallint[],  '{}')) >= 1
        AND labels && COALESCE(sqlc.narg('surah_labels')::smallint[], '{}')
        AND labels && COALESCE(sqlc.narg('ayah_labels')::smallint[],  '{}')
      )
      -- single surah only
      OR (
        cardinality(COALESCE(sqlc.narg('surah_labels')::smallint[], '{}')) = 1
        AND cardinality(COALESCE(sqlc.narg('ayah_labels')::smallint[],  '{}')) = 0
        AND labels && COALESCE(sqlc.narg('surah_labels')::smallint[], '{}')
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
    ayah,
    parent_id
  FROM ranked_chunks
  ORDER BY raw_chunk, score DESC, id  -- deterministic tiebreaker
)
SELECT id, score, embedded_chunk, source, surah, ayah, parent_id
FROM deduped_chunks
ORDER BY score DESC, id
LIMIT sqlc.arg('number_of_chunks');
