package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type TelegramConfig struct {
	BotToken string `json:"bot_token"`
	ChatID   string `json:"chat_id"`
}

type FirebaseEnv struct {
	AppID          string `json:"app_id"`
	Groups         string `json:"groups"`
	ServiceAccountPath string `json:"service_account_path"`
}

type DiawiConfig struct {
	Token string `json:"token"`
}

type Config struct {
	Telegram TelegramConfig         `json:"telegram"`
	Firebase map[string]FirebaseEnv `json:"firebase"`
	Diawi    DiawiConfig            `json:"diawi"`
}

// Load reads the configuration from deploy-config.json.
func Load() (*Config, error) {
	data, err := os.ReadFile("deploy-config.json")
	if err != nil {
		return nil, fmt.Errorf("read deploy-config.json: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse deploy-config.json: %w", err)
	}
	return &cfg, nil
}
