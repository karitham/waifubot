package main

import (
	"context"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/discord"
)

// FakeTrackingService is a fake implementation of discord.TrackingService
// for deterministic testing.
type FakeTrackingService struct {
	CharID     int64
	CharName   string
	CharImage  string
	MediaTitle string
}

// RandomChar returns the configured test character.
func (f *FakeTrackingService) RandomChar(ctx context.Context, notIn ...int64) (collection.MediaCharacter, error) {
	return collection.MediaCharacter{
		ID:         f.CharID,
		Name:       f.CharName,
		ImageURL:   f.CharImage,
		MediaTitle: f.MediaTitle,
	}, nil
}

// Anime returns empty results for testing.
func (f *FakeTrackingService) Anime(ctx context.Context, name string) ([]collection.Media, error) {
	return nil, nil
}

// Manga returns empty results for testing.
func (f *FakeTrackingService) Manga(ctx context.Context, name string) ([]collection.Media, error) {
	return nil, nil
}

// User returns empty results for testing.
func (f *FakeTrackingService) User(ctx context.Context, name string) ([]collection.TrackerUser, error) {
	return nil, nil
}

// Character returns empty results for testing.
func (f *FakeTrackingService) Character(ctx context.Context, name string) ([]collection.MediaCharacter, error) {
	return nil, nil
}

// SearchMedia returns empty results for testing.
func (f *FakeTrackingService) SearchMedia(ctx context.Context, search string) ([]collection.Media, error) {
	return nil, nil
}

// GetMediaCharacters returns empty results for testing.
func (f *FakeTrackingService) GetMediaCharacters(ctx context.Context, mediaId int64) ([]collection.MediaCharacter, error) {
	return nil, nil
}

// Ensure FakeTrackingService implements discord.TrackingService.
var _ discord.TrackingService = (*FakeTrackingService)(nil)
