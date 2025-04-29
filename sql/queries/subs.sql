-- name: GetChannelsForUser :many
SELECT channels.id, channels.name FROM subs
INNER JOIN channels ON subs.channel_id = channels.id
WHERE subs.user_id = $1;

-- name: GetUsersForChannel :many 
SELECT users.id, users.username FROM subs
INNER JOIN users ON subs.user_id = users.id
WHERE subs.channel_id = $1;

-- name: CreateSub :many
INSERT INTO subs (user_id, channel_id)
VALUES ($1, $2) RETURNING *;
