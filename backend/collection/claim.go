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
func Claim(ctx context.Context, store Store, userID UserID, channelID ChannelID, charName string) (catalog.Character, error) {
	tx, err := store.WithTx(ctx)
	if err != nil {
		return catalog.Character{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	drop, err := tx.GetDropForUpdate(ctx, channelID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return catalog.Character{}, ErrNoDropInChannel
		}
		return catalog.Character{}, fmt.Errorf("failed to get drop: %w", err)
	}

	if !strings.EqualFold(SanitizeName(drop.Name), charName) {
		return catalog.Character{}, ErrWrongCharacterName
	}

	// Ensure user exists
	_, err = tx.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			if err = tx.CreateUser(ctx, userID); err != nil {
				return catalog.Character{}, fmt.Errorf("failed to create user: %w", err)
			}
		} else {
			return catalog.Character{}, fmt.Errorf("failed to get user: %w", err)
		}
	}

	err = tx.UpsertCharacter(ctx, drop)
	if err != nil {
		return catalog.Character{}, fmt.Errorf("failed to upsert character: %w", err)
	}

	now := time.Now()
	err = tx.AddToCollection(ctx, userID, drop.ID, SourceClaim, now)
	if err != nil {
		if errors.Is(err, ErrAlreadyOwned) {
			return catalog.Character{}, ErrAlreadyOwned
		}
		return catalog.Character{}, fmt.Errorf("failed to insert character into collection: %w", err)
	}

	_ = tx.RemoveFromWishlist(ctx, userID, drop.ID)

	err = tx.DeleteDrop(ctx, channelID)
	if err != nil {
		return catalog.Character{}, fmt.Errorf("failed to delete drop: %w", err)
	}

	err = tx.Commit(ctx)
	committed = err == nil
	if err != nil {
		return catalog.Character{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return drop, nil
}

// SanitizeName collapses whitespace in a character name to a single space.
func SanitizeName(name string) string {
	return strings.Join(strings.Fields(name), " ")
}
