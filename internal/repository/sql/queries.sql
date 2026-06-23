-- name: GetKnownVideoIDs :many
SELECT id FROM videos WHERE playlist_id = ?;

-- name: SaveVideo :exec
INSERT INTO videos (id, title, url, published_at, playlist_id) VALUES (?, ?, ?, ?, ?);
