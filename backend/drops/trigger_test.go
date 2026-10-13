package drops_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/drops"
)

func TestDropMatches(t *testing.T) {
	tests := []struct {
		name    string
		dropped string
		input   string
		want    bool
	}{
		{"exact", "Naruto Uzumaki", "Naruto Uzumaki", true},
		{"different case", "Naruto Uzumaki", "naruto uzumaki", true},
		{"collapsed whitespace", "Naruto Uzumaki", "Naruto  Uzumaki", true},
		{"surrounding whitespace", "Naruto Uzumaki", "  Naruto Uzumaki  ", true},
		{"tabs and newlines", "Naruto Uzumaki", "\tNaruto \n Uzumaki", true},
		{"wrong name", "Naruto Uzumaki", "Sasuke", false},
		{"empty", "Naruto Uzumaki", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drop := drops.Drop{Character: catalog.Character{Name: tt.dropped}}
			assert.Equal(t, tt.want, drop.Matches(tt.input))
		})
	}
}

// mockStore implements drops.DB and drops.Tx and records the calls Trigger
// makes.
type mockStore struct {
	count        int64
	incrementErr error

	resets     int
	picked     catalog.Character
	pickErr    error
	weight     float64
	stored     drops.Drop
	storedSet  bool
	setDropErr error

	committed bool
}

func (m *mockStore) Begin(ctx context.Context) (drops.Tx, error) { return m, nil }

func (m *mockStore) IncrementInteraction(ctx context.Context, channel drops.ChannelID) (int64, error) {
	if m.incrementErr != nil {
		return 0, m.incrementErr
	}
	return m.count, nil
}

func (m *mockStore) ResetInteraction(ctx context.Context, channel drops.ChannelID) error {
	m.resets++
	return nil
}

func (m *mockStore) RandomActiveChar(ctx context.Context, weightExponent float64) (catalog.Character, error) {
	m.weight = weightExponent
	return m.picked, m.pickErr
}

func (m *mockStore) SetDrop(ctx context.Context, drop drops.Drop) error {
	m.stored = drop
	m.storedSet = true
	return m.setDropErr
}

func (m *mockStore) Commit(ctx context.Context) error {
	m.committed = true
	return nil
}

func (m *mockStore) Rollback(ctx context.Context) error { return nil }

func TestTrigger(t *testing.T) {
	const threshold = 3
	channel := drops.ChannelID(42)
	char := catalog.Character{ID: 7, Name: "Naruto"}

	tests := []struct {
		name          string
		count         int64
		incrementErr  error
		pickErr       error
		setDropErr    error
		wantTriggered bool
		wantErr       bool
		wantCommitted bool
		wantResets    int
		wantStored    bool
	}{
		{
			name:          "below threshold records only",
			count:         threshold - 1,
			wantCommitted: true,
		},
		{
			name:          "reaches threshold resets and stores",
			count:         threshold,
			wantTriggered: true,
			wantCommitted: true,
			wantResets:    1,
			wantStored:    true,
		},
		{
			name:         "increment fails",
			count:        threshold,
			incrementErr: errors.New("db down"),
			wantErr:      true,
		},
		{
			name:       "pick fails after reset",
			count:      threshold,
			pickErr:    errors.New("no characters"),
			wantErr:    true,
			wantResets: 1,
		},
		{
			name:       "store fails after reset",
			count:      threshold,
			setDropErr: errors.New("db down"),
			wantErr:    true,
			wantResets: 1,
			wantStored: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockStore{
				count:        tt.count,
				incrementErr: tt.incrementErr,
				picked:       char,
				pickErr:      tt.pickErr,
				setDropErr:   tt.setDropErr,
			}

			drop, triggered, err := drops.Trigger(t.Context(), store, channel, threshold)

			if tt.wantErr {
				require.Error(t, err)
				assert.False(t, triggered)
				assert.False(t, store.committed, "a failed trigger must roll back")
				assert.Equal(t, tt.wantResets, store.resets)
				assert.Equal(t, tt.wantStored, store.storedSet)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantTriggered, triggered)
			assert.Equal(t, tt.wantCommitted, store.committed)
			assert.Equal(t, tt.wantResets, store.resets)
			assert.Equal(t, tt.wantStored, store.storedSet)

			if tt.wantTriggered {
				assert.Equal(t, drops.Drop{Channel: channel, Character: char}, drop)
				assert.Equal(t, drops.DropWeightExponent, store.weight)
				assert.Equal(t, drop, store.stored)
			} else {
				assert.Equal(t, drops.Drop{}, drop)
			}
		})
	}
}
