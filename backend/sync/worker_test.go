package sync

import (
	"context"
	"errors"
	"testing"
	"time"

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

func TestBuildDiscoveryPool(t *testing.T) {
	t.Run("deduplicates overlapping active IDs and discovery window", func(t *testing.T) {
		store := &collectiontest.MockStore{
			GetActiveIDsFunc: func(_ context.Context) ([]int64, error) {
				return []int64{1, 50, 100, 150, 200}, nil
			},
		}

		svc := &Service{Store: store}
		svc.maxID.Store(200) // maxID < discoveryWindow ⇒ window covers [1, 200]

		pool := svc.buildDiscoveryPool(t.Context())
		require.NotNil(t, pool)

		// No duplicates: pool length must equal number of unique IDs.
		seen := make(map[int64]struct{}, len(pool))
		for _, id := range pool {
			seen[id] = struct{}{}
		}
		assert.Equal(t, len(pool), len(seen), "pool must not contain duplicate IDs")

		// All active IDs must be present in the pool.
		for _, id := range []int64{1, 50, 100, 150, 200} {
			_, ok := seen[id]
			assert.True(t, ok, "active ID %d should be in pool", id)
		}
	})

	t.Run("returns nil on store error", func(t *testing.T) {
		store := &collectiontest.MockStore{
			GetActiveIDsFunc: func(_ context.Context) ([]int64, error) {
				return nil, errors.New("store error")
			},
		}

		svc := &Service{Store: store}
		svc.maxID.Store(200)

		pool := svc.buildDiscoveryPool(t.Context())
		assert.Nil(t, pool)
	})
}

func TestRun(t *testing.T) {
	var upserted int

	store := &collectiontest.MockStore{
		GetActiveIDsFunc: func(_ context.Context) ([]int64, error) {
			return []int64{1, 50, 100, 150, 200}, nil
		},
		UpsertCharacterFunc: func(_ context.Context, _ catalog.Character) error {
			upserted++
			return nil
		},
		MarkCharactersInactiveFunc: func(_ context.Context, _ []int64) error {
			return nil
		},
	}

	fetcher := &mockFetcher{
		CharactersByIDsFunc: func(_ context.Context, ids []int64) ([]collection.MediaCharacter, error) {
			chars := make([]collection.MediaCharacter, len(ids))
			for i, id := range ids {
				chars[i] = collection.MediaCharacter{ID: id, Name: "Char"}
			}
			return chars, nil
		},
	}

	svc := &Service{Store: store, Anilist: fetcher}
	svc.maxID.Store(200)

	ctx, cancel := context.WithCancel(t.Context())
	time.AfterFunc(100*time.Millisecond, cancel)

	svc.Run(ctx)

	assert.Greater(t, upserted, 0, "expected at least one upserted character")
}
