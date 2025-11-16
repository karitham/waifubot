package discord

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
)

func (b *Bot) verify(m *corde.Mux) {
	m.SlashCommand("", wrap(
		b.verifyCommand,
		trace[corde.SlashCommandInteractionData],
		indexMiddleware[corde.SlashCommandInteractionData](b),
		interact(b.InterStore, onInteraction[corde.SlashCommandInteractionData](b)),
	))
	m.Autocomplete("id", b.verifyAutocomplete)
}

func (b *Bot) verifyCommand(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	user := i.Member.User
	if len(i.Data.Resolved.Members) > 0 {
		user = i.Data.Resolved.Users.First()
	}
	charID, _ := i.Data.Options.Int64("id")

	has, char, err := collection.CheckOwnership(ctx, b.Store, user.ID, charID)
	if err != nil {
		w.Respond(newErrf("Error checking ownership: %s", err.Error()))
		return
	}

	if has {
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

	chars, err := collection.SearchGlobalCharacters(ctx, b.Store, id)
	if err != nil {
		slog.Error("Error searching global characters", "error", err)
		return
	}

	resp := corde.NewResp()
	for _, c := range chars {
		resp.Choice(fmt.Sprintf("%s (%d)", c.Name, c.ID), c.ID)
	}

	w.Autocomplete(resp)
}
