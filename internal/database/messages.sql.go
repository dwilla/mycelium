// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: messages.sql

package database

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

const addMessage = `-- name: AddMessage :one
INSERT INTO messages (id, author, channel, body)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    $3
) RETURNING id, author, channel, body, created_at, updated_at
`

type AddMessageParams struct {
	Author  uuid.UUID
	Channel uuid.UUID
	Body    string
}

func (q *Queries) AddMessage(ctx context.Context, arg AddMessageParams) (Message, error) {
	row := q.db.QueryRowContext(ctx, addMessage, arg.Author, arg.Channel, arg.Body)
	var i Message
	err := row.Scan(
		&i.ID,
		&i.Author,
		&i.Channel,
		&i.Body,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getMessagesForChannel = `-- name: GetMessagesForChannel :many
SELECT messages.id, messages.author, messages.channel, messages.body, messages.created_at, messages.updated_at, users.username 
FROM messages
INNER JOIN users ON messages.author = users.id
WHERE channel = $1
ORDER BY created_at ASC
`

type GetMessagesForChannelRow struct {
	ID        uuid.UUID
	Author    uuid.UUID
	Channel   uuid.UUID
	Body      string
	CreatedAt sql.NullTime
	UpdatedAt sql.NullTime
	Username  string
}

func (q *Queries) GetMessagesForChannel(ctx context.Context, channel uuid.UUID) ([]GetMessagesForChannelRow, error) {
	rows, err := q.db.QueryContext(ctx, getMessagesForChannel, channel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetMessagesForChannelRow
	for rows.Next() {
		var i GetMessagesForChannelRow
		if err := rows.Scan(
			&i.ID,
			&i.Author,
			&i.Channel,
			&i.Body,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Username,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
