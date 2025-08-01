-- name: CreateSession :one
INSERT INTO sessions (id, user_id)
VALUES (@id, @user_id)
RETURNING *;

-- name: GetSessionByID :one
SELECT * FROM sessions
WHERE id = @id;

-- name: GetSessionsByUserID :many
SELECT * FROM sessions
WHERE user_id = @user_id
ORDER BY updated_at DESC
LIMIT @number_of_sessions;
