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
	fmt.Printf("[Telegram][Log][Start] Shard: %s, Level: %d, Content length: %d\n", telegram.Alias, level, len(content))
	if telegram.URL == "" {
		fmt.Printf("[Telegram][Log][Error] No URL configured for shard '%s', logging to console only\n", telegram.Alias)
		fmt.Printf("[Telegram][Log] [%s] [%d] %s\n", telegram.Alias, level, content)
		return
	}
	if telegram.ChatID == "" {
		fmt.Printf("[Telegram][Log][Error] chat_id is required for shard '%s'\n", telegram.Alias)
		return
	}

	// Telegram Bot API endpoint format: https://api.telegram.org/bot<token>/sendMessage
	// URL should contain the bot token, chat_id goes in the JSON payload
	apiURL := telegram.URL

	method := "POST"
	// Create JSON payload with proper escaping
	fmt.Printf("[Telegram][Log] Creating JSON payload\n")
	payloadData := map[string]string{
		"chat_id": telegram.ChatID,
		"text":    content,
	}
	jsonPayload, err := json.Marshal(payloadData)
	if err != nil {
		fmt.Printf("[Telegram][Log][Error] Failed to marshal JSON payload: %v\n", err)
		return
	}
	fmt.Printf("[Telegram][Log] JSON payload created: %d bytes\n", len(jsonPayload))

	client := &http.Client{}
	req, err := http.NewRequest(method, apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Printf("[Telegram][Log][Error] Failed to create request: %v\n", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	fmt.Printf("[Telegram][Log] Sending HTTP request to Telegram API\n")
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("[Telegram][Log][Error] Failed to send request: %v\n", err)
		return
	}
	fmt.Printf("[Telegram][Log] Received response: Status=%s, StatusCode=%d\n", res.Status, res.StatusCode)
	if telegram.Debug {
		fmt.Println("[Telegram][Log]", "[", telegram.Alias, "]", "[", level, "]", res.Status)
	}
	defer res.Body.Close()
	fmt.Printf("[Telegram][Log][Complete] Successfully sent message to Telegram shard '%s'\n", telegram.Alias)
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

func (telegram *Telegram) VariantLog(title string, description string, color int) {
	if telegram.URL == "" {
		return
	}
	if telegram.ChatID == "" {
		return
	}
	if telegram.Variant == nil {
		fmt.Printf("[Telegram][VariantLog][Error] No variant config for shard: %s\n", telegram.Alias)
		return
	}
	fmt.Printf("[Telegram][VariantLog][Start] Shard: %s, Title: %s, Description: %s, Color: %d\n", telegram.Alias, title, description, color)

	payloadBytes, err := BuildTelegramVariantMessage(telegram.Variant, title, description)
	if err != nil {
		fmt.Printf("[Telegram][VariantLog][Error] Error building message: %v\n", err)
		return
	}

	// Parse the JSON to add chat_id
	var payloadData map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payloadData); err != nil {
		return
	}
	payloadData["chat_id"] = telegram.ChatID

	finalPayload, err := json.Marshal(payloadData)
	if err != nil {
		return
	}

	method := "POST"
	client := &http.Client{}
	req, err := http.NewRequest(method, telegram.URL, bytes.NewBuffer(finalPayload))
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")
	fmt.Printf("[Telegram][VariantLog] Sending HTTP request to Telegram API\n")
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("[Telegram][VariantLog][Error] Failed to send request: %v\n", err)
		return
	}
	fmt.Printf("[Telegram][VariantLog] Received response: Status=%s, StatusCode=%d\n", res.Status, res.StatusCode)
	defer res.Body.Close()
	fmt.Printf("[Telegram][VariantLog][Complete] Successfully sent variant message to Telegram shard '%s'\n", telegram.Alias)
}
