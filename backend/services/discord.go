package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Karitham/corde"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/karitham/waifubot/discord"
	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/userstore"
)

// DiscordService handles Discord user information updates
type DiscordService struct {
	discordClient *discord.Client
}

// NewDiscordService creates a new Discord service
func NewDiscordService(botToken string) *DiscordService {
	return &DiscordService{
		discordClient: discord.NewClient(botToken),
	}
}

// UpdateUserInfo fetches and updates Discord user information
func (ds *DiscordService) UpdateUserInfo(ctx context.Context, store storage.Store, userID corde.Snowflake) error {
	// Fetch user from Discord API
	discordUser, err := ds.discordClient.GetUser(ctx, fmt.Sprintf("%d", uint64(userID)))
	if err != nil {
		return fmt.Errorf("failed to fetch Discord user: %w", err)
	}

	// Update user in database
	return store.UserStore().UpdateDiscordInfo(ctx, userstore.UpdateDiscordInfoParams{
		DiscordUsername: discordUser.Username,
		DiscordAvatar:   discordUser.Avatar,
		LastUpdated:     pgtype.Timestamp{Time: time.Now(), Valid: true},
		UserID:          uint64(userID),
	})
}

// ShouldUpdateInfo checks if Discord info should be updated (older than 24 hours)
func ShouldUpdateInfo(ctx context.Context, store storage.Store, userID corde.Snowflake) (bool, error) {
	u, err := store.UserStore().Get(ctx, uint64(userID))
	if err != nil {
		return false, err
	}

	// If never updated, or updated more than 24 hours ago
	if !u.LastUpdated.Valid || time.Since(u.LastUpdated.Time) > 24*time.Hour {
		return true, nil
	}

	return false, nil
}

// UpdateIfNeeded updates Discord user info if it's stale
func (ds *DiscordService) UpdateIfNeeded(ctx context.Context, store storage.Store, userID corde.Snowflake) error {
	shouldUpdate, err := ShouldUpdateInfo(ctx, store, userID)
	if err != nil {
		return err
	}

	if shouldUpdate {
		return ds.UpdateUserInfo(ctx, store, userID)
	}

	return nil
}
