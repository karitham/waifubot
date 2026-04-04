package discord

import (
	"context"
	"errors"
	"testing"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/collectiontest"
	"github.com/karitham/waifubot/discord/cordetest"
	"github.com/stretchr/testify/assert"
)

func TestSearchHandler_SearchAnime(t *testing.T) {
	tests := []struct {
		name        string
		cmd         CommandContext
		animeFunc   func(ctx context.Context, name string) ([]collection.Media, error)
		wantContent string
		wantTitle   string
	}{
		{
			name: "no name provided",
			cmd: &MockCommandContext{
				OptStringVals: map[string]string{},
			},
			wantContent: "No anime found with that name",
		},
		{
			name: "no results",
			cmd: &MockCommandContext{
				OptStringVals: map[string]string{"name": "nonexistent"},
			},
			animeFunc: func(ctx context.Context, name string) ([]collection.Media, error) {
				return nil, nil
			},
			wantContent: "No anime found with that name",
		},
		{
			name: "service error",
			cmd: &MockCommandContext{
				OptStringVals: map[string]string{"name": "naruto"},
			},
			animeFunc: func(ctx context.Context, name string) ([]collection.Media, error) {
				return nil, errors.New("service error")
			},
			wantContent: "Error searching for this anime",
		},
		{
			name: "success",
			cmd: &MockCommandContext{
				OptStringVals: map[string]string{"name": "naruto"},
			},
			animeFunc: func(ctx context.Context, name string) ([]collection.Media, error) {
				return []collection.Media{
					{ID: 1, Title: "Naruto", URL: "https://anilist.co/anime/20", CoverImageURL: "https://example.com/cover.png"},
				}, nil
			},
			wantTitle: "Naruto",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &SearchHandler{
				animeService: &collectiontest.MockAnimeService{AnimeFunc: tt.animeFunc},
			}

			h.SearchAnime(t.Context(), w, tt.cmd)

			assert.True(t, w.RespondCalled)
			if tt.wantContent != "" {
				data := w.LastRespond.InteractionRespData()
				assert.Contains(t, data.Content, tt.wantContent)
			}
			if tt.wantTitle != "" {
				data := w.LastRespond.InteractionRespData()
				if assert.Len(t, data.Embeds, 1) {
					assert.Equal(t, tt.wantTitle, data.Embeds[0].Title)
				}
			}
		})
	}
}

func TestSearchHandler_SearchManga(t *testing.T) {
	tests := []struct {
		name        string
		cmd         CommandContext
		mangaFunc   func(ctx context.Context, name string) ([]collection.Media, error)
		wantContent string
		wantTitle   string
	}{
		{
			name: "no name provided",
			cmd: &MockCommandContext{
				OptStringVals: map[string]string{},
			},
			wantContent: "No manga found with that name",
		},
		{
			name: "no results",
			cmd: &MockCommandContext{
				OptStringVals: map[string]string{"name": "nonexistent"},
			},
			mangaFunc: func(ctx context.Context, name string) ([]collection.Media, error) {
				return nil, nil
			},
			wantContent: "No manga found with that name",
		},
		{
			name: "service error",
			cmd: &MockCommandContext{
				OptStringVals: map[string]string{"name": "one piece"},
			},
			mangaFunc: func(ctx context.Context, name string) ([]collection.Media, error) {
				return nil, errors.New("service error")
			},
			wantContent: "Error searching for this manga",
		},
		{
			name: "success",
			cmd: &MockCommandContext{
				OptStringVals: map[string]string{"name": "one piece"},
			},
			mangaFunc: func(ctx context.Context, name string) ([]collection.Media, error) {
				return []collection.Media{
					{ID: 2, Title: "One Piece", URL: "https://anilist.co/manga/30", CoverImageURL: "https://example.com/cover.png"},
				}, nil
			},
			wantTitle: "One Piece",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &SearchHandler{
				animeService: &collectiontest.MockAnimeService{MangaFunc: tt.mangaFunc},
			}

			h.SearchManga(t.Context(), w, tt.cmd)

			assert.True(t, w.RespondCalled)
			if tt.wantContent != "" {
				data := w.LastRespond.InteractionRespData()
				assert.Contains(t, data.Content, tt.wantContent)
			}
			if tt.wantTitle != "" {
				data := w.LastRespond.InteractionRespData()
				if assert.Len(t, data.Embeds, 1) {
					assert.Equal(t, tt.wantTitle, data.Embeds[0].Title)
				}
			}
		})
	}
}

func TestSearchHandler_SearchUser(t *testing.T) {
	tests := []struct {
		name        string
		cmd         CommandContext
		userFunc    func(ctx context.Context, name string) ([]collection.TrackerUser, error)
		wantContent string
		wantTitle   string
	}{
		{
			name: "no name provided",
			cmd: &MockCommandContext{
				OptStringVals: map[string]string{},
			},
			wantContent: "No user found with that name",
		},
		{
			name: "no results",
			cmd: &MockCommandContext{
				OptStringVals: map[string]string{"name": "nonexistent"},
			},
			userFunc: func(ctx context.Context, name string) ([]collection.TrackerUser, error) {
				return nil, nil
			},
			wantContent: "No user found with that name",
		},
		{
			name: "service error",
			cmd: &MockCommandContext{
				OptStringVals: map[string]string{"name": "testuser"},
			},
			userFunc: func(ctx context.Context, name string) ([]collection.TrackerUser, error) {
				return nil, errors.New("service error")
			},
			wantContent: "Error searching for this user",
		},
		{
			name: "success",
			cmd: &MockCommandContext{
				OptStringVals: map[string]string{"name": "testuser"},
			},
			userFunc: func(ctx context.Context, name string) ([]collection.TrackerUser, error) {
				return []collection.TrackerUser{
					{Name: "TestUser", URL: "https://anilist.co/user/testuser", About: "Hello"},
				}, nil
			},
			wantTitle: "TestUser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &SearchHandler{
				animeService: &collectiontest.MockAnimeService{UserFunc: tt.userFunc},
			}

			h.SearchUser(t.Context(), w, tt.cmd)

			assert.True(t, w.RespondCalled)
			if tt.wantContent != "" {
				data := w.LastRespond.InteractionRespData()
				assert.Contains(t, data.Content, tt.wantContent)
			}
			if tt.wantTitle != "" {
				data := w.LastRespond.InteractionRespData()
				if assert.Len(t, data.Embeds, 1) {
					assert.Equal(t, tt.wantTitle, data.Embeds[0].Title)
				}
			}
		})
	}
}

func TestSearchHandler_SearchChar(t *testing.T) {
	tests := []struct {
		name        string
		cmd         CommandContext
		charFunc    func(ctx context.Context, name string) ([]collection.MediaCharacter, error)
		wantContent string
		wantTitle   string
	}{
		{
			name: "no name provided",
			cmd: &MockCommandContext{
				OptStringVals: map[string]string{},
			},
			wantContent: "No character found with that name",
		},
		{
			name: "no results",
			cmd: &MockCommandContext{
				OptStringVals: map[string]string{"name": "nonexistent"},
			},
			charFunc: func(ctx context.Context, name string) ([]collection.MediaCharacter, error) {
				return nil, nil
			},
			wantContent: "No character found with that name",
		},
		{
			name: "service error",
			cmd: &MockCommandContext{
				OptStringVals: map[string]string{"name": "sakura"},
			},
			charFunc: func(ctx context.Context, name string) ([]collection.MediaCharacter, error) {
				return nil, errors.New("service error")
			},
			wantContent: "Error searching for this character",
		},
		{
			name: "success",
			cmd: &MockCommandContext{
				OptStringVals: map[string]string{"name": "sakura"},
			},
			charFunc: func(ctx context.Context, name string) ([]collection.MediaCharacter, error) {
				return []collection.MediaCharacter{
					{ID: 10, Name: "Sakura", URL: "https://anilist.co/character/10", ImageURL: "https://example.com/sakura.png"},
				}, nil
			},
			wantTitle: "Sakura",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &SearchHandler{
				animeService: &collectiontest.MockAnimeService{CharacterFunc: tt.charFunc},
			}

			h.SearchChar(t.Context(), w, tt.cmd)

			assert.True(t, w.RespondCalled)
			if tt.wantContent != "" {
				data := w.LastRespond.InteractionRespData()
				assert.Contains(t, data.Content, tt.wantContent)
			}
			if tt.wantTitle != "" {
				data := w.LastRespond.InteractionRespData()
				if assert.Len(t, data.Embeds, 1) {
					assert.Equal(t, tt.wantTitle, data.Embeds[0].Title)
				}
			}
		})
	}
}
