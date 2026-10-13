package collectiontest

import (
	"context"

	"github.com/karitham/waifubot/tracker"
)

type MockTrackerService struct {
	GetMediaCharactersFunc func(ctx context.Context, mediaID tracker.MediaID) ([]tracker.Character, error)
	AnimeFunc              func(ctx context.Context, name string) ([]tracker.Media, error)
	MangaFunc              func(ctx context.Context, name string) ([]tracker.Media, error)
	UserFunc               func(ctx context.Context, name string) ([]tracker.User, error)
	CharacterFunc          func(ctx context.Context, name string) ([]tracker.Character, error)
	SearchMediaFunc        func(ctx context.Context, search string) ([]tracker.Media, error)
}

func (m *MockTrackerService) GetMediaCharacters(ctx context.Context, mediaID tracker.MediaID) ([]tracker.Character, error) {
	if m.GetMediaCharactersFunc != nil {
		return m.GetMediaCharactersFunc(ctx, mediaID)
	}
	return nil, nil
}

func (m *MockTrackerService) Anime(ctx context.Context, name string) ([]tracker.Media, error) {
	if m.AnimeFunc != nil {
		return m.AnimeFunc(ctx, name)
	}
	return nil, nil
}

func (m *MockTrackerService) Manga(ctx context.Context, name string) ([]tracker.Media, error) {
	if m.MangaFunc != nil {
		return m.MangaFunc(ctx, name)
	}
	return nil, nil
}

func (m *MockTrackerService) User(ctx context.Context, name string) ([]tracker.User, error) {
	if m.UserFunc != nil {
		return m.UserFunc(ctx, name)
	}
	return nil, nil
}

func (m *MockTrackerService) Character(ctx context.Context, name string) ([]tracker.Character, error) {
	if m.CharacterFunc != nil {
		return m.CharacterFunc(ctx, name)
	}
	return nil, nil
}

func (m *MockTrackerService) SearchMedia(ctx context.Context, search string) ([]tracker.Media, error) {
	if m.SearchMediaFunc != nil {
		return m.SearchMediaFunc(ctx, search)
	}
	return nil, nil
}

var _ tracker.Service = (*MockTrackerService)(nil)
