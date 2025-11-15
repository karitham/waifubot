package discord

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Karitham/corde"
	"github.com/rs/zerolog/log"
)

func (b *Bot) verify(m *corde.Mux) {
	m.SlashCommand("", wrap(
		b.verifyCommand,
		trace[corde.SlashCommandInteractionData],
		indexMiddleware[corde.SlashCommandInteractionData](b),
		interact(b.Inter, onInteraction[corde.SlashCommandInteractionData](b)),
	))
	m.Autocomplete("id", b.verifyAutocomplete)
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

func (b *Bot) verifyAutocomplete(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.AutocompleteInteractionData]) {
	id, err := i.Data.Options.String("id")
	if err != nil {
		i, _ := i.Data.Options.Int("id")
		id = strconv.Itoa(i)
	}

	chars, err := b.Store.GlobalCharsStartingWith(ctx, id)
	if err != nil {
		log.Err(err).Msg("Error searching global characters")
		return
	}
	if len(chars) > 25 {
		chars = chars[:25]
	}

	resp := corde.NewResp()
	for _, c := range chars {
		resp.Choice(fmt.Sprintf("%s (%d)", c.Name, c.ID), c.ID)
	}

	w.Autocomplete(resp)
}
