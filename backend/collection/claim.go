package collection

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/karitham/waifubot/catalog"
)

// ErrNoDropInChannel is returned when there is no drop in the channel.
var ErrNoDropInChannel = errors.New("no drop in this channel")

// ErrWrongCharacterName is returned when the character name is wrong.
var ErrWrongCharacterName = errors.New("wrong character name")

// Claim claims a dropped character for a user.
func Claim(ctx context.Context, store DB, userID UserID, channelID ChannelID, charName string) (catalog.Character, error) {
	var drop catalog.Character
	err := withUsers(ctx, store, []UserID{userID}, func(tx Tx) error {
		var err error
		drop, err = tx.GetDropForUpdate(ctx, channelID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return ErrNoDropInChannel
			}
			return fmt.Errorf("failed to get drop: %w", err)
		}

		if !strings.EqualFold(SanitizeName(drop.Name), SanitizeName(charName)) {
			return ErrWrongCharacterName
		}

		if err := tx.UpsertCharacter(ctx, drop); err != nil {
			return fmt.Errorf("failed to upsert character: %w", err)
		}

		err = tx.AddToCollection(ctx, userID, drop.ID, SourceClaim, time.Now())
		if err != nil {
			if errors.Is(err, ErrAlreadyOwned) {
				return ErrAlreadyOwned
			}
			return fmt.Errorf("failed to insert character into collection: %w", err)
		}

		if err := tx.RemoveFromWishlist(ctx, userID, drop.ID); err != nil {
			return err
		}

		if err := tx.DeleteDrop(ctx, channelID); err != nil {
			return fmt.Errorf("failed to delete drop: %w", err)
		}
		return nil
	})
	if err != nil {
		return catalog.Character{}, err
	}

	return drop, nil
}

// SanitizeName collapses whitespace in a character name to a single space.
func SanitizeName(name string) string {
	return strings.Join(strings.Fields(name), " ")
}
