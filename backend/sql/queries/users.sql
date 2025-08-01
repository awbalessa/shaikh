-- name: CreateUser :one
INSERT INTO users (id, email)
VALUES (@id, @email)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = @id;
