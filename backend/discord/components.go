package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Karitham/corde"
)

const discordAPIURL = "https://discord.com/api/v10"

// messageCleaner strips components from the message a button was clicked on.
// corde serializes components with omitempty, so its response builder cannot
// send the empty array Discord needs to remove the button.
type messageCleaner interface {
	ClearComponents(token string, content string) error
}

type discordMessageCleaner struct {
	api      string
	appID    corde.Snowflake
	botToken string
	client   *http.Client
}

func newDiscordMessageCleaner(m *corde.Mux) discordMessageCleaner {
	return discordMessageCleaner{
		api:      discordAPIURL,
		appID:    m.AppID,
		botToken: m.BotToken,
		client:   m.Client,
	}
}

func (c discordMessageCleaner) ClearComponents(token string, content string) error {
	payload, err := json.Marshal(struct {
		Content    string            `json:"content"`
		Components []corde.Component `json:"components"`
	}{Content: content, Components: []corde.Component{}})
	if err != nil {
		return fmt.Errorf("marshal message edit: %w", err)
	}

	url := fmt.Sprintf("%s/webhooks/%d/%s/messages/@original", c.api, c.appID, token)
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create message edit: %w", err)
	}
	req.Header.Set("authorization", "Bot "+c.botToken)
	req.Header.Set("content-type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("edit follow-up message: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("edit follow-up message: status %d: %s", resp.StatusCode, body)
	}
	return nil
}
