-- name: GetChannels :many
SELECT * FROM channels;

-- name: GetChannelByName :one
SELECT * FROM channels
WHERE name SIMILAR TO $1;

-- name: GetChannelByID :one
SELECT * FROM channels 
WHERE id = $1;
