package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/deliseev/video-notifier/internal/config"
	"github.com/deliseev/video-notifier/internal/infrastructure/memory"
	"github.com/deliseev/video-notifier/internal/infrastructure/notifier"
	"github.com/deliseev/video-notifier/internal/infrastructure/youtube"
	"github.com/deliseev/video-notifier/internal/usecase"
	"github.com/fsnotify/fsnotify"
)

const interval = 30 * time.Minute

var (
	// Используем sync для безопасного управления воркерами
	wg     sync.WaitGroup
	cancel context.CancelFunc
)

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// 1. Настраиваем канал для сигналов завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt) // Ловим Ctrl+C

	// Начальный запуск
	reloadWorkers()

	// Следим за файлом
	err = watcher.Add("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// В macOS событие Write может срабатывать дважды, это нормально
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("Config modified. Reloading...")
				reloadWorkers()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		case <-sigChan: // <--- ОП! Пришел Ctrl+C
			log.Println("Shutdown signal received...")
			if cancel != nil {
				cancel() // Останавливаем воркеры
			}
			wg.Wait() // Ждем их завершения
			log.Println("Graceful shutdown complete.")
			return // Выходим из программы
		}
	}
}

func reloadWorkers() {
	// Останавливаем старые воркеры (если были)
	if cancel != nil {
		cancel()
		wg.Wait() // Ждем, пока старые завершатся
	}

	// Создаем новый контекст для новых воркеров
	var ctx context.Context
	ctx, cancel = context.WithCancel(context.Background())

	// Загружаем конфиг
	f, err := os.Open("config.yaml")
	if err != nil {
		fmt.Println("config.yaml not found.")
	}
	cfg, err := config.LoadConfig(f)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	repo := memory.NewInMemoryVideoRepo()
	notifier := notifier.NewLogNotifier()

	for _, pl := range cfg.Playlists {
		fetcher, err := getFetcher(pl.Source)
		if err != nil {
			log.Printf("skipping playlist %s: %v", pl.ID, err)
			continue
		}

		checker := usecase.NewVideoChecker(fetcher, repo, notifier)

		wg.Add(1)
		go func(p config.PlaylistConfig) {
			defer wg.Done()
			runWorker(ctx, checker, p.ID, interval)
		}(pl)
	}
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
