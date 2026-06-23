package notifier

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/deliseev/video-notifier/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramNotifier struct {
	bot    *tgbotapi.BotAPI
	chatID int64
}

func NewTelegramNotifier(token string, chatID int64) (*TelegramNotifier, error) {
	// Создаем кастомный транспорт, который форсит IPv4
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Принудительно используем tcp4
			return dialer.DialContext(ctx, "tcp4", addr)
		},
	}

	bot, err := tgbotapi.NewBotAPIWithClient(token, tgbotapi.APIEndpoint, &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return &TelegramNotifier{bot: bot, chatID: chatID}, nil
}

func (n *TelegramNotifier) NotifyNewVideo(ctx context.Context, video domain.Video) error {
	msgText := fmt.Sprintf("🎥 *New Video!* \n%s\n%s", video.Title, video.URL)
	msg := tgbotapi.NewMessage(n.chatID, msgText)
	msg.ParseMode = "Markdown"

	_, err := n.bot.Send(msg)
	return err
}
