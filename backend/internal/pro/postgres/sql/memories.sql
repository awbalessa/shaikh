-- name: CreateMemory :one
INSERT INTO memories (user_id, source_message, confidence, unique_key, memory)
VALUES (@user_id, @source_message, @confidence, @unique_key, @memory)
RETURNING *;

-- name: GetMemoryByID :one
SELECT * FROM memories
WHERE id = @id;

-- name: GetMemoriesByUserID :many
SELECT * FROM memories
WHERE user_id = @user_id
ORDER BY created_at DESC
LIMIT @number_of_memories;

-- name: UpdateMemoryByID :one
UPDATE memories
SET memory = @memory,
    updated_at = NOW()
WHERE id = @id
RETURNING *;

-- name: DeleteMemoryByUserIDKey :exec
DELETE FROM memories
WHERE user_id = @user_id AND unique_key = @key;

-- name: UpsertMemory :one
INSERT INTO memories (user_id, source_message, confidence, unique_key, memory)
VALUES (@user_id, @source_message, @confidence, @unique_key, @memory)
ON CONFLICT (user_id, unique_key) DO UPDATE
SET source_message = EXCLUDED.source_message,
    confidence = EXCLUDED.confidence,
    memory = EXCLUDED.memory,
    updated_at = NOW()
RETURNING *;
