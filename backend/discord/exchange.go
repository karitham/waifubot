package discord

import (
	"context"

	"github.com/Karitham/corde"
)

func (b *Bot) exchange(m *corde.Mux) {
	m.SlashCommand("", wrap(
		b.exchangeCommand,
		trace[corde.SlashCommandInteractionData],
		interact(b.Inter, onInteraction[corde.SlashCommandInteractionData](b)),
	))
	m.Autocomplete("id", b.profileEditFavoriteComplete)
}

func (b *Bot) exchangeCommand(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	user := i.Member.User
	char, _ := i.Data.Options.Int64("id")
	var c Character

	if err := b.Store.Tx(ctx, func(s Store) error {
		var err error
		c, err = s.DeleteChar(ctx, user.ID, char)
		if err != nil {
			w.Respond(newErrf("%s doesn't own this character; you can't exchange it!", user.Username))
			return err
		}

		return s.AddDropToken(ctx, user.ID)
	}); err != nil {
		return
	}

	w.Respond(newErrf("Good job %s, you just exchanged %s for a token!", user.Username, c.Name))
}
