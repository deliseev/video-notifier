package usecase

import (
	"context"

	"github.com/deliseev/video-notifier/internal/domain"
)

// Определяем интерфейсы прямо здесь, в слое использования (Use Case).
// Это и есть инверсия зависимостей в Go: потребитель задает правила.

type VideoFetcher interface {
	FetchLatestVideos(ctx context.Context, playlistID string) ([]domain.Video, error)
}

type VideoRepository interface {
	GetKnownVideoIDs(ctx context.Context, playlistID string) ([]string, error)
	SaveVideos(ctx context.Context, videos []domain.Video) error
}

type Notifier interface {
	NotifyNewVideo(ctx context.Context, video domain.Video) error
}

// Checker — это наш сервис (Use Case)
type Checker struct {
	fetcher  VideoFetcher
	repo     VideoRepository
	notifier Notifier
}

func NewVideoChecker(f VideoFetcher, r VideoRepository, n Notifier) *Checker {
	return &Checker{fetcher: f, repo: r, notifier: n}
}

// Check — основной метод бизнес-логики
func (c *Checker) Check(ctx context.Context, playlistID string) error {
	// 1. Получаем видео из внешнего мира
	fetched, err := c.fetcher.FetchLatestVideos(ctx, playlistID)
	if err != nil {
		return err
	}

	// 2. Получаем ID тех, что мы уже знаем
	knownIDs, err := c.repo.GetKnownVideoIDs(ctx, playlistID)
	if err != nil {
		return err
	}

	// Делаем map для быстрого поиска (O(1))
	knownMap := make(map[string]bool)
	for _, id := range knownIDs {
		knownMap[id] = true
	}

	// 3. Сравниваем и находим новые
	var newVideos []domain.Video
	for _, v := range fetched {
		if !knownMap[v.ID] {
			newVideos = append(newVideos, v)
		}
	}

	// 4. Уведомляем, а потом сохраняем
	// - Если Notify падает — мы ничего не сохранили. При следующем запуске Check
	// сработает снова и попробует уведомить еще раз. (Это называется идемпотентность).
	// - Если Save падает — мы уведомили пользователя, но не сохранили ID.
	// При следующем запуске мы отправим дубль уведомления.
	if len(newVideos) > 0 {
		for _, v := range newVideos {
			if err := c.notifier.NotifyNewVideo(ctx, v); err != nil {
				return err
			}
		}
		if err := c.repo.SaveVideos(ctx, newVideos); err != nil {
			return err
		}
	}

	return nil
}
