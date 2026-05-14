package slack

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
)

type Client struct {
	webhookURL string
	channel    string
}

func NewClient(webhookURL, channel string) *Client {
	return &Client{
		webhookURL: webhookURL,
		channel:    channel,
	}
}

func (c *Client) SendAlert(ctx context.Context, channel string, title string, fields map[string]string) error {
	if c.webhookURL == "" {
		return fmt.Errorf("slack webhook URL is not configured")
	}

	targetChannel := channel
	if targetChannel == "" {
		targetChannel = c.channel
	}

	attachmentFields := make([]slack.AttachmentField, 0, len(fields))
	for k, v := range fields {
		attachmentFields = append(attachmentFields, slack.AttachmentField{
			Title: k,
			Value: v,
			Short: true,
		})
	}

	msg := &slack.WebhookMessage{
		Channel: targetChannel,
		Attachments: []slack.Attachment{
			{
				Color:   "#ff0000",
				Pretext: title,
				Fields:  attachmentFields,
				Footer:  "IDP Core Budget Alert System",
			},
		},
	}

	return slack.PostWebhookContext(ctx, c.webhookURL, msg)
}

func (c *Client) Channel() string {
	return c.channel
}