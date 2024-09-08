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
		fmt.Errorf("[Slack][Log]", "[", slack.Alias, "]", "[", level, "]", content)
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
