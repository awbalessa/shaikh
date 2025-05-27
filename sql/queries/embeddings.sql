-- name: CreateEmbedding :one
INSERT INTO
    embeddings (
        created_at,
        granularity,
        content_type,
        content,
        lang,
        literature_source,
        embedding_title,
        embedding,
        metadata
    )
VALUES
    (NOW (), $1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateEmbedding :exec
UPDATE embeddings
SET
    embedding = $1
WHERE id = $2;

-- name: UpdateEmbeddingAndContent :exec
UPDATE embeddings
SET
    content = $1,
    embedding = $2
WHERE id = $3;

-- name: CosineSimilarity :many
SELECT
    content,
    metadata,
    literature_source,
    (1.0 - (embedding <=> $1))::double precision AS similarity
FROM embeddings
ORDER BY similarity DESC
LIMIT $2;

-- name: ResetEmbeddings :exec
DELETE FROM embeddings;
