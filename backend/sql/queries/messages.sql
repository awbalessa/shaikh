-- name: CreateMessage :one
INSERT INTO messages (session_id, user_id, role, content, model, token_count)
VALUES (@session_id, @user_id, @role, @content, @model, @token_count)
RETURNING *;

-- name: GetMessageByID :one
SELECT * FROM messages
WHERE id = @id;

-- name: GetMessagesBySessionID :many
SELECT * FROM messages
WHERE session_id = @session_id
ORDER BY created_at DESC;

-- name: GetUserMessagesByUserID :many
SELECT m.*
FROM messages m
JOIN sessions s ON m.session_id = s.id
WHERE m.user_id = @user_id
  AND m.role = 'user'::messages_role
ORDER BY s.updated_at DESC, m.created_at DESC
LIMIT @number_of_messages;
