package sync_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/collectiontest"
	"github.com/karitham/waifubot/sync"
)

type mockFetcher struct {
	CharactersByIDsFunc func(ctx context.Context, ids []int64) ([]collection.MediaCharacter, error)
}

func (m *mockFetcher) CharactersByIDs(ctx context.Context, ids []int64) ([]collection.MediaCharacter, error) {
	if m.CharactersByIDsFunc != nil {
		return m.CharactersByIDsFunc(ctx, ids)
	}
	return nil, nil
}

func TestProcessBatch(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name              string
		characters        []catalog.Character
		fetcherResult     []collection.MediaCharacter
		fetcherErr        error
		updateErr         error
		wantCursorUpdated time.Time
		wantCursorID      int64
	}{
		{
			name: "happy path - all characters updated",
			characters: []catalog.Character{
				{ID: 1, Name: "A", UpdatedAt: now.Add(-2 * time.Hour)},
				{ID: 2, Name: "B", UpdatedAt: now.Add(-time.Hour)},
			},
			fetcherResult: []collection.MediaCharacter{
				{ID: 1, Name: "A Updated", Favorites: 100},
				{ID: 2, Name: "B Updated", Favorites: 200},
			},
			wantCursorUpdated: now.Add(-time.Hour),
			wantCursorID:      2,
		},
		{
			name: "character not found in anilist - cursor still advances",
			characters: []catalog.Character{
				{ID: 1, Name: "A", UpdatedAt: now.Add(-2 * time.Hour)},
				{ID: 2, Name: "Missing", UpdatedAt: now.Add(-time.Hour)},
			},
			fetcherResult: []collection.MediaCharacter{
				{ID: 1, Name: "A Updated", Favorites: 100},
			},
			wantCursorUpdated: now.Add(-time.Hour),
			wantCursorID:      2,
		},
		{
			name: "anilist fetch fails - cursor still advances",
			characters: []catalog.Character{
				{ID: 1, Name: "A", UpdatedAt: now.Add(-2 * time.Hour)},
				{ID: 2, Name: "B", UpdatedAt: now.Add(-time.Hour)},
			},
			fetcherErr:        errors.New("network error"),
			wantCursorUpdated: now.Add(-time.Hour),
			wantCursorID:      2,
		},
		{
			name: "update fails - cursor still advances",
			characters: []catalog.Character{
				{ID: 1, Name: "A", UpdatedAt: now.Add(-2 * time.Hour)},
				{ID: 2, Name: "B", UpdatedAt: now.Add(-time.Hour)},
			},
			fetcherResult: []collection.MediaCharacter{
				{ID: 1, Name: "A Updated", Favorites: 100},
				{ID: 2, Name: "B Updated", Favorites: 200},
			},
			updateErr:         errors.New("db error"),
			wantCursorUpdated: now.Add(-time.Hour),
			wantCursorID:      2,
		},
		{
			name: "single character",
			characters: []catalog.Character{
				{ID: 42, Name: "Solo", UpdatedAt: now.Add(-3 * time.Hour)},
			},
			fetcherResult: []collection.MediaCharacter{
				{ID: 42, Name: "Solo Updated", Favorites: 500},
			},
			wantCursorUpdated: now.Add(-3 * time.Hour),
			wantCursorID:      42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &collectiontest.MockStore{
				UpdateCharacterSyncFunc: func(_ context.Context, char catalog.Character) (catalog.Character, error) {
					if tt.updateErr != nil {
						return catalog.Character{}, tt.updateErr
					}
					return char, nil
				},
			}

			fetcher := &mockFetcher{
				CharactersByIDsFunc: func(_ context.Context, _ []int64) ([]collection.MediaCharacter, error) {
					return tt.fetcherResult, tt.fetcherErr
				},
			}

			svc := &sync.Service{
				Store:   store,
				Anilist: fetcher,
			}

			cursorUpdated, cursorID := svc.ProcessBatch(t.Context(), tt.characters)
			assert.Equal(t, tt.wantCursorUpdated, cursorUpdated)
			assert.Equal(t, tt.wantCursorID, cursorID)
		})
	}
}
