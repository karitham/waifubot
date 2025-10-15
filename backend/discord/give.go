package discord

import (
	"context"

	"github.com/Karitham/corde"
	"github.com/rs/zerolog/log"
)

func (b *Bot) give(m *corde.Mux) {
	m.SlashCommand("", wrap(
		b.giveCommand,
		trace[corde.SlashCommandInteractionData],
		interact(b.Inter, onInteraction[corde.SlashCommandInteractionData](b)),
	))
	m.Autocomplete("id", b.userCollectionAutocomplete)
}

func (b *Bot) giveCommand(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	user, errUserOK := i.Data.OptionsUser("user")
	if errUserOK != nil {
		w.Respond(rspErr("select a user to give to"))
		return
	}
	charID, errCharOK := i.Data.Options.Int("id")
	if errCharOK != nil {
		w.Respond(rspErr("select a character to give"))
		return
	}
	log.Ctx(ctx).Trace().Stringer("src", i.Member.User.ID).Stringer("dst", user.ID).Int("charID", charID).Send()

	c, err := b.Store.VerifyChar(ctx, i.Member.User.ID, int64(charID))
	if err != nil {
		w.Respond(newErrf("Error giving character %d, not in your collection.", charID))
		return
	}

	_, err = b.Store.VerifyChar(ctx, user.ID, int64(charID))
	if err == nil {
		w.Respond(newErrf("%s already has this character in their collection.", user.Username))
		return
	}

	err = b.Store.GiveUserChar(ctx, user.ID, i.Member.User.ID, int64(charID))
	if err != nil {
		w.Respond(newErrf("Error giving %s (%d) to %s", c.Name, charID, user.Username))
		return
	}

	w.Respond(corde.NewResp().Contentf("Gave %s (%d) to %s", c.Name, charID, user.Username))
}
