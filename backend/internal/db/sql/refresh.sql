-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
VALUES (@id, @user_id, @token_hash, @expires_at)
RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT * FROM refresh_tokens
WHERE token_hash = @token_hash;

-- name: RevokeRefreshTokenByID :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE id = @id;

-- name: RevokeRefreshTokenByHash :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE token_hash = @token_hash;

-- name: RevokeAllUserTokens :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE user_id = @user_id
  AND revoked_at IS NULL;

-- name: DeleteExpiredTokens :exec
DELETE FROM refresh_tokens
WHERE expires_at < now() OR revoked_at IS NOT NULL;

-- name: ListActiveTokensByUser :many
SELECT * FROM refresh_tokens
WHERE user_id = @user_id
  AND revoked_at IS NULL
  AND expires_at > now()
ORDER BY created_at DESC;
