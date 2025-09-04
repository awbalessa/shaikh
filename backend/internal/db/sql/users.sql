-- name: CreateUser :one
INSERT INTO users (id, email, password_hash)
VALUES (@id, @email, @password_hash)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = @id;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = @email;

-- name: IncrementUserMessagesByID :one
UPDATE users
SET total_messages = total_messages + @delta_messages,
    total_messages_memorized = total_messages_memorized + @delta_messages_memorized,
    updated_at = NOW()
WHERE id = @id
RETURNING *;

-- name: ListUsersWithBacklog :many
SELECT * FROM users
WHERE total_messages > total_messages_memorized;

-- name: DeleteUserByID :exec
DELETE FROM users
WHERE id = @id;
