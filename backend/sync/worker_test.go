package sync_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/collectiontest"
	"github.com/karitham/waifubot/sync"
)

type mockFetcher struct {
	CharactersByIDsFunc func(ctx context.Context, ids []int64) ([]collection.MediaCharacter, error)
	BackfillPageFunc    func(ctx context.Context, page, perPage int) ([]collection.MediaCharacter, error)
}

func (m *mockFetcher) CharactersByIDs(ctx context.Context, ids []int64) ([]collection.MediaCharacter, error) {
	if m.CharactersByIDsFunc != nil {
		return m.CharactersByIDsFunc(ctx, ids)
	}
	return nil, nil
}

func (m *mockFetcher) BackfillPage(ctx context.Context, page, perPage int) ([]collection.MediaCharacter, error) {
	if m.BackfillPageFunc != nil {
		return m.BackfillPageFunc(ctx, page, perPage)
	}
	return nil, nil
}

func TestBackfill(t *testing.T) {
	tests := []struct {
		name      string
		pages     map[int][]collection.MediaCharacter // page -> characters
		perPage   int
		wantPages int // number of times BackfillPage was called
		wantChars int // total characters upserted
		wantErr   bool
	}{
		{
			name: "single page full",
			pages: map[int][]collection.MediaCharacter{
				1: {{ID: 1, Name: "A", Favorites: 100}, {ID: 2, Name: "B", Favorites: 50}},
			},
			perPage:   50,
			wantPages: 1,
			wantChars: 2,
		},
		{
			name: "multi page",
			pages: map[int][]collection.MediaCharacter{
				1: {{ID: 1, Name: "A"}},
				2: {{ID: 2, Name: "B"}},
				3: nil, // empty page — signals end
			},
			perPage:   1,
			wantPages: 3, // page 3 is fetched (returns empty) before breaking
			wantChars: 2,
		},
		{
			name: "partial last page",
			pages: map[int][]collection.MediaCharacter{
				1: {{ID: 1, Name: "A"}, {ID: 2, Name: "B"}},
				2: {{ID: 3, Name: "C"}}, // fewer than perPage = last page
			},
			perPage:   3,
			wantPages: 1, // stops after page 1 (2 < 3 = partial page)
			wantChars: 2,
		},
		{
			name: "fetch error stops backfill",
			pages: map[int][]collection.MediaCharacter{
				1: {{ID: 1, Name: "A"}},
			},
			perPage: 1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var callCount int
			var upsertedCount int

			store := &collectiontest.MockStore{
				UpsertCharacterFunc: func(_ context.Context, char catalog.Character) error {
					upsertedCount++
					return nil
				},
				GetActiveIDsFunc: func(_ context.Context) ([]int64, error) {
					return nil, nil // no active IDs to sweep
				},
			}

			fetcher := &mockFetcher{
				BackfillPageFunc: func(_ context.Context, page, perPage int) ([]collection.MediaCharacter, error) {
					callCount++
					chars, ok := tt.pages[page]
					if !ok {
						return nil, nil
					}
					if tt.wantErr && callCount >= 1 {
						return nil, errors.New("fetch error")
					}
					return chars, nil
				},
			}

			svc := &sync.Service{Store: store, Anilist: fetcher}
			err := svc.Backfill(t.Context(), tt.perPage)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if !tt.wantErr {
				assert.Equal(t, tt.wantPages, callCount)
				assert.Equal(t, tt.wantChars, upsertedCount)
			}
		})
	}
}

func TestDeletionSweep(t *testing.T) {
	tests := []struct {
		name         string
		activeIDs    []int64
		fetchedChars []collection.MediaCharacter
		fetchErr     error
		wantMarked   []int64 // IDs that should be marked inactive
	}{
		{
			name:         "no deletions",
			activeIDs:    []int64{1, 2, 3},
			fetchedChars: []collection.MediaCharacter{{ID: 1}, {ID: 2}, {ID: 3}},
			wantMarked:   nil,
		},
		{
			name:         "one deleted",
			activeIDs:    []int64{1, 2, 3},
			fetchedChars: []collection.MediaCharacter{{ID: 1}, {ID: 3}},
			wantMarked:   []int64{2},
		},
		{
			name:         "all deleted",
			activeIDs:    []int64{1, 2},
			fetchedChars: nil,
			wantMarked:   []int64{1, 2},
		},
		{
			name:       "fetch error skips batch",
			activeIDs:  []int64{1, 2},
			fetchErr:   errors.New("network error"),
			wantMarked: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var markedInactive []int64

			store := &collectiontest.MockStore{
				GetActiveIDsFunc: func(_ context.Context) ([]int64, error) {
					return tt.activeIDs, nil
				},
				MarkCharacterInactiveFunc: func(_ context.Context, charID int64) error {
					markedInactive = append(markedInactive, charID)
					return nil
				},
				UpsertCharacterFunc: func(_ context.Context, char catalog.Character) error {
					return nil
				},
			}

			fetcher := &mockFetcher{
				BackfillPageFunc: func(_ context.Context, page, perPage int) ([]collection.MediaCharacter, error) {
					// Return empty page so backfill finishes immediately
					return nil, nil
				},
				CharactersByIDsFunc: func(_ context.Context, ids []int64) ([]collection.MediaCharacter, error) {
					if tt.fetchErr != nil {
						return nil, tt.fetchErr
					}
					return tt.fetchedChars, nil
				},
			}

			svc := &sync.Service{Store: store, Anilist: fetcher}
			err := svc.Backfill(t.Context(), 50)
			require.NoError(t, err)

			assert.ElementsMatch(t, tt.wantMarked, markedInactive)
		})
	}
}
