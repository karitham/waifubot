package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DiscordUser represents a Discord user object
type DiscordUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

// Client represents a Discord API client
type Client struct {
	httpClient *http.Client
	token      string
}

// NewClient creates a new Discord API client
func NewClient(token string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		token: token,
	}
}

// GetUser fetches a Discord user by ID
func (c *Client) GetUser(ctx context.Context, userID string) (*DiscordUser, error) {
	url := fmt.Sprintf("https://discord.com/api/v10/users/%s", userID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bot "+c.token)
	req.Header.Set("User-Agent", "WaifuBot (https://github.com/karitham/waifubot)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close() // Ignore error as we're just cleaning up
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discord API returned status %d", resp.StatusCode)
	}

	var user DiscordUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &user, nil
}

func DiscordAvatarURL(userID uint64, avatar string) string {
	return fmt.Sprintf("https://cdn.discordapp.com/avatars/%d/%s.png", userID, avatar)
}
