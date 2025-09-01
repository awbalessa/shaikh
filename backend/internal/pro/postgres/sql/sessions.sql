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

-- name: UpdateSessionByID :one
UPDATE sessions
SET updated_at = @updated_at, ended_at = @ended_at, max_turn = @max_turn, summary = @summary
WHERE id = @id
RETURNING *;

-- name: GetMaxTurnByID :one
SELECT max_turn FROM sessions
WHERE id = @id;

-- name: ListWithBacklog :many
SELECT * FROM sessions
WHERE max_turn > max_turn_summarized;
