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
	if slack.URL == "" {
		fmt.Printf("[Slack][Log] [%s] [%d] %s\n", slack.Alias, level, content)
		return
	}
	if slack.Debug {
		fmt.Println("[Slack][Log]", "[", slack.Alias, "]", "[", level, "]", content)
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
		if slack.Debug {
			fmt.Printf("[Slack][VariantLog] No variant config for shard: %s\n", slack.Alias)
		}
		return
	}
	if slack.Debug {
		fmt.Println("[Slack][VariantLog]", "[", slack.Alias, "]", title, description)
	}

	// Convert RGB to hex color for Slack
	colorHex := fmt.Sprintf("#%06x", color)
	if slack.Variant.Color > 0 {
		colorHex = fmt.Sprintf("#%06x", slack.Variant.Color)
	}

	payloadBytes, err := BuildSlackVariantMessage(slack.Variant, title, description, colorHex)
	if err != nil {
		if slack.Debug {
			fmt.Printf("[Slack][VariantLog] Error building message: %v\n", err)
		}
		return
	}

	method := "POST"
	client := &http.Client{}
	req, err := http.NewRequest(method, slack.URL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return
	}
	if slack.Debug {
		fmt.Println("[Slack][VariantLog]", "[", slack.Alias, "]", res.Status)
	}
	defer res.Body.Close()
}
