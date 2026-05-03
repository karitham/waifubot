package sync

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/collectiontest"
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

func TestSync(t *testing.T) {
	tests := []struct {
		name          string
		fetchErr      error
		numReturned   int // number of characters to return from the requested batch
		wantErr       bool
		wantUpserted  int
		wantMarkCount int // expected number of IDs marked inactive
	}{
		{
			name:          "all found - nothing to mark",
			numReturned:   50,
			wantUpserted:  50,
			wantMarkCount: 0,
		},
		{
			name:          "some missing - marked inactive",
			numReturned:   1,
			wantUpserted:  1,
			wantMarkCount: 49,
		},
		{
			name:          "all missing - all marked inactive",
			numReturned:   0,
			wantUpserted:  0,
			wantMarkCount: 50,
		},
		{
			name:     "fetch error",
			fetchErr: errors.New("api error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var upserted int
			var markedInactive []int64

			store := &collectiontest.MockStore{
				UpsertCharacterFunc: func(_ context.Context, _ catalog.Character) error {
					upserted++
					return nil
				},
				MarkCharactersInactiveFunc: func(_ context.Context, ids []int64) error {
					markedInactive = append(markedInactive, ids...)
					return nil
				},
			}

			fetcher := &mockFetcher{
				CharactersByIDsFunc: func(_ context.Context, ids []int64) ([]collection.MediaCharacter, error) {
					if tt.fetchErr != nil {
						return nil, tt.fetchErr
					}
					// Return the first numReturned IDs from the batch so they
					// are guaranteed to intersect with requested IDs.
					chars := make([]collection.MediaCharacter, tt.numReturned)
					for i := range chars {
						chars[i] = collection.MediaCharacter{ID: ids[i], Name: "Char"}
					}
					return chars, nil
				},
			}

			svc := &Service{Store: store, Anilist: fetcher}
			svc.maxID.Store(100)
			err := svc.Sync(t.Context())

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantUpserted, upserted)
			assert.Len(t, markedInactive, tt.wantMarkCount)
		})
	}
}
