package discord

import (
	"context"
	"log/slog"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
)

func (b *Bot) give(m *corde.Mux) {
	m.SlashCommand("", wrap(
		b.giveCommand,
		trace[corde.SlashCommandInteractionData],
		interact(b.InterStore, onInteraction[corde.SlashCommandInteractionData](b)),
	))
	m.Autocomplete("id", b.userCollectionAutocomplete)
}

func (b *Bot) giveCommand(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	logger := slog.With("user_id", uint64(i.Member.User.ID), "guild_id", uint64(i.GuildID))

	user, errUserOK := i.Data.OptionsUser("user")
	if errUserOK != nil {
		logger.Debug("give command: no user selected")
		w.Respond(rspErr("select a user to give to"))
		return
	}
	charID, errCharOK := i.Data.Options.Int("id")
	if errCharOK != nil {
		logger.Debug("give command: no character selected")
		w.Respond(rspErr("select a character to give"))
		return
	}
	logger.Debug("giving character", "to_user_id", uint64(user.ID), "character_id", charID)

	char, err := collection.Give(ctx, b.Store, i.Member.User.ID, user.ID, int64(charID))
	if err != nil {
		logger.Error("error performing give", "error", err, "to_user_id", uint64(user.ID), "character_id", charID)
		w.Respond(newErrf("Error: %s", err.Error()))
		return
	}

	w.Respond(corde.NewResp().Contentf("Gave %s (%d) to %s", char.Name, charID, user.Username))
}
