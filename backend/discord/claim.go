package discord

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/drops"
)

// ClaimHandler handles the /claim command.
type ClaimHandler struct {
	store collection.DB
}

// claimOptions holds the parsed options for the claim command.
type claimOptions struct {
	name string
}

func parseClaimOptions(cmd CommandContext) (claimOptions, error) {
	name, err := cmd.OptString("name")
	if err != nil || name == "" {
		return claimOptions{}, errors.New("enter a name to claim the character")
	}
	return claimOptions{name: name}, nil
}

// Claim processes a character claim.
func (h *ClaimHandler) Claim(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	logger := slog.With("user_id", cmd.UserID(), "channel_id", cmd.ChannelID(), "guild_id", cmd.GuildID())

	opts, err := parseClaimOptions(cmd)
	if err != nil {
		logger.Debug("claim: no name provided")
		w.Respond(rspErr(err.Error()))
		return
	}

	drop, err := collection.Claim(ctx, h.store, cmd.UserID(), cmd.ChannelID(), opts.name)
	if err != nil {
		logger.Debug("failed to claim", "error", err)
		switch {
		case errors.Is(err, drops.ErrNoDropInChannel):
			w.Respond(rspErr("No character to claim in this channel. Wait for the next drop!"))
		case errors.Is(err, drops.ErrWrongCharacterName):
			w.Respond(rspErr("Wrong name! Check the hint and try again."))
		case errors.Is(err, collection.ErrAlreadyOwned):
			w.Respond(rspErr("You already have this character in your collection!"))
		default:
			w.Respond(rspErr("Failed to claim character"))
		}
		return
	}

	logger.Info("user claimed character",
		"character_id", drop.Character.ID,
		"character_name", drop.Character.Name)

	rarity := collection.RarityFromPopularity(drop.Character.Popularity)

	w.Respond(claimEmbed(drop.Character, rarity.String()))
}
