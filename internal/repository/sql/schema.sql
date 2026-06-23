CREATE TABLE IF NOT EXISTS videos (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    published_at DATETIME NOT NULL,
    playlist_id TEXT NOT NULL
);
