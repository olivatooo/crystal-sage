package shards

import (
	"fmt"
	"net/http"
	"strings"
)

var ErrorLevel = map[string]int{
	"info":        3426654,
	"warn":        15844367,
	"err":         15548997,
	"success":     5763719,
	"aqua":        1752220,
	"blue":        3447003,
	"purple":      10181046,
	"yellow":      15844367,
	"orange":      15105570,
	"red":         15158332,
	"grey":        9807270,
	"dark_green":  3066993,
	"light_green": 3066993,
	"light_grey":  9807270,
	"navy":        3426654,
	"dark_blue":   2123412,
	"green":       3066993,
}

type Discord struct {
	*Shard
}

func (discord *Discord) Log(content string, level uint8) {
	fmt.Printf("[Discord][Log][Start] Shard: %s, Level: %d, Content length: %d\n", discord.Alias, level, len(content))
	if discord.URL == "" {
		fmt.Printf("[Discord][Log][Error] No URL configured for shard '%s', logging to console only\n", discord.Alias)
		fmt.Println("[Discord][Log]", "[", discord.Alias, "]", "[", level, "]", content)
		return
	}
	method := "POST"
	payload := strings.NewReader(fmt.Sprintf("{\"content\": \"%s\"}", content))
	fmt.Printf("[Discord][Log] Creating HTTP request to Discord webhook\n")
	client := &http.Client{}
	req, err := http.NewRequest(method, discord.URL, payload)
	if err != nil {
		fmt.Printf("[Discord][Log][Error] Failed to create request: %v\n", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	fmt.Printf("[Discord][Log] Sending HTTP request to Discord\n")
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("[Discord][Log][Error] Failed to send request: %v\n", err)
		return
	}
	fmt.Printf("[Discord][Log] Received response: Status=%s, StatusCode=%d\n", res.Status, res.StatusCode)
	defer res.Body.Close()
	fmt.Printf("[Discord][Log][Complete] Successfully sent message to Discord shard '%s'\n", discord.Alias)
}

func (discord *Discord) RawLog(content string) {
	if discord.URL == "" {
		return
	}
	method := "POST"
	payload := strings.NewReader(fmt.Sprintf("{\"content\": \"%s\"}", content))
	client := &http.Client{}
	req, err := http.NewRequest(method, discord.URL, payload)
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

func (discord *Discord) VariantLog(variant *VariantConfig, title string, description string, color int) {
	if discord.URL == "" {
		return
	}
	if variant == nil {
		fmt.Printf("[Discord][VariantLog][Error] No variant config provided for shard: %s\n", discord.Alias)
		return
	}
	fmt.Printf("[Discord][VariantLog][Start] Shard: %s, Title: %s, Description: %s, Color: %d\n", discord.Alias, title, description, color)

	payloadBytes, err := BuildDiscordVariantMessage(variant, title, description, color)
	if err != nil {
		fmt.Printf("[Discord][VariantLog][Error] Error building message: %v\n", err)
		return
	}

	method := "POST"
	client := &http.Client{}
	req, err := http.NewRequest(method, discord.URL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return
	}
	fmt.Printf("[Discord][VariantLog] Received response: Status=%s, StatusCode=%d\n", res.Status, res.StatusCode)
	defer res.Body.Close()
	fmt.Printf("[Discord][VariantLog][Complete] Successfully sent variant message to Discord shard '%s'\n", discord.Alias)
}
