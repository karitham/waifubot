package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
)

func indexMiddleware[T corde.InteractionDataConstraint](b *Bot) func(func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T])) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
	return func(next func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T])) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
		return func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
			go func() {
				if err := b.GuildIndexer.IndexGuildIfNeeded(context.Background(), i.GuildID); err != nil {
					slog.Error("failed to index guild", "error", err, "guild_id", i.GuildID)
				}
			}()

			next(ctx, w, i)
		}
	}
}

func (b *Bot) holders(m *corde.Mux) {
	m.SlashCommand("", wrap(
		b.holdersCommand,
		indexMiddleware[corde.SlashCommandInteractionData](b),
		trace[corde.SlashCommandInteractionData],
		interact(b.InterStore, onInteraction[corde.SlashCommandInteractionData](b)),
	))
	m.Autocomplete("id", b.verifyAutocomplete)
}

func (b *Bot) holdersCommand(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	charID, errCharOK := i.Data.Options.Int("id")
	if errCharOK != nil {
		w.Respond(rspErr("select a character to check"))
		return
	}

	charName, holderIDs, err := collection.CharacterHolders(ctx, b.Store, i.GuildID, int64(charID))
	if err != nil {
		w.Respond(newErrf("Error: %s", err.Error()))
		return
	}

	if len(holderIDs) == 0 {
		w.Respond(corde.NewResp().Contentf("No one in this server has **%s** (ID: %d)", charName, charID).Ephemeral())
		return
	}

	var mentions strings.Builder

	mentions.WriteString(fmt.Sprintf("Users in this server who have **%s** (ID: %d):\n", charName, charID))
	for _, holderID := range holderIDs {
		mentions.WriteString(fmt.Sprintf("- <@%d>\n", holderID))
	}

	w.Respond(corde.NewResp().Content(mentions.String()).Ephemeral())
}

func FetchGuildMemberIDs(ctx context.Context, botToken string, guildID corde.Snowflake) ([]corde.Snowflake, error) {
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
	defer resp.Body.Close() //nolint: errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch guild members: status %d", resp.StatusCode)
	}

	var members []corde.Member
	if err := json.NewDecoder(resp.Body).Decode(&members); err != nil {
		return nil, fmt.Errorf("failed to decode members: %w", err)
	}

	return members, nil
}
