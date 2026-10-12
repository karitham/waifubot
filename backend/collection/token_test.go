package collection_test

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/collectiontest"
)

// tokenAmount builds a TokenAmount for tests that need a known-positive amount.
func tokenAmount(t *testing.T, n int) collection.TokenAmount {
	t.Helper()

	amount, err := collection.NewTokenAmount(n)
	require.NoError(t, err)
	return amount
}

func TestTokenAmountValid(t *testing.T) {
	assert.False(t, collection.TokenAmount{}.Valid())
	assert.True(t, tokenAmount(t, 1).Valid())
}

func TestBalance_AbsentUser(t *testing.T) {
	store := &collectiontest.MockStore{
		GetUserFunc: func(_ context.Context, _ collection.UserID) (collection.User, error) {
			return collection.User{}, collection.ErrNotFound
		},
	}

	tokens, err := collection.Balance(t.Context(), store, 7)
	require.NoError(t, err)
	assert.Equal(t, int32(0), tokens)
}

func TestNewTokenAmount(t *testing.T) {
	tests := []struct {
		name    string
		input   int
		want    int32
		wantErr error
	}{
		{name: "positive", input: 5, want: 5},
		{name: "max int32", input: math.MaxInt32, want: math.MaxInt32},
		{name: "zero", input: 0, wantErr: collection.ErrInvalidAmount},
		{name: "negative", input: -3, wantErr: collection.ErrInvalidAmount},
		{name: "above max int32", input: math.MaxInt32 + 1, wantErr: collection.ErrInvalidAmount},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := collection.NewTokenAmount(tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Int32())
		})
	}
}
