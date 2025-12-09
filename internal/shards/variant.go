package shards

import (
	"encoding/json"
	"fmt"
)

// VariantConfig holds the configuration for variant-based messages
type VariantConfig struct {
	Title       string                   `yaml:"title"`
	Description string                   `yaml:"description"`
	Color       int                      `yaml:"color"`
	Author      string                   `yaml:"author"`
	Username    string                   `yaml:"username"`
	Fields      []map[string]interface{} `yaml:"fields"`
	Content     string                   `yaml:"content"`
}

// DiscordEmbed represents a Discord embed structure
type DiscordEmbed struct {
	Content  string        `json:"content,omitempty"`
	Embeds   []EmbedObject `json:"embeds,omitempty"`
	Username string        `json:"username,omitempty"`
}

// EmbedObject represents a Discord embed object
type EmbedObject struct {
	Title       string                   `json:"title,omitempty"`
	Description string                   `json:"description,omitempty"`
	Color       int                      `json:"color,omitempty"`
	Author      EmbedAuthor              `json:"author,omitempty"`
	Fields      []map[string]interface{} `json:"fields,omitempty"`
}

// EmbedAuthor represents the author field in Discord embeds
type EmbedAuthor struct {
	Name string `json:"name,omitempty"`
}

// BuildDiscordVariantMessage builds a Discord embed message from variant config and dynamic values
func BuildDiscordVariantMessage(variant *VariantConfig, title string, description string, color int) ([]byte, error) {
	embed := DiscordEmbed{}

	// Use variant username if provided
	if variant != nil && variant.Username != "" {
		embed.Username = variant.Username
	}

	// Use variant content if provided
	if variant != nil && variant.Content != "" {
		embed.Content = variant.Content
	}

	embedObj := EmbedObject{
		Title:       title,
		Description: description,
		Color:       color,
	}

	// Use variant values as defaults, but allow override with passed parameters
	if variant != nil {
		if variant.Author != "" {
			embedObj.Author = EmbedAuthor{
				Name: variant.Author,
			}
		}
		if variant.Fields != nil {
			embedObj.Fields = variant.Fields
		}
		// Use variant color if provided and no color passed, otherwise use passed color
		if variant.Color > 0 && color == 0 {
			embedObj.Color = variant.Color
		} else if color > 0 {
			embedObj.Color = color
		}
		// Use variant title as default if no title passed
		if variant.Title != "" && title == "" {
			embedObj.Title = variant.Title
		}
		// Use variant description as default if no description passed
		if variant.Description != "" && description == "" {
			embedObj.Description = variant.Description
		}
	}

	embed.Embeds = []EmbedObject{embedObj}

	return json.Marshal(embed)
}

// SlackBlock represents a Slack block structure
type SlackBlock struct {
	Text   SlackText   `json:"text"`
	Type   string      `json:"type"`
	Color  string      `json:"color,omitempty"`
	Title  SlackText   `json:"title,omitempty"`
	Fields []SlackText `json:"fields,omitempty"`
}

// SlackText represents text in Slack blocks
type SlackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// SlackMessage represents a Slack message payload
type SlackMessage struct {
	Text        string       `json:"text,omitempty"`
	Username    string       `json:"username,omitempty"`
	Attachments []SlackBlock `json:"attachments,omitempty"`
}

// BuildSlackVariantMessage builds a Slack message from variant config and dynamic values
func BuildSlackVariantMessage(variant *VariantConfig, title string, description string, color string) ([]byte, error) {
	msg := SlackMessage{}

	// Use variant username if provided
	if variant != nil && variant.Username != "" {
		msg.Username = variant.Username
	}

	// Use variant content if provided
	if variant != nil && variant.Content != "" {
		msg.Text = variant.Content
	}

	if title != "" || description != "" || color != "" {
		attachment := SlackBlock{
			Type: "mrkdwn",
			Text: SlackText{
				Type: "mrkdwn",
				Text: description,
			},
		}

		if title != "" {
			attachment.Title = SlackText{
				Type: "plain_text",
				Text: title,
			}
		}

		// Use variant values as defaults
		if variant != nil {
			// Use variant color if provided and no color passed
			if variant.Color > 0 && color == "" {
				attachment.Color = fmt.Sprintf("#%06x", variant.Color)
			} else if color != "" {
				attachment.Color = color
			}

			// Use variant title as default if no title passed
			if variant.Title != "" && title == "" {
				attachment.Title = SlackText{
					Type: "plain_text",
					Text: variant.Title,
				}
			}

			// Use variant description as default if no description passed
			if variant.Description != "" && description == "" {
				attachment.Text.Text = variant.Description
			}
		}

		msg.Attachments = []SlackBlock{attachment}
	}

	return json.Marshal(msg)
}

// TelegramVariantMessage represents a Telegram message with formatting
type TelegramVariantMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// BuildTelegramVariantMessage builds a Telegram message from variant config and dynamic values
func BuildTelegramVariantMessage(variant *VariantConfig, title string, description string) ([]byte, error) {
	var text string

	// Use variant values as defaults, but allow override with passed parameters
	finalTitle := title
	finalDescription := description

	if variant != nil {
		// Use variant title as default if no title passed
		if variant.Title != "" && title == "" {
			finalTitle = variant.Title
		}
		// Use variant description as default if no description passed
		if variant.Description != "" && description == "" {
			finalDescription = variant.Description
		}
		// Use variant content if provided
		if variant.Content != "" {
			text = variant.Content
		}
	}

	// Build text from title and description if content not set
	if text == "" {
		if finalTitle != "" && finalDescription != "" {
			text = fmt.Sprintf("*%s*\n\n%s", finalTitle, finalDescription)
		} else if finalTitle != "" {
			text = fmt.Sprintf("*%s*", finalTitle)
		} else if finalDescription != "" {
			text = finalDescription
		}
	}

	msg := TelegramVariantMessage{
		Text:      text,
		ParseMode: "Markdown",
	}

	return json.Marshal(msg)
}
