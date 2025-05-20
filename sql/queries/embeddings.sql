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

-- name: ResetEmbeddings :exec
DELETE FROM embeddings;
