-- name: CreateUser :one
INSERT INTO users (id, username, email, password_hash)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    $3
) RETURNING *;

-- name: GetUsers :many
SELECT * FROM users;

-- name: GetUserFromId :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id FROM users
WHERE email = $1;

-- name: GetUserByUsername :one
SELECT id FROM users
WHERE username = $1;