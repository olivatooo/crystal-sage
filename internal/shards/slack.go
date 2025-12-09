package shards

import (
	"fmt"
	"net/http"
	"strings"
)

type Slack struct {
	*Shard
}

func (slack *Slack) Log(content string, level uint8) {
	fmt.Printf("[Slack][Log][Start] Shard: %s, Level: %d, Content length: %d\n", slack.Alias, level, len(content))
	if slack.URL == "" {
		fmt.Printf("[Slack][Log][Error] No URL configured for shard '%s', logging to console only\n", slack.Alias)
		fmt.Printf("[Slack][Log] [%s] [%d] %s\n", slack.Alias, level, content)
		return
	}
	method := "POST"
	payload := strings.NewReader(fmt.Sprintf("{\"text\": \"%s\"}", content))
	fmt.Printf("[Slack][Log] Creating HTTP request to Slack webhook\n")
	client := &http.Client{}
	req, err := http.NewRequest(method, slack.URL, payload)
	if err != nil {
		fmt.Printf("[Slack][Log][Error] Failed to create request: %v\n", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	fmt.Printf("[Slack][Log] Sending HTTP request to Slack\n")
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("[Slack][Log][Error] Failed to send request: %v\n", err)
		return
	}
	fmt.Printf("[Slack][Log] Received response: Status=%s, StatusCode=%d\n", res.Status, res.StatusCode)
	defer res.Body.Close()
	fmt.Printf("[Slack][Log][Complete] Successfully sent message to Slack shard '%s'\n", slack.Alias)
}

func (slack *Slack) RawLog(content string) {
	if slack.URL == "" {
		return
	}
	method := "POST"
	payload := strings.NewReader(fmt.Sprintf("{\"text\": \"%s\"}", content))
	client := &http.Client{}
	req, err := http.NewRequest(method, slack.URL, payload)
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

func (slack *Slack) VariantLog(title string, description string, color int) {
	if slack.URL == "" {
		return
	}
	if slack.Variant == nil {
		fmt.Printf("[Slack][VariantLog][Error] No variant config for shard: %s\n", slack.Alias)
		return
	}
	fmt.Printf("[Slack][VariantLog][Start] Shard: %s, Title: %s, Description: %s, Color: %d\n", slack.Alias, title, description, color)

	// Convert RGB to hex color for Slack
	colorHex := fmt.Sprintf("#%06x", color)
	if slack.Variant.Color > 0 {
		colorHex = fmt.Sprintf("#%06x", slack.Variant.Color)
	}

	payloadBytes, err := BuildSlackVariantMessage(slack.Variant, title, description, colorHex)
	if err != nil {
		fmt.Printf("[Slack][VariantLog][Error] Error building message: %v\n", err)
		return
	}

	method := "POST"
	client := &http.Client{}
	req, err := http.NewRequest(method, slack.URL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")
	fmt.Printf("[Slack][VariantLog] Sending HTTP request to Slack\n")
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("[Slack][VariantLog][Error] Failed to send request: %v\n", err)
		return
	}
	fmt.Printf("[Slack][VariantLog] Received response: Status=%s, StatusCode=%d\n", res.Status, res.StatusCode)
	defer res.Body.Close()
	fmt.Printf("[Slack][VariantLog][Complete] Successfully sent variant message to Slack shard '%s'\n", slack.Alias)
}
