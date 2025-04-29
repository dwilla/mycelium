-- +goose Up
CREATE TABLE subs (
    user_id UUID NOT NULL REFERENCES users,
    channel_id UUID NOT NULL REFERENCES channels ON DELETE CASCADE,
    notifications BOOLEAN DEFAULT false,
    UNIQUE (user_id, channel_id)
);

-- +goose Down
DROP TABLE subs;