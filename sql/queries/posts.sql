-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: GetPostsForUser :many
SELECT posts.* FROM posts
INNER JOIN feeds ON feeds.id=posts.feed_id
INNER JOIN feed_follows ON feed_follows.feed_id=feeds.id
INNER JOIN users ON feed_follows.user_id=users.id
WHERE users.name=$1
ORDER BY posts.updated_at DESC
limit $2;