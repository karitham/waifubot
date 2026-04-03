package discord

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
)

// GiveHandler handles the /give command and its autocomplete.
type GiveHandler struct {
	store collection.Store
}

// giveOptions holds the parsed options for the give command.
type giveOptions struct {
	recipientID uint64
	recipient   corde.User
	charID      int64
}

func parseGiveOptions(cmd CommandContext) (giveOptions, error) {
	user, err := cmd.OptUser("user")
	if err != nil {
		return giveOptions{}, fmt.Errorf("select a user to give to: %w", err)
	}
	charID, err := cmd.OptInt64("id")
	if err != nil {
		return giveOptions{}, fmt.Errorf("select a character to give: %w", err)
	}
	return giveOptions{
		recipientID: uint64(user.ID),
		recipient:   user,
		charID:      charID,
	}, nil
}

// Register wires the give sub-routes on the mux.
func (h *GiveHandler) Register(m *corde.Mux) {
	m.SlashCommand("", wrap(
		wrapCtx(h.Give),
		trace[corde.SlashCommandInteractionData],
	))
	m.Autocomplete("id", h.Autocomplete)
}

// Give transfers a character to another user.
func (h *GiveHandler) Give(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	logger := slog.With("user_id", cmd.UserID(), "guild_id", cmd.GuildID())

	opts, err := parseGiveOptions(cmd)
	if err != nil {
		logger.Debug("give command: no user selected")
		w.Respond(rspErr(err.Error()))
		return
	}
	logger.Debug("giving character", "to_user_id", opts.recipientID, "character_id", opts.charID)

	char, err := collection.Give(ctx, h.store, cmd.UserID(), opts.recipientID, opts.charID)
	if err != nil {
		if errors.Is(err, collection.ErrUserDoesNotOwnCharacter) {
			w.Respond(rspErr("You don't own that character"))
			return
		}
		logger.Error("error performing give", "error", err, "to_user_id", opts.recipientID, "character_id", opts.charID)
		w.Respond(Privf("Failed to give character"))
		return
	}

	w.Respond(corde.NewResp().Contentf("Gave %s (%d) to %s", char.Name, opts.charID, opts.recipient.Username))
}

// Autocomplete provides character suggestions for the give command.
func (h *GiveHandler) Autocomplete(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.AutocompleteInteractionData]) {
	autocomplete(ctx, w, i, "id", func(ctx context.Context, input string) ([]catalog.Character, error) {
		return h.store.SearchCharacters(ctx, uint64(i.Member.User.ID), input)
	}, formatCharacterChoice)
}
