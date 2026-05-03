package collectiontest

import (
	"context"

	"github.com/karitham/waifubot/collection"
)

type MockAnimeService struct {
	GetMediaCharactersFunc func(ctx context.Context, mediaId int64) ([]collection.MediaCharacter, error)
	AnimeFunc              func(ctx context.Context, name string) ([]collection.Media, error)
	MangaFunc              func(ctx context.Context, name string) ([]collection.Media, error)
	UserFunc               func(ctx context.Context, name string) ([]collection.TrackerUser, error)
	CharacterFunc          func(ctx context.Context, name string) ([]collection.MediaCharacter, error)
	SearchMediaFunc        func(ctx context.Context, search string) ([]collection.Media, error)
}

func (m *MockAnimeService) GetMediaCharacters(ctx context.Context, mediaId int64) ([]collection.MediaCharacter, error) {
	if m.GetMediaCharactersFunc != nil {
		return m.GetMediaCharactersFunc(ctx, mediaId)
	}
	return nil, nil
}

func (m *MockAnimeService) Anime(ctx context.Context, name string) ([]collection.Media, error) {
	if m.AnimeFunc != nil {
		return m.AnimeFunc(ctx, name)
	}
	return nil, nil
}

func (m *MockAnimeService) Manga(ctx context.Context, name string) ([]collection.Media, error) {
	if m.MangaFunc != nil {
		return m.MangaFunc(ctx, name)
	}
	return nil, nil
}

func (m *MockAnimeService) User(ctx context.Context, name string) ([]collection.TrackerUser, error) {
	if m.UserFunc != nil {
		return m.UserFunc(ctx, name)
	}
	return nil, nil
}

func (m *MockAnimeService) Character(ctx context.Context, name string) ([]collection.MediaCharacter, error) {
	if m.CharacterFunc != nil {
		return m.CharacterFunc(ctx, name)
	}
	return nil, nil
}

func (m *MockAnimeService) SearchMedia(ctx context.Context, search string) ([]collection.Media, error) {
	if m.SearchMediaFunc != nil {
		return m.SearchMediaFunc(ctx, search)
	}
	return nil, nil
}

var _ collection.AnimeService = (*MockAnimeService)(nil)
