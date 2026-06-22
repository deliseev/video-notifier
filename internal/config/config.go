package config

import (
	"io"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DatabasePath  string           `yaml:"database_path"`
	TelegramToken string           `yaml:"telegram_token"`
	Playlists     []PlaylistConfig `yaml:"playlists"`
}

type PlaylistConfig struct {
	ID     string `yaml:"id"`
	Source string `yaml:"source"`
}

func LoadConfig(r io.Reader) (*Config, error) {
	var cfg Config
	if err := yaml.NewDecoder(r).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
