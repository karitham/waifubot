package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/karitham/waifubot/collection"
)

// UserStore abstracts user lookup and creation for the auth flow.
type UserStore interface {
	GetOrCreateUser(ctx context.Context, discordID uint64, username, avatar string) (uint64, error)
}

// UserStoreAdapter implements UserStore using collection.UserRepository.
type UserStoreAdapter struct {
	users collection.UserRepository
}

// NewUserStoreAdapter wraps a collection.UserRepository into a UserStore.
func NewUserStoreAdapter(users collection.UserRepository) *UserStoreAdapter {
	return &UserStoreAdapter{users: users}
}

func (a *UserStoreAdapter) GetOrCreateUser(ctx context.Context, discordID uint64, username, avatar string) (uint64, error) {
	u, err := a.users.GetUser(ctx, discordID)
	if err == nil {
		if err := a.users.UpdateDiscordInfo(ctx, discordID, username, avatar, time.Now()); err != nil {
			slog.With("err", err, "discord_id", discordID).Warn("failed to update Discord info")
		}
		return u.UserID, nil
	}

	// Only treat ErrNotFound as "user doesn't exist" - propagate other errors
	if !errors.Is(err, collection.ErrNotFound) {
		return 0, fmt.Errorf("get user: %w", err)
	}

	if err := a.users.CreateUser(ctx, discordID); err != nil {
		// Handle duplicate key error - user was created by concurrent request
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// Fetch the existing user to get its ID
			existingUser, fetchErr := a.users.GetUser(ctx, discordID)
			if fetchErr != nil {
				return 0, fmt.Errorf("get user after duplicate key: %w", fetchErr)
			}
			if err := a.users.UpdateDiscordInfo(ctx, discordID, username, avatar, time.Now()); err != nil {
				slog.With("err", err, "discord_id", discordID).Warn("failed to update Discord info")
			}
			return existingUser.UserID, nil
		}
		return 0, fmt.Errorf("create user: %w", err)
	}

	if err := a.users.UpdateDiscordInfo(ctx, discordID, username, avatar, time.Now()); err != nil {
		slog.With("err", err, "discord_id", discordID).Warn("failed to update Discord info")
	}

	// The user ID is the same as discordID - no need to fetch after create.
	return discordID, nil
}
