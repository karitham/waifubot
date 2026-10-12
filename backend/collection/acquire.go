package collection

import (
	"context"
	"errors"
	"time"

	"github.com/karitham/waifubot/catalog"
)

// acquisition describes how a character enters a collection: how an unowned
// character is picked, what the user must have to acquire it, and how the
// cost is applied. Roll and SeriesRoll are the two implementations.
type acquisition struct {
	// source labels the collection entry (ROLL, SERIES_ROLL).
	source Source
	// check reports why the user cannot acquire the character right now,
	// e.g. an active cooldown or insufficient tokens. It runs once before
	// picking and again inside the transaction, both with the acquisition's
	// single timestamp.
	check func(user User, now time.Time) error
	// pick selects an unowned character, adding it to the catalog first
	// when it comes from an external source.
	pick func(ctx context.Context, store Store, userID UserID) (catalog.Character, error)
	// charge applies the cost of the acquisition inside the transaction:
	// tokens for a paid roll, the next cooldown for a free one. It is the
	// authoritative gate under concurrency; check only narrows the window.
	charge func(ctx context.Context, tx Store, now time.Time) error
}

// acquire runs an acquisition: check eligibility, pick an unowned character,
// then add it to the collection and apply the acquisition's cost in one
// transaction.
//
// The check runs twice on purpose: the first call fails fast before picking,
// the second narrows the race window inside the transaction. Neither call is
// a lock; the charge is the authoritative gate, because both costs are
// applied by conditional updates (SpendTokens, RecordRoll).
func acquire(ctx context.Context, store DB, userID UserID, a acquisition) (catalog.Character, error) {
	now := time.Now()

	user, err := store.GetUser(ctx, userID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return catalog.Character{}, err
	}

	// A missing user is the zero User: checks that reject a new user
	// (insufficient tokens) fail here, checks that pass (cooldown) proceed.
	if err := a.check(user, now); err != nil {
		return catalog.Character{}, err
	}

	char, err := a.pick(ctx, store, userID)
	if err != nil {
		return catalog.Character{}, err
	}

	err = withTx(ctx, store, func(tx Tx) error {
		user, err := ensureUser(ctx, tx, userID)
		if err != nil {
			return err
		}

		if err := a.check(user, now); err != nil {
			return err
		}

		if err := tx.AddToCollection(ctx, userID, char.ID, a.source, now); err != nil {
			return err
		}

		if err := tx.RemoveFromWishlist(ctx, userID, char.ID); err != nil {
			return err
		}

		return a.charge(ctx, tx, now)
	})
	if err != nil {
		return catalog.Character{}, err
	}

	return char, nil
}

// ensureUser fetches the user, creating the row when it does not exist yet.
func ensureUser(ctx context.Context, store Store, userID UserID) (User, error) {
	user, err := store.GetUser(ctx, userID)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return User{}, err
	}

	if err := store.CreateUser(ctx, userID); err != nil {
		return User{}, err
	}

	return store.GetUser(ctx, userID)
}
