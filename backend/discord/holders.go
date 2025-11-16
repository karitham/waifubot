package discord

import (
	"context"
	"fmt"
	"log/slog"
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
