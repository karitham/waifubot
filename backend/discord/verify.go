package discord

import (
	"context"

	"github.com/Karitham/corde"
)

func (b *Bot) verify(m *corde.Mux) {
	m.SlashCommand("", wrap(
		b.verifyCommand,
		trace[corde.SlashCommandInteractionData],
		interact(b.Inter, onInteraction[corde.SlashCommandInteractionData](b)),
	))
	m.Autocomplete("id", b.profileEditFavoriteComplete)
}

func (b *Bot) verifyCommand(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	user := i.Member.User
	if len(i.Data.Resolved.Members) > 0 {
		user = i.Data.Resolved.Users.First()
	}
	charOpt, _ := i.Data.Options.Int64("id")

	char, _ := b.Store.VerifyChar(ctx, user.ID, charOpt)
	if char.ID == charOpt {
		w.Respond(newErrf("%s has %s in their collection", user.Username, char.Name))
		return
	}

	w.Respond(newErrf("%s doesn't have this character", user.Username))
}
