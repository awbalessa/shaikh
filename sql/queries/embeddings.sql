-- name: CreateEmbedding :one
INSERT INTO
    embeddings (
        created_at,
        updated_at,
        granularity,
        content_type,
        lang,
        literature_source,
        embedding_title,
        embedding,
        metadata
    )
VALUES
    (NOW (), NOW (), $1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ResetEmbeddings :exec
DELETE FROM embeddings;
