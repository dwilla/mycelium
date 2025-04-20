-- name: GetUsers :many
SELECT * FROM users;

-- name: GetUserFromId :one
SELECT * FROM users
WHERE id = $1;