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
        WHEN cardinality(sqlc.narg('content_types')::content_type[]) > 0
        THEN ARRAY[
           paradedb.term_set(terms => (
            SELECT ARRAY_AGG(paradedb.term('content_type', ct))
            FROM UNNEST(sqlc.narg('content_types')::content_type[]) AS ct
           ))
        ]
        ELSE ARRAY[]::paradedb.searchqueryinput[]
    END
    ||
    CASE
      WHEN cardinality(sqlc.narg('sources')::source[]) > 0
      THEN ARRAY[
        paradedb.term_set(terms => (
          SELECT ARRAY_AGG(paradedb.term('source', s))
          FROM UNNEST(sqlc.narg('sources')::source[]) AS s
        ))
      ]
      ELSE ARRAY[]::paradedb.searchqueryinput[]
    END
    ||
    CASE
        -- Case 1: surahs length > 1 → filter by surahs only
        WHEN cardinality(sqlc.narg('surahs')::int[]) > 1 THEN
            ARRAY[
                paradedb.term_set(terms => (
                    SELECT ARRAY_AGG(paradedb.term('surah', s))
                    FROM UNNEST(sqlc.narg('surahs')::int[]) AS s
                ))
            ]

        -- Case 2: surahs length = 1 → filter by that surah and optional ayahs
        WHEN cardinality(sqlc.narg('surahs')::int[]) = 1 THEN
            ARRAY[
                paradedb.term('surah', (SELECT s FROM UNNEST(sqlc.narg('surahs')::int[]) AS s))
            ] || (
                CASE
                    WHEN cardinality(sqlc.narg('ayahs')::int[]) > 0 THEN
                        (
                            SELECT ARRAY_AGG(paradedb.term('ayah', a))
                            FROM UNNEST(sqlc.narg('ayahs')::int[]) AS a
                        )
                    ELSE ARRAY[]::paradedb.searchqueryinput[]
                END
            )

        -- Case 3: no surahs → nothing
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
