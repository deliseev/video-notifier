package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/deliseev/video-notifier/internal/config"
	"github.com/deliseev/video-notifier/internal/infrastructure/memory"
	"github.com/deliseev/video-notifier/internal/infrastructure/notifier"
	"github.com/deliseev/video-notifier/internal/infrastructure/youtube"
	"github.com/deliseev/video-notifier/internal/usecase"
)

func main() {
	// Создаем контекст, который отменяется при сигнале ОС
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop() // Важно: очищаем ресурсы сигнала

	// Загружаем конфиг
	f, err := os.Open("config.yaml")
	if err != nil {
		fmt.Println("config.yaml not found.")
	}
	cfg, err := config.LoadConfig(f)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Инициализируем общие зависимости
	// repo := sqlite.NewRepository(cfg.DatabasePath)
	// notifier := telegram.NewNotifier(cfg.TelegramToken)

	// Временно используем заглушку, пока нет БД и Telegram
	repo := memory.NewInMemoryVideoRepo()
	notifier := notifier.NewLogNotifier()

	// Запускаем воркеры для каждого плейлиста
	for _, pl := range cfg.Playlists {
		fetcher, err := getFetcher(pl.Source)
		if err != nil {
			log.Printf("skipping playlist %s: %v", pl.ID, err)
			continue
		}

		checker := usecase.NewVideoChecker(fetcher, repo, notifier)

		// Запускаем каждый воркер в своей горутине
		go runWorker(ctx, checker, pl.ID, 30*time.Minute)
	}

	// Блокируемся, пока контекст не будет отменен
	<-ctx.Done()
	log.Println("Shutting down gracefully...")

	// Даем время на завершение (опционально)
	time.Sleep(1 * time.Second)
	log.Println("Goodbye!")
}

func runWorker(ctx context.Context, checker *usecase.Checker, playlistID string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Первый запуск сразу, не дожидаясь тика
	if err := checker.Check(ctx, playlistID); err != nil {
		log.Printf("error checking playlist %s: %v", playlistID, err)
	}

	for {
		select {
		case <-ticker.C:
			if err := checker.Check(ctx, playlistID); err != nil {
				log.Printf("error checking playlist %s: %v", playlistID, err)
			}
		case <-ctx.Done():
			log.Printf("Worker for %s stopping...", playlistID)
			return // Аккуратно выходим
		}
	}
}

func getFetcher(source string) (usecase.VideoFetcher, error) {
	switch source {
	case "youtube":
		return youtube.NewYoutubeFetcher(), nil
	case "rutube":
		// return rutube.NewRutubeFetcher(), nil
		return nil, fmt.Errorf("rutube not implemented yet")
	default:
		return nil, fmt.Errorf("unknown source: %s", source)
	}
}
