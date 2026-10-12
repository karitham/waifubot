package userpg

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage/userstore"
)

type Pg struct {
	Q userstore.Querier
}

func New(q userstore.Querier) *Pg {
	return &Pg{Q: q}
}

func (p *Pg) GetUser(ctx context.Context, userID collection.UserID) (collection.User, error) {
	u, err := p.Q.Get(ctx, uint64(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return collection.User{}, collection.ErrNotFound
		}
		return collection.User{}, err
	}
	return toUser(u), nil
}

func (p *Pg) CreateUser(ctx context.Context, userID collection.UserID) error {
	return p.Q.Create(ctx, uint64(userID))
}

func (p *Pg) GetUserByAnilist(ctx context.Context, anilistURL string) (collection.User, error) {
	u, err := p.Q.GetByAnilist(ctx, anilistURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return collection.User{}, collection.ErrNotFound
		}
		return collection.User{}, err
	}
	return toUser(u), nil
}

func (p *Pg) GetUserByDiscordUsername(ctx context.Context, username string) (collection.User, error) {
	u, err := p.Q.GetByDiscordUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return collection.User{}, collection.ErrNotFound
		}
		return collection.User{}, err
	}
	return toUser(u), nil
}

func (p *Pg) RecordRoll(ctx context.Context, userID collection.UserID, now time.Time, cooldown time.Duration) error {
	rows, err := p.Q.RecordRoll(ctx, userstore.RecordRollParams{
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
	u, err := p.Q.Get(ctx, uint64(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return collection.ErrNotFound
		}
		return err
	}
	return collection.ErrRollCooldown{Until: u.Date.Time.Add(cooldown)}
}

func (p *Pg) SpendTokens(ctx context.Context, userID collection.UserID, amount int32) (collection.User, error) {
	u, err := p.Q.SpendTokens(ctx, userstore.SpendTokensParams{
		Tokens: amount,
		UserID: uint64(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return collection.User{}, collection.ErrInsufficientTokens
		}
		return collection.User{}, err
	}
	return toUser(u), nil
}

func (p *Pg) AddTokens(ctx context.Context, userID collection.UserID, amount int32) (collection.User, error) {
	u, err := p.Q.UpdateTokens(ctx, userstore.UpdateTokensParams{
		Tokens: amount,
		UserID: uint64(userID),
	})
	if err != nil {
		return collection.User{}, err
	}
	return toUser(u), nil
}

func (p *Pg) UpdateFavorite(ctx context.Context, userID collection.UserID, charID catalog.CharacterID) error {
	return p.Q.UpdateFavorite(ctx, userstore.UpdateFavoriteParams{
		Favorite: pgtype.Int8{Int64: int64(charID), Valid: true},
		UserID:   uint64(userID),
	})
}

func (p *Pg) UpdateQuote(ctx context.Context, userID collection.UserID, quote string) error {
	return p.Q.UpdateQuote(ctx, userstore.UpdateQuoteParams{
		Quote:  quote,
		UserID: uint64(userID),
	})
}

func (p *Pg) UpdateAnilistURL(ctx context.Context, userID collection.UserID, url string) error {
	return p.Q.UpdateAnilistURL(ctx, userstore.UpdateAnilistURLParams{
		AnilistUrl: url,
		UserID:     uint64(userID),
	})
}

func (p *Pg) UpdateDiscordInfo(ctx context.Context, userID collection.UserID, username, avatar string, lastUpdated time.Time) error {
	return p.Q.UpdateDiscordInfo(ctx, userstore.UpdateDiscordInfoParams{
		DiscordUsername: username,
		DiscordAvatar:   avatar,
		LastUpdated:     pgtype.Timestamp{Time: lastUpdated.UTC(), Valid: true},
		UserID:          uint64(userID),
	})
}

func toUser(u userstore.User) collection.User {
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
