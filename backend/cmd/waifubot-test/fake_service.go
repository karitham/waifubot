package main

import (
	"context"

	"github.com/karitham/waifubot/discord"
	"github.com/karitham/waifubot/tracker"
)

// FakeTrackingService is a fake implementation of discord.TrackingService
// for deterministic testing.
type FakeTrackingService struct {
	CharID     int64
	CharName   string
	CharImage  string
	MediaTitle string
}

// Anime returns empty results for testing.
func (f *FakeTrackingService) Anime(ctx context.Context, name string) ([]tracker.Media, error) {
	return nil, nil
}

// Manga returns empty results for testing.
func (f *FakeTrackingService) Manga(ctx context.Context, name string) ([]tracker.Media, error) {
	return nil, nil
}

// User returns empty results for testing.
func (f *FakeTrackingService) User(ctx context.Context, name string) ([]tracker.User, error) {
	return nil, nil
}

// Character returns empty results for testing.
func (f *FakeTrackingService) Character(ctx context.Context, name string) ([]tracker.Character, error) {
	return nil, nil
}

// SearchMedia returns empty results for testing.
func (f *FakeTrackingService) SearchMedia(ctx context.Context, search string) ([]tracker.Media, error) {
	return nil, nil
}

// GetMediaCharacters returns empty results for testing.
func (f *FakeTrackingService) GetMediaCharacters(ctx context.Context, mediaId int64) ([]tracker.Character, error) {
	return nil, nil
}

// Ensure FakeTrackingService implements discord.TrackingService.
var _ discord.TrackingService = (*FakeTrackingService)(nil)
