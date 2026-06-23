package config_test

import (
	"strings"
	"testing"

	"github.com/deliseev/video-notifier/internal/config"
)

func TestLoadConfig(t *testing.T) {
	yamlData := `
database_path: test.db
telegram_token: xxx
chat_id: 123
playlists:
- id: test-id
  source: youtube
`
	reader := strings.NewReader(yamlData)

	// Act
	cfg, err := config.LoadConfig(reader)

	// Assert
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if cfg.DatabasePath != "test.db" {
		t.Errorf("expected test.db, got %s", cfg.DatabasePath)
	}
	if len(cfg.Playlists) != 1 || cfg.Playlists[0].ID != "test-id" {
		t.Error("config playlists not parsed correctly")
	}
	if cfg.ChatID != 123 {
		t.Errorf("got %v want 123", cfg.ChatID)
	}
}
