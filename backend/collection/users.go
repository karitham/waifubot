package collection

import (
	"context"
	"errors"
	"slices"

	"github.com/karitham/waifubot/catalog"
)

// ensureUser materializes a user: it creates the users row when it is absent.
// It is the only path that creates users. Create is idempotent, so losing a
// create race is not an error.
func ensureUser(ctx context.Context, store Store, userID UserID) error {
	_, err := store.GetUser(ctx, userID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, ErrNotFound) {
		return err
	}

	return store.CreateUser(ctx, userID)
}

// withUsers runs fn in a transaction after materializing the given users.
// It is the transaction shape for domain writes: the users an operation acts
// on exist for the write and roll back with it. IDs are sorted and deduped so
// concurrent operations lock the same users in the same order.
func withUsers(ctx context.Context, store DB, userIDs []UserID, fn func(tx Tx) error) error {
	users := slices.Clone(userIDs)
	slices.Sort(users)
	users = slices.Compact(users)

	tx, err := store.Begin(ctx)
	if err != nil {
		return err
	}

	for _, userID := range users {
		if err := ensureUser(ctx, tx, userID); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return nil
}

// SetQuote sets the user's quote, materializing the user first.
func SetQuote(ctx context.Context, store DB, userID UserID, quote string) error {
	return withUsers(ctx, store, []UserID{userID}, func(tx Tx) error {
		return tx.UpdateQuote(ctx, userID, quote)
	})
}

// SetAnilistURL sets the user's AniList URL, materializing the user first.
func SetAnilistURL(ctx context.Context, store DB, userID UserID, url string) error {
	return withUsers(ctx, store, []UserID{userID}, func(tx Tx) error {
		return tx.UpdateAnilistURL(ctx, userID, url)
	})
}

// SetFavorite sets the user's favorite character; a zero charID clears it.
// The character must be owned when setting.
func SetFavorite(ctx context.Context, store DB, userID UserID, charID catalog.CharacterID) error {
	return withUsers(ctx, store, []UserID{userID}, func(tx Tx) error {
		return tx.SetFavorite(ctx, userID, charID)
	})
}
