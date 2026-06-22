package notifier

import (
	"context"
	"log"

	"github.com/deliseev/video-notifier/internal/domain"
)

type LogNotifier struct{}

func NewLogNotifier() *LogNotifier {
	return &LogNotifier{}
}

func (n *LogNotifier) NotifyNewVideo(ctx context.Context, video domain.Video) error {
	// Используем ANSI-цвета для красоты в консоли (это работает на macOS)
	const colorGreen = "\033[32m"
	const colorReset = "\033[0m"

	log.Printf("%s[NOTIFICATION]%s New video found: %s (%s)\n",
		colorGreen, colorReset, video.Title, video.URL)

	return nil
}
