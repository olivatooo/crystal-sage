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
	if discord.URL == "" {
		fmt.Println("[Discord][Log]", "[", discord.Alias, "]", "[", level, "]", content)
		return
	}
	if discord.Debug {
		fmt.Println("[Discord][Log]", "[", discord.Alias, "]", "[", level, "]", content)
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
	if discord.Debug {
		fmt.Println("[Discord][Log]", "[", discord.Alias, "]", "[", level, "]", res.Status)
	}
	defer res.Body.Close()
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

func (discord *Discord) VariantLog(title string, description string, color int) {
	if discord.URL == "" {
		return
	}
	if discord.Variant == nil {
		if discord.Debug {
			fmt.Printf("[Discord][VariantLog] No variant config for shard: %s\n", discord.Alias)
		}
		return
	}
	if discord.Debug {
		fmt.Println("[Discord][VariantLog]", "[", discord.Alias, "]", title, description)
	}

	payloadBytes, err := BuildDiscordVariantMessage(discord.Variant, title, description, color)
	if err != nil {
		if discord.Debug {
			fmt.Printf("[Discord][VariantLog] Error building message: %v\n", err)
		}
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
	if discord.Debug {
		fmt.Println("[Discord][VariantLog]", "[", discord.Alias, "]", res.Status)
	}
	defer res.Body.Close()
}
