-- name: CreateChunk :one
INSERT INTO
    chunks (
        granularity,
        content_type,
        raw_content,
        embedded_content,
        lang,
        literature_source,
        embedding_title,
        embedding,
        metadata
    )
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateEmbedding :exec
UPDATE chunks
SET
    embedding = $1
WHERE id = $2;

-- name: UpdateEmbeddingAndContent :exec
UPDATE chunks
SET
    embedded_content = $1,
    embedding = $2
WHERE id = $3;

-- name: CosineSimilarity :many
SELECT
    raw_content,
    metadata,
    literature_source,
    (1.0 - (embedding <=> $1))::double precision AS similarity
FROM chunks
ORDER BY similarity DESC
LIMIT $2;
