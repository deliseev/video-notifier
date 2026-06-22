package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/deliseev/video-notifier/internal/domain"
	"github.com/deliseev/video-notifier/internal/usecase"
)

// 1. Создаем ручные моки (заглушки) для наших интерфейсов.
// В Go это просто структуры, которые мы настроим так, как нам нужно для теста.

type mockVideoFetcher struct {
	videosToReturn []domain.Video
}

func (m *mockVideoFetcher) FetchLatestVideos(ctx context.Context, playlistID string) ([]domain.Video, error) {
	return m.videosToReturn, nil
}

type mockVideoRepo struct {
	knownIDs    []string
	savedVideos []domain.Video
	// Поле для имитации ошибки
	shouldFail bool
}

func (m *mockVideoRepo) GetKnownVideoIDs(ctx context.Context, playlistID string) ([]string, error) {
	return m.knownIDs, nil
}
func (m *mockVideoRepo) SaveVideos(ctx context.Context, videos []domain.Video) error {
	if m.shouldFail {
		return errors.New("database connection error")
	}
	m.savedVideos = append(m.savedVideos, videos...) // Сохраняем в память мока, чтобы потом проверить
	return nil
}

type mockNotifier struct {
	notifiedVideos []domain.Video
}

func (m *mockNotifier) NotifyNewVideo(ctx context.Context, video domain.Video) error {
	m.notifiedVideos = append(m.notifiedVideos, video)
	return nil
}

// 2. Пишем сам тест с использованием Table-Driven подхода
func TestCheckNewVideos(t *testing.T) {
	// Подготовка данных для теста
	v1 := domain.Video{ID: "vid1", Title: "Old Video"}
	v2 := domain.Video{ID: "vid2", Title: "Brand New Video!"}

	// Таблица тест-кейсов
	tests := []struct {
		name           string
		fetchedVideos  []domain.Video
		knownIDs       []string
		failSave       bool // Флаг: должен ли репозиторий вернуть ошибку?
		expectErr      bool // Ожидаем ли мы ошибку от самого Checker.Check?
		expectSaved    int  // сколько видео должно сохраниться в базу
		expectNotified int  // сколько уведомлений должно уйти
	}{
		{
			name:           "1 new video found",
			fetchedVideos:  []domain.Video{v1, v2}, // Из API пришло два видео
			knownIDs:       []string{"vid1"},       // Но в базе мы уже знаем про первое
			expectSaved:    1,                      // Значит сохранить должны только 1
			expectNotified: 1,                      // И уведомить только об 1
		},
		{
			name:           "No new videos",
			fetchedVideos:  []domain.Video{v1},
			knownIDs:       []string{"vid1"},
			expectSaved:    0,
			expectNotified: 0,
		},
		{
			name:           "Save fails after notification",
			fetchedVideos:  []domain.Video{v2},
			knownIDs:       []string{},
			failSave:       true, // <-- имитируем сбой
			expectErr:      true, // Ожидаем, что сервис вернет ошибку
			expectSaved:    0,
			expectNotified: 1, // А вот тут ты увидишь проблему: уведомление ушло!
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: настраиваем моки для конкретного тест-кейса
			fetcher := &mockVideoFetcher{videosToReturn: tt.fetchedVideos}
			repo := &mockVideoRepo{knownIDs: tt.knownIDs, shouldFail: tt.failSave}
			notifier := &mockNotifier{}

			// Собираем наш сервис
			checker := usecase.NewVideoChecker(fetcher, repo, notifier)

			// Act: вызываем бизнес-логику
			err := checker.Check(context.Background(), "playlist_123")
			// Проверка на ошибку
			if (err != nil) != tt.expectErr {
				t.Fatalf("unexpected error state: %v", err)
			}

			// Assert: проверяем результаты
			if len(repo.savedVideos) != tt.expectSaved {
				t.Errorf("expected %d saved videos, got %d", tt.expectSaved, len(repo.savedVideos))
			}
			if len(notifier.notifiedVideos) != tt.expectNotified {
				t.Errorf("expected %d notified videos, got %d", tt.expectNotified, len(notifier.notifiedVideos))
			}
		})
	}
}
