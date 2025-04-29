-- name: GetChannels :many
SELECT * FROM channels;

-- name: GetChannelByName :one
SELECT * FROM channels
WHERE name SIMILAR TO $1;

-- name: GetChannelByID :one
SELECT * FROM channels 
WHERE id = $1;

-- name: CreateChannel :one
INSERT INTO channels (id, name, creator)
VALUES (
    gen_random_uuid(),
    $1,
    $2
) RETURNING *;
