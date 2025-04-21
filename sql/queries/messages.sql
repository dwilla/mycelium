-- name: GetMessagesForChannel :many
SELECT * FROM messages
WHERE channel = $1
ORDER BY created_at ASC;