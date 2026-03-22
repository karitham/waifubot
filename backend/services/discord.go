package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/discord"
)

// DiscordService handles Discord user information updates.
type DiscordService struct {
	discordClient *discord.Client
}

// NewDiscordService creates a new Discord service.
func NewDiscordService(botToken string) *DiscordService {
	return &DiscordService{
		discordClient: discord.NewClient(botToken),
	}
}

// UpdateUserInfo fetches and updates Discord user information.
func (ds *DiscordService) UpdateUserInfo(ctx context.Context, store collection.Store, userID corde.Snowflake) error {
	discordUser, err := ds.discordClient.GetUser(ctx, fmt.Sprintf("%d", uint64(userID)))
	if err != nil {
		return fmt.Errorf("failed to fetch Discord user: %w", err)
	}

	return store.UpdateDiscordInfo(ctx, uint64(userID), discordUser.Username, discordUser.Avatar, time.Now())
}

// ShouldUpdateInfo checks if Discord info should be updated (older than 24 hours).
func ShouldUpdateInfo(ctx context.Context, store collection.Store, userID corde.Snowflake) (bool, error) {
	u, err := store.GetUser(ctx, uint64(userID))
	if err != nil {
		return false, err
	}

	// If never updated (zero time), or updated more than 24 hours ago
	if u.LastUpdated.IsZero() || time.Since(u.LastUpdated) > 24*time.Hour {
		return true, nil
	}

	return false, nil
}

// UpdateIfNeeded updates Discord user info if it's stale.
func (ds *DiscordService) UpdateIfNeeded(ctx context.Context, store collection.Store, userID corde.Snowflake) error {
	shouldUpdate, err := ShouldUpdateInfo(ctx, store, userID)
	if err != nil {
		return err
	}

	if shouldUpdate {
		return ds.UpdateUserInfo(ctx, store, userID)
	}

	return nil
}
