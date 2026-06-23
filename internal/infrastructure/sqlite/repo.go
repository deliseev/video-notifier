package sqlite

import (
	"context"
	"database/sql"

	"github.com/deliseev/video-notifier/internal/domain"
	"github.com/deliseev/video-notifier/internal/repository"
	_ "modernc.org/sqlite"
)

type SqliteRepo struct {
	db      *sql.DB
	queries *repository.Queries
}

func NewSqliteRepo(path string) (*SqliteRepo, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	// Инициализируем схему (в MVP можно делать так,
	// в продакшене лучше использовать инструмент миграций)
	if err := createSchema(db); err != nil {
		return nil, err
	}

	return &SqliteRepo{
		db:      db,
		queries: repository.New(db),
	}, nil
}

func (r *SqliteRepo) GetKnownVideoIDs(ctx context.Context, playlistID string) ([]string, error) {
	return r.queries.GetKnownVideoIDs(ctx, playlistID)
}

func (r *SqliteRepo) SaveVideos(ctx context.Context, videos []domain.Video) error {
	// sqlite транзакция для атомарности
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)
	for _, v := range videos {
		err := qtx.SaveVideo(ctx, repository.SaveVideoParams{
			ID:          v.ID,
			Title:       v.Title,
			Url:         v.URL,
			PublishedAt: v.Published,
			PlaylistID:  v.PlaylistID,
		})
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func createSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS videos (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		url TEXT NOT NULL,
		published_at DATETIME NOT NULL,
		playlist_id TEXT NOT NULL
	);`
	_, err := db.Exec(schema)
	return err
}
