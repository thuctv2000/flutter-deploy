package telegram

import (
	"bytes"
	"encoding/json"
	"flutter-deploy/internal/config"
	"fmt"
	"net/http"
)

// Notify sends a deployment notification via Telegram.
const apiBase = "https://api.telegram.org/bot"

func SendMessage(cfg *config.TelegramConfig, message string) error {
	return post(cfg.BotToken, "sendMessage", map[string]any{
		"chat_id": cfg.ChatID,
		"text":    message,
	})
}

func post(token, method string, payload any) error {
	b, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s%s/%s", apiBase, token, method)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
