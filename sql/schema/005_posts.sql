-- +goose Up
CREATE TABLE posts (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    title TEXT,
    url TEXT NOT NULL UNIQUE,
    description TEXT,
    published_at TEXT NOT NULL,
    feed_id UUID NOT NULL REFERENCES feeds ON DELETE CASCADE
);
-- +goose Down
DROP TABLE posts;