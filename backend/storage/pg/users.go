package pg

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage"
)

func (b *base) GetUser(ctx context.Context, userID collection.UserID) (collection.User, error) {
	u, err := b.q.GetUser(ctx, uint64(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return collection.User{}, collection.ErrNotFound
		}
		return collection.User{}, err
	}
	return toUser(u), nil
}

// CreateUser creates the user when it is absent. It is idempotent.
func (t *Tx) CreateUser(ctx context.Context, userID collection.UserID) error {
	return t.q.CreateUser(ctx, uint64(userID))
}

// LockUser locks the user's row for the rest of the transaction. It returns
// ErrNotFound when the user does not exist.
func (t *Tx) LockUser(ctx context.Context, userID collection.UserID) error {
	_, err := t.q.LockUser(ctx, uint64(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return collection.ErrNotFound
		}
		return err
	}
	return nil
}

func (b *base) GetUserByAnilist(ctx context.Context, anilistURL string) (collection.User, error) {
	u, err := b.q.GetByAnilist(ctx, anilistURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return collection.User{}, collection.ErrNotFound
		}
		return collection.User{}, err
	}
	return toUser(u), nil
}

func (b *base) GetUserByDiscordUsername(ctx context.Context, username string) (collection.User, error) {
	u, err := b.q.GetByDiscordUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return collection.User{}, collection.ErrNotFound
		}
		return collection.User{}, err
	}
	return toUser(u), nil
}

func (t *Tx) RecordRoll(ctx context.Context, userID collection.UserID, now time.Time, cooldown time.Duration) error {
	rows, err := t.q.RecordRoll(ctx, storage.RecordRollParams{
		Date:      pgtype.Timestamp{Time: now.UTC(), Valid: true},
		UserID:    uint64(userID),
		Threshold: pgtype.Timestamp{Time: now.Add(-cooldown).UTC(), Valid: true},
	})
	if err != nil {
		return err
	}
	if rows > 0 {
		return nil
	}

	// The cooldown is still active. The user row exists (the acquisition
	// ensures it inside the transaction), and this read sees the winner's
	// committed timestamp.
	u, err := t.q.GetUser(ctx, uint64(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return collection.ErrNotFound
		}
		return err
	}
	return collection.ErrRollCooldown{Until: u.Date.Time.Add(cooldown)}
}

func (t *Tx) SpendTokens(ctx context.Context, userID collection.UserID, amount collection.TokenAmount) (collection.User, error) {
	if !amount.Valid() {
		return collection.User{}, collection.ErrInvalidAmount
	}

	u, err := t.q.SpendTokens(ctx, storage.SpendTokensParams{
		Tokens: amount.Int32(),
		UserID: uint64(userID),
	})
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return collection.User{}, err
		}

		// No row updated: either the user does not exist or the balance
		// is too low. Distinguish so callers can report the right thing.
		if _, getErr := t.q.GetUser(ctx, uint64(userID)); getErr != nil {
			if errors.Is(getErr, pgx.ErrNoRows) {
				return collection.User{}, collection.ErrNotFound
			}
			return collection.User{}, getErr
		}
		return collection.User{}, collection.ErrInsufficientTokens
	}
	return toUser(u), nil
}

func (t *Tx) AddTokens(ctx context.Context, userID collection.UserID, amount collection.TokenAmount) (collection.User, error) {
	if !amount.Valid() {
		return collection.User{}, collection.ErrInvalidAmount
	}

	u, err := t.q.UpdateTokens(ctx, storage.UpdateTokensParams{
		Tokens: amount.Int32(),
		UserID: uint64(userID),
	})
	if err != nil {
		return collection.User{}, err
	}
	return toUser(u), nil
}

func (t *Tx) UpdateQuote(ctx context.Context, userID collection.UserID, quote string) error {
	return t.q.UpdateQuote(ctx, storage.UpdateQuoteParams{
		Quote:  quote,
		UserID: uint64(userID),
	})
}

func (t *Tx) UpdateAnilistURL(ctx context.Context, userID collection.UserID, url string) error {
	return t.q.UpdateAnilistURL(ctx, storage.UpdateAnilistURLParams{
		AnilistUrl: url,
		UserID:     uint64(userID),
	})
}

func (t *Tx) UpdateDiscordInfo(ctx context.Context, userID collection.UserID, username, avatar string, lastUpdated time.Time) error {
	return t.q.UpdateDiscordInfo(ctx, storage.UpdateDiscordInfoParams{
		DiscordUsername: username,
		DiscordAvatar:   avatar,
		LastUpdated:     pgtype.Timestamp{Time: lastUpdated.UTC(), Valid: true},
		UserID:          uint64(userID),
	})
}

func toUser(u storage.User) collection.User {
	return collection.User{
		UserID:              collection.UserID(u.UserID),
		LastRoll:            u.Date.Time,
		Tokens:              u.Tokens,
		Quote:               u.Quote,
		FavoriteCharacterID: catalog.CharacterID(u.Favorite.Int64),
		AnilistURL:          u.AnilistUrl,
		DiscordUsername:     u.DiscordUsername,
		DiscordAvatar:       u.DiscordAvatar,
		LastUpdated:         u.LastUpdated.Time,
	}
}
