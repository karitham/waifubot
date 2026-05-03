package sync

import (
	"context"
	"errors"
	"sync/atomic"
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

			svc := &Service{store: store, anilist: fetcher}
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
	tests := []struct {
		name      string
		activeIDs []int64
		maxID     int64
		storeErr  error
		wantNil   bool
		wantIDs   []int64 // IDs that must be in the pool
	}{
		{
			name:      "deduplicates overlapping active and discovery IDs",
			activeIDs: []int64{1, 50, 100, 150, 200},
			maxID:     200,
			wantNil:   false,
			wantIDs:   []int64{1, 50, 100, 150, 200},
		},
		{
			name:     "returns nil on store error",
			storeErr: errors.New("store error"),
			maxID:    200,
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &collectiontest.MockStore{
				GetActiveIDsFunc: func(_ context.Context) ([]int64, error) {
					if tt.storeErr != nil {
						return nil, tt.storeErr
					}
					return tt.activeIDs, nil
				},
			}

			svc := &Service{store: store}
			svc.maxID.Store(tt.maxID)

			pool := svc.buildDiscoveryPool(t.Context())

			if tt.wantNil {
				assert.Nil(t, pool)
				return
			}

			require.NotNil(t, pool)

			// No duplicates
			seen := make(map[int64]struct{}, len(pool))
			for _, id := range pool {
				seen[id] = struct{}{}
			}
			assert.Equal(t, len(pool), len(seen), "pool must not contain duplicate IDs")

			// All expected IDs present
			for _, id := range tt.wantIDs {
				assert.Contains(t, seen, id, "active ID %d should be in pool", id)
			}

			// All IDs are within [1, maxID]
			for _, id := range pool {
				assert.GreaterOrEqual(t, id, int64(1))
				assert.LessOrEqual(t, id, tt.maxID)
			}
		})
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name      string
		maxID     int64
		activeIDs []int64
		wantMin   int // minimum number of upserted characters
	}{
		{
			name:      "processes batches until cancelled",
			maxID:     200,
			activeIDs: []int64{1, 50, 100, 150, 200},
			wantMin:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var upserted atomic.Int64

			store := &collectiontest.MockStore{
				GetActiveIDsFunc: func(_ context.Context) ([]int64, error) {
					return tt.activeIDs, nil
				},
				UpsertCharacterFunc: func(_ context.Context, _ catalog.Character) error {
					upserted.Add(1)
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

			svc := &Service{store: store, anilist: fetcher}
			svc.maxID.Store(tt.maxID)

			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			t.Cleanup(cancel)
			time.AfterFunc(100*time.Millisecond, cancel)

			svc.Run(ctx)

			assert.GreaterOrEqual(t, int(upserted.Load()), tt.wantMin)
		})
	}
}
