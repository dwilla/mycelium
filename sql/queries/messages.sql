-- name: GetMessagesForChannel :many
SELECT messages.*, users.username 
FROM messages
INNER JOIN users ON messages.author = users.id
WHERE channel = $1
ORDER BY created_at ASC;

-- name: AddMessage :one
INSERT INTO messages (id, author, channel, body)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    $3
) RETURNING *;