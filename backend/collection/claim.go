package collection

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrNoDropInChannel is returned when there is no drop in the channel.
var ErrNoDropInChannel = errors.New("no drop in this channel")

// ErrWrongCharacterName is returned when the character name is wrong.
var ErrWrongCharacterName = errors.New("wrong character name")

// Claim claims a dropped character for a user.
func Claim(ctx context.Context, store Store, userID, channelID uint64, charName string) (Character, error) {
	tx, err := store.WithTx(ctx)
	if err != nil {
		return Character{}, fmt.Errorf("failed to begin transaction: %w", err)
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
			return Character{}, ErrNoDropInChannel
		}
		return Character{}, fmt.Errorf("failed to get drop: %w", err)
	}

	if !strings.EqualFold(sanitizeName(drop.Name), charName) {
		return Character{}, ErrWrongCharacterName
	}

	// Ensure user exists
	_, err = tx.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			if err = tx.CreateUser(ctx, userID); err != nil {
				return Character{}, fmt.Errorf("failed to create user: %w", err)
			}
		} else {
			return Character{}, fmt.Errorf("failed to get user: %w", err)
		}
	}

	err = tx.UpsertCharacter(ctx, Character{
		ID:    drop.ID,
		Name:  drop.Name,
		Image: drop.Image,
	})
	if err != nil {
		return Character{}, fmt.Errorf("failed to upsert character: %w", err)
	}

	now := time.Now()
	err = tx.AddToCollection(ctx, userID, Character(drop), "CLAIM", now)
	if err != nil {
		if errors.Is(err, ErrAlreadyOwned) {
			return Character{}, ErrAlreadyOwned
		}
		return Character{}, fmt.Errorf("failed to insert character into collection: %w", err)
	}

	_ = tx.RemoveFromWishlist(ctx, userID, drop.ID)

	err = tx.DeleteDrop(ctx, channelID)
	if err != nil {
		return Character{}, fmt.Errorf("failed to delete drop: %w", err)
	}

	err = tx.Commit(ctx)
	committed = err == nil
	if err != nil {
		return Character{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return Character(drop), nil
}

func sanitizeName(name string) string {
	return strings.Join(strings.Fields(name), " ")
}
