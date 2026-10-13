package main

import (
	"context"

	"github.com/karitham/waifubot/tracker"
)

// FakeTrackerService is a fake implementation of tracker.Service
// for deterministic testing.
type FakeTrackerService struct {
	CharID     int64
	CharName   string
	CharImage  string
	MediaTitle string
}

// Anime returns empty results for testing.
func (f *FakeTrackerService) Anime(ctx context.Context, name string) ([]tracker.Media, error) {
	return nil, nil
}

// Manga returns empty results for testing.
func (f *FakeTrackerService) Manga(ctx context.Context, name string) ([]tracker.Media, error) {
	return nil, nil
}

// User returns empty results for testing.
func (f *FakeTrackerService) User(ctx context.Context, name string) ([]tracker.User, error) {
	return nil, nil
}

// Character returns empty results for testing.
func (f *FakeTrackerService) Character(ctx context.Context, name string) ([]tracker.Character, error) {
	return nil, nil
}

// SearchMedia returns empty results for testing.
func (f *FakeTrackerService) SearchMedia(ctx context.Context, search string) ([]tracker.Media, error) {
	return nil, nil
}

// GetMediaCharacters returns empty results for testing.
func (f *FakeTrackerService) GetMediaCharacters(ctx context.Context, mediaID tracker.MediaID) ([]tracker.Character, error) {
	return nil, nil
}

// Ensure FakeTrackerService implements tracker.Service.
var _ tracker.Service = (*FakeTrackerService)(nil)
