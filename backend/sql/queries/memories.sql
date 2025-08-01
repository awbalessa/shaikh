-- name: CreateMemory :one
INSERT INTO memories (user_id, memory)
VALUES (@user_id, @memory)
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
    updated_at = now()
WHERE id = @id
RETURNING *;
