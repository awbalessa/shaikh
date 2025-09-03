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
SET updated_at = NOW(),
max_turn = sqlc.narg(max_turn),
max_turn_summarized = sqlc.narg(max_turn_summarized),
archived_at = sqlc.narg(archived_at),
summary = sqlc.narg(summary)
WHERE id = @id
RETURNING *;

-- name: GetMaxTurnByID :one
SELECT max_turn FROM sessions
WHERE id = @id;

-- name: ListWithSummaryBacklog :many
SELECT * FROM sessions
WHERE max_turn > max_turn_summarized;
