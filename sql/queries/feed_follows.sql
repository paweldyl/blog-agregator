-- name: GetFeedFollowsForUser :many
SELECT feeds.name AS feed_name, users.name AS user_name FROM feed_follows 
INNER JOIN users ON feed_follows.user_id=users.id 
INNER JOIN feeds ON feed_follows.feed_id=feeds.id
WHERE users.name=$1;

-- name: DeleteFeedFollowsForUser :one
DELETE FROM feed_follows 
USING users, feeds
WHERE users.name = $1 
  AND feeds.url = $2
  AND feed_follows.user_id = users.id
  AND feed_follows.feed_id = feeds.id
RETURNING NULL;

-- name: CreateFeedFollow :one
WITH inserted AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id) VALUES (
        $1,
        $2,
        $3,
        $4,
        $5
    )
    RETURNING *
)
SELECT inserted.*, users.name AS user_name, feeds.name AS feed_name FROM inserted
INNER JOIN users ON users.id=inserted.user_id
INNER JOIN feeds ON feeds.id=inserted.feed_id;