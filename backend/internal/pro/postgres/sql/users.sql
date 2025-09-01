-- name: CreateUser :one
INSERT INTO users (id, email)
VALUES (@id, @email)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = @id;

-- name: IncrementUserMessagesByID :one
UPDATE users
SET total_messages = total_messages + @total_messages,
    total_messages_memorized = total_messages_memorized + @total_messages_memorized,
    updated_at = @updated_at
WHERE id = @id
RETURNING *;

-- name: ListWithBacklog :many
SELECT * FROM users
WHERE total_messages > total_messages_memorized;
