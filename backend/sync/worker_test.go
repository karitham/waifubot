package sync

import (
	"context"
	"errors"
	"math/rand"
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
		name       string
		fetchErr   error
		fetchChars []collection.MediaCharacter
		wantErr    bool
		wantCount  int
	}{
		{
			name:       "characters found",
			fetchChars: []collection.MediaCharacter{{ID: 1, Name: "A"}, {ID: 2, Name: "B"}},
			wantCount:  2,
		},
		{
			name:       "no matches",
			fetchChars: nil,
			wantCount:  0,
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
			store := &collectiontest.MockStore{
				UpsertCharacterFunc: func(_ context.Context, _ catalog.Character) error {
					upserted++
					return nil
				},
			}
			fetcher := &mockFetcher{
				CharactersByIDsFunc: func(_ context.Context, _ []int64) ([]collection.MediaCharacter, error) {
					return tt.fetchChars, tt.fetchErr
				},
			}
			svc := &Service{Store: store, Anilist: fetcher, rng: rand.New(rand.NewSource(0))}
			err := svc.Sync(t.Context())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantCount, upserted)
			}
		})
	}
}
