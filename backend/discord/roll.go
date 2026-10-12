package discord

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/wishlist"
)

// RollHandler handles the /roll command.
type RollHandler struct {
	store    collection.DB
	config   collection.Config
	wishlist wishlist.Store
}

// Roll performs a character roll.
func (h *RollHandler) Roll(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	logger := slog.With("user_id", cmd.UserID(), "guild_id", cmd.GuildID())

	char, err := collection.Roll(ctx, h.store, cmd.UserID(), h.config)

	var cd collection.ErrRollCooldown
	switch {
	case errors.As(err, &cd):
		w.Respond(rspErr(fmt.Sprintf("You can roll again in %s.", time.Until(cd.Until).Round(time.Second))))
		return
	case err != nil:
		logger.Error("error performing roll", "error", err)
		w.Respond(rspErr("An error occurred, please try again later"))
		return
	}

	wantingUsers, err := h.wishlist.GetUsersWantingCharacter(ctx, char.ID, cmd.GuildID(), cmd.UserID())
	if err != nil {
		logger.Error("error getting users wanting character", "error", err)
	}

	w.Respond(rollEmbed(char, formatUsersWantingCharacter(wantingUsers, cmd.UserID())))
}
