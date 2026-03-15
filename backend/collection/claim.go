package collection

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Karitham/corde"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	collectionstore "github.com/karitham/waifubot/storage/collectionstore"
	"github.com/karitham/waifubot/storage/wishliststore"
)

var (
	ErrNoDropInChannel    = errors.New("no drop in this channel")
	ErrWrongCharacterName = errors.New("wrong character name")
)

func Claim(ctx context.Context, store Store, userID, channelID uint64, charName string) (Character, error) {
	tx, err := store.Tx(ctx)
	if err != nil {
		return Character{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	drop, err := tx.DropStore().GetDropForUpdate(ctx, channelID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Character{}, ErrNoDropInChannel
		}
		return Character{}, fmt.Errorf("failed to get drop: %w", err)
	}

	if !strings.EqualFold(drop.Name, charName) {
		return Character{}, ErrWrongCharacterName
	}

	_, err = tx.UserStore().Get(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = tx.UserStore().Create(ctx, userID)
			if err != nil {
				return Character{}, fmt.Errorf("failed to create user: %w", err)
			}
		} else {
			return Character{}, fmt.Errorf("failed to get user: %w", err)
		}
	}

	_, err = tx.CollectionStore().UpsertCharacter(ctx, collectionstore.UpsertCharacterParams{
		ID:    drop.ID,
		Name:  drop.Name,
		Image: drop.Image,
	})
	if err != nil {
		return Character{}, fmt.Errorf("failed to upsert character: %w", err)
	}

	now := time.Now()
	_, err = tx.CollectionStore().Insert(ctx, collectionstore.InsertParams{
		UserID:      userID,
		CharacterID: drop.ID,
		Source:      "CLAIM",
		AcquiredAt:  pgtype.Timestamp{Time: now, Valid: true},
	})
	if err != nil {
		return Character{}, fmt.Errorf("failed to insert character into collection: %w", err)
	}

	_ = tx.WishlistStore().RemoveCharacterFromWishlist(ctx, wishliststore.RemoveCharacterFromWishlistParams{
		UserID:      userID,
		CharacterID: drop.ID,
	})

	err = tx.DropStore().DeleteDrop(ctx, channelID)
	if err != nil {
		return Character{}, fmt.Errorf("failed to delete drop: %w", err)
	}

	err = tx.Commit(ctx)
	committed = err == nil
	if err != nil {
		return Character{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return Character{
		Date:   now,
		Image:  drop.Image,
		Name:   drop.Name,
		Type:   "CLAIM",
		UserID: corde.Snowflake(userID),
		ID:     drop.ID,
	}, nil
}
