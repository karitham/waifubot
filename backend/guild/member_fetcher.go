package guild

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Karitham/corde"
)

// DiscordFetcher fetches guild members from the Discord API.
type DiscordFetcher struct {
	botToken   string
	httpClient *http.Client
}

// NewDiscordFetcher creates a new Discord member fetcher.
func NewDiscordFetcher(botToken string) *DiscordFetcher {
	return &DiscordFetcher{
		botToken:   botToken,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchMemberIDs fetches all member IDs from a Discord guild.
func (f *DiscordFetcher) FetchMemberIDs(ctx context.Context, guildID corde.Snowflake) ([]corde.Snowflake, error) {
	var allMemberIDs []corde.Snowflake
	after := corde.Snowflake(0)

	for {
		url := fmt.Sprintf("https://discord.com/api/v10/guilds/%d/members?limit=1000", guildID)
		if after != 0 {
			url = fmt.Sprintf("%s&after=%d", url, after)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Authorization", "Bot "+f.botToken)

		resp, err := f.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch guild members: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to fetch guild members: status %d", resp.StatusCode)
		}

		var members []corde.Member
		if err := json.NewDecoder(resp.Body).Decode(&members); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode members: %w", err)
		}
		resp.Body.Close()

		if len(members) == 0 {
			break
		}

		for _, member := range members {
			allMemberIDs = append(allMemberIDs, member.User.ID)
		}

		if len(members) < 1000 {
			break
		}

		after = members[len(members)-1].User.ID
	}

	return allMemberIDs, nil
}
