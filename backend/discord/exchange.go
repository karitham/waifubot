package discord

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
)

// followUpSender posts a follow-up message to the invoking interaction.
type followUpSender interface {
	FollowUpInteraction(token string, data corde.InteractionResponder) error
}

// exchangeCustomID ties a button to the user who rolled and the character they got.
func exchangeCustomID(owner collection.UserID, charID catalog.CharacterID) string {
	return fmt.Sprintf("exchange:%d:%d", owner, charID)
}

// exchangePrompt is the ephemeral follow-up shown only to the rolling user.
func exchangePrompt(char catalog.Character, owner collection.UserID) *corde.RespB {
	return corde.NewResp().
		Contentf("Exchange %s for %d token?", char.Name, collection.SellReward().Int32()).
		ActionRow(corde.Component{
			Type:     corde.COMPONENT_BUTTON,
			Style:    corde.BUTTON_SUCCESS,
			Label:    "Exchange",
			CustomID: exchangeCustomID(owner, char.ID),
		}).
		Ephemeral()
}

// Exchange handles clicks on the exchange button of a roll's ephemeral prompt.
func (h *RollHandler) Exchange(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.ButtonInteractionData]) {
	owner, charID, err := parseExchangeCustomID(i.Data.CustomID)
	if err != nil {
		slog.Warn("invalid exchange button", "error", err)
		w.Respond(Privf("This exchange button is no longer valid."))
		return
	}

	caller := callerUserID(i.Member, i.User)
	if caller != owner {
		w.Respond(Privf("Only the user who rolled this character can exchange it."))
		return
	}

	char, err := collection.Exchange(ctx, h.store, caller, charID)
	if err != nil {
		if errors.Is(err, collection.ErrUserDoesNotOwnCharacter) {
			w.Respond(Privf("You don't own that character anymore."))
			return
		}

		slog.Error("error exchanging rolled character", "error", err, "user_id", caller, "character_id", charID)
		w.Respond(Privf("Failed to exchange character."))
		return
	}

	confirmation := fmt.Sprintf("Sold %s for %d token", char.Name, collection.SellReward().Int32())
	w.DeferedUpdate()

	token := i.Token
	go func() {
		if err := h.cleaner.ClearComponents(token, confirmation); err != nil {
			slog.Warn("error removing exchange button", "error", err)
		}
	}()
}

func parseExchangeCustomID(customID string) (collection.UserID, catalog.CharacterID, error) {
	parts := strings.Split(customID, ":")
	if len(parts) != 3 || parts[0] != "exchange" {
		return 0, 0, fmt.Errorf("invalid exchange custom ID %q", customID)
	}

	owner, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid exchange owner %q: %w", parts[1], err)
	}

	charID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid exchange character %q: %w", parts[2], err)
	}

	return collection.UserID(owner), catalog.CharacterID(charID), nil
}
