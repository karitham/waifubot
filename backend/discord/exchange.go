package discord

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
)

func (b *Bot) exchange(m *corde.Mux) {
	m.SlashCommand("", wrap(
		b.exchangeCommand,
		trace[corde.SlashCommandInteractionData],
		interact(b.InterStore, onInteraction[corde.SlashCommandInteractionData](b)),
	))
	m.Autocomplete("id", b.userCollectionAutocomplete)
}

func (b *Bot) exchangeCommand(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	logger := slog.With("user_id", uint64(i.Member.User.ID), "guild_id", uint64(i.GuildID))

	user := i.Member.User
	charID, _ := i.Data.Options.Int64("id")

	char, err := collection.Exchange(ctx, b.Store, user.ID, charID)
	if err != nil {
		if errors.Is(err, collection.ErrUserDoesNotOwnCharacter) {
			w.Respond(rspErr("You don't own that character"))
			return
		}
		logger.Error("error performing exchange", "error", err, "character_id", charID)
		w.Respond(newErrf("Error: %s", err.Error()))
		return
	}

	w.Respond(newErrf("Exchanged %s for a token!", char.Name))
}
