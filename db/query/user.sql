-- name: CreateUser :one
INSERT INTO users (
  username,
  password
) VALUES (
  $1, $2
) RETURNING *;

-- name: FindUserByUsername :one
SELECT * FROM users
WHERE username = $1;

-- name: FindUserByID :one
SELECT * FROM users
WHERE id = $1;