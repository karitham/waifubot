package discord

import (
	"context"
	"fmt"
	"time"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
)

// Service refreshes stored Discord user information.
type Service struct {
	client *Client
}

func NewService(botToken string) *Service {
	return &Service{
		client: NewClient(botToken),
	}
}

func (s *Service) UpdateUserInfo(ctx context.Context, store collection.DB, userID corde.Snowflake) error {
	discordUser, err := s.client.GetUser(ctx, fmt.Sprintf("%d", uint64(userID)))
	if err != nil {
		return fmt.Errorf("failed to fetch Discord user: %w", err)
	}

	return collection.SetDiscordInfo(ctx, store, collection.UserID(userID), discordUser.Username, discordUser.Avatar, time.Now())
}

func (s *Service) UpdateIfNeeded(ctx context.Context, store collection.DB, userID corde.Snowflake) error {
	shouldUpdate, err := shouldUpdateInfo(ctx, store, userID)
	if err != nil {
		return err
	}

	if shouldUpdate {
		return s.UpdateUserInfo(ctx, store, userID)
	}

	return nil
}

// shouldUpdateInfo reports whether Discord info is older than 24 hours.
func shouldUpdateInfo(ctx context.Context, store collection.Store, userID corde.Snowflake) (bool, error) {
	u, err := store.GetUser(ctx, collection.UserID(userID))
	if err != nil {
		return false, err
	}

	return u.LastUpdated.IsZero() || time.Since(u.LastUpdated) > 24*time.Hour, nil
}
