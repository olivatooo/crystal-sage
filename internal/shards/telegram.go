package shards

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Telegram struct {
	*Shard
	ChatID string
}

func (telegram *Telegram) Log(content string, level uint8) {
	if telegram.URL == "" {
		fmt.Printf("[Telegram][Log] [%s] [%d] %s\n", telegram.Alias, level, content)
		return
	}
	if telegram.ChatID == "" {
		fmt.Printf("[Telegram][Log] [%s] chat_id is required\n", telegram.Alias)
		return
	}
	if telegram.Debug {
		fmt.Println("[Telegram][Log]", "[", telegram.Alias, "]", "[", level, "]", content)
	}

	// Telegram Bot API endpoint format: https://api.telegram.org/bot<token>/sendMessage
	// URL should contain the bot token, chat_id goes in the JSON payload
	apiURL := telegram.URL

	method := "POST"
	// Create JSON payload with proper escaping
	payloadData := map[string]string{
		"chat_id": telegram.ChatID,
		"text":    content,
	}
	jsonPayload, err := json.Marshal(payloadData)
	if err != nil {
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return
	}
	if telegram.Debug {
		fmt.Println("[Telegram][Log]", "[", telegram.Alias, "]", "[", level, "]", res.Status)
	}
	defer res.Body.Close()
}

func (telegram *Telegram) RawLog(content string) {
	if telegram.URL == "" {
		return
	}
	if telegram.ChatID == "" {
		return
	}

	apiURL := telegram.URL

	method := "POST"
	// Create JSON payload with proper escaping
	payloadData := map[string]string{
		"chat_id": telegram.ChatID,
		"text":    content,
	}
	jsonPayload, err := json.Marshal(payloadData)
	if err != nil {
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
}
