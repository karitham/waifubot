package collection

import (
	"context"
	"slices"
	"time"

	"github.com/karitham/waifubot/catalog"
)

// withTx runs fn in a transaction, committing on success and rolling back on
// failure. It is the transaction shape for writes that do not materialize or
// lock the user; withUsers builds on it. The returned error is the one
// reported; a rollback failure is not.
func withTx(ctx context.Context, store DB, fn func(tx Tx) error) error {
	tx, err := store.Begin(ctx)
	if err != nil {
		return err
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

// withUsers runs fn in a transaction after materializing and locking the
// given users. It is the transaction shape for domain writes: the users an
// operation acts on exist for the write and roll back with it, and their
// rows are locked for its duration. IDs are sorted and deduped so concurrent
// operations acquire the locks in the same order; closures must act only on
// the users passed here for that order to hold.
func withUsers(ctx context.Context, store DB, userIDs []UserID, fn func(tx Tx) error) error {
	for _, userID := range userIDs {
		if !userID.Valid() {
			return ErrInvalidUserID
		}
	}

	users := slices.Clone(userIDs)
	slices.Sort(users)
	users = slices.Compact(users)

	return withTx(ctx, store, func(tx Tx) error {
		for _, userID := range users {
			if err := tx.CreateUser(ctx, userID); err != nil {
				return err
			}
			if err := tx.LockUser(ctx, userID); err != nil {
				return err
			}
		}
		return fn(tx)
	})
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

// SetDiscordInfo records the user's fetched Discord profile, materializing
// the user first.
func SetDiscordInfo(ctx context.Context, store DB, userID UserID, username, avatar string, fetchedAt time.Time) error {
	return withUsers(ctx, store, []UserID{userID}, func(tx Tx) error {
		return tx.UpdateDiscordInfo(ctx, userID, username, avatar, fetchedAt)
	})
}
