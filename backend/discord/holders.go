package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Karitham/corde"
)

func (b *Bot) holders(m *corde.Mux) {
	m.SlashCommand("", wrap(
		b.holdersCommand,
		trace[corde.SlashCommandInteractionData],
		interact(b.Inter, onInteraction[corde.SlashCommandInteractionData](b)),
	))
	m.Autocomplete("id", b.verifyAutocomplete)
}

func (b *Bot) holdersCommand(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	charID, errCharOK := i.Data.Options.Int("id")
	if errCharOK != nil {
		w.Respond(rspErr("select a character to check"))
		return
	}

	if i.GuildID == 0 {
		w.Respond(rspErr("this command can only be used in servers"))
		return
	}

	char, err := b.Store.GetCharByID(ctx, int64(charID))
	if err != nil {
		w.Respond(newErrf("no one in this server has %d", charID))
		return
	}

	memberIDs, err := fetchGuildMemberIDs(ctx, b.BotToken, i.GuildID)
	if err != nil {
		w.Respond(newErrf("failed to fetch guild members: %v", err))
		return
	}

	holderIDs, err := b.Store.UsersOwningCharFiltered(ctx, int64(charID), memberIDs)
	if err != nil {
		w.Respond(newErrf("failed to fetch character holders: %v", err))
		return
	}

	if len(holderIDs) == 0 {
		w.Respond(corde.NewResp().Contentf("No one in this server has **%s** (ID: %d)", char.Name, charID).Ephemeral())
		return
	}

	var mentions strings.Builder

	mentions.WriteString(fmt.Sprintf("Users in this server who have **%s** (ID: %d):\n", char.Name, charID))
	for _, holderID := range holderIDs {
		mentions.WriteString(fmt.Sprintf("- <@%d>\n", holderID))
	}

	w.Respond(corde.NewResp().Content(mentions.String()).Ephemeral())
}

func fetchGuildMemberIDs(ctx context.Context, botToken string, guildID corde.Snowflake) ([]corde.Snowflake, error) {
	var allMemberIDs []corde.Snowflake

	after := corde.Snowflake(0)

	for {
		members, err := fetchGuildMembersPage(ctx, botToken, guildID, after)
		if err != nil {
			return nil, err
		}

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

func fetchGuildMembersPage(ctx context.Context, botToken string, guildID, after corde.Snowflake) ([]corde.Member, error) {
	url := fmt.Sprintf("https://discord.com/api/v10/guilds/%d/members?limit=1000", guildID)
	if after != 0 {
		url = fmt.Sprintf("%s&after=%d", url, after)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bot "+botToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch guild members: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch guild members: status %d", resp.StatusCode)
	}

	var members []corde.Member
	if err := json.NewDecoder(resp.Body).Decode(&members); err != nil {
		return nil, fmt.Errorf("failed to decode members: %w", err)
	}

	return members, nil
}
