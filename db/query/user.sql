-- name: CreateUser :one
INSERT INTO users (
  username
) VALUES (
  $1
) RETURNING *;

-- name: FindUser :one
SELECT * FROM users
WHERE username = $1;