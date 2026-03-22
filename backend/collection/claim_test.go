package collection_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/collectiontest"
)

func TestClaim(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(m *collectiontest.MockStore)
		userID     uint64
		channelID  uint64
		charName   string
		wantErr    error
		wantCharID int64
		wantSource string
		wantUserID uint64
	}{
		{
			name: "success",
			setup: func(m *collectiontest.MockStore) {
				m.GetDropForUpdateFunc = func(_ context.Context, _ uint64) (collection.Drop, error) {
					return collection.Drop{ID: 1, Name: "Test Character", Image: "test.jpg", MediaTitle: "Test Anime"}, nil
				}
				m.GetUserFunc = func(_ context.Context, userID uint64) (collection.User, error) {
					return collection.User{UserID: userID}, nil
				}
				m.UpsertCharacterFunc = func(_ context.Context, _ catalog.Character) error { return nil }
				m.AddToCollectionFunc = func(_ context.Context, _ uint64, _ collection.Character, _ string, _ time.Time) error {
					return nil
				}
				m.RemoveFromWishlistFunc = func(_ context.Context, _ uint64, _ int64) error { return nil }
				m.DeleteDropFunc = func(_ context.Context, _ uint64) error { return nil }
			},
			userID:     123,
			channelID:  456,
			charName:   "Test Character",
			wantCharID: 1,
		},
		{
			name: "no_drop_in_channel",
			setup: func(m *collectiontest.MockStore) {
				m.GetDropForUpdateFunc = func(_ context.Context, _ uint64) (collection.Drop, error) {
					return collection.Drop{}, collection.ErrNotFound
				}
			},
			userID:    123,
			channelID: 456,
			charName:  "Test",
			wantErr:   collection.ErrNoDropInChannel,
		},
		{
			name: "wrong_name",
			setup: func(m *collectiontest.MockStore) {
				m.GetDropForUpdateFunc = func(_ context.Context, _ uint64) (collection.Drop, error) {
					return collection.Drop{ID: 1, Name: "Test Character"}, nil
				}
			},
			userID:    123,
			channelID: 456,
			charName:  "Wrong Name",
			wantErr:   collection.ErrWrongCharacterName,
		},
		{
			name: "already_owned",
			setup: func(m *collectiontest.MockStore) {
				m.GetDropForUpdateFunc = func(_ context.Context, _ uint64) (collection.Drop, error) {
					return collection.Drop{ID: 1, Name: "Test Character"}, nil
				}
				m.GetUserFunc = func(_ context.Context, userID uint64) (collection.User, error) {
					return collection.User{UserID: userID}, nil
				}
				m.UpsertCharacterFunc = func(_ context.Context, _ catalog.Character) error { return nil }
				m.AddToCollectionFunc = func(_ context.Context, _ uint64, _ collection.Character, _ string, _ time.Time) error {
					return collection.ErrAlreadyOwned
				}
			},
			userID:    123,
			channelID: 456,
			charName:  "Test Character",
			wantErr:   collection.ErrAlreadyOwned,
		},
		{
			name: "new_user",
			setup: func(m *collectiontest.MockStore) {
				m.GetDropForUpdateFunc = func(_ context.Context, _ uint64) (collection.Drop, error) {
					return collection.Drop{ID: 1, Name: "Test Character", Image: "test.jpg"}, nil
				}
				m.GetUserFunc = func(_ context.Context, _ uint64) (collection.User, error) {
					return collection.User{}, collection.ErrNotFound
				}
				m.CreateUserFunc = func(_ context.Context, _ uint64) error { return nil }
				m.UpsertCharacterFunc = func(_ context.Context, _ catalog.Character) error { return nil }
				m.AddToCollectionFunc = func(_ context.Context, _ uint64, _ collection.Character, _ string, _ time.Time) error {
					return nil
				}
				m.RemoveFromWishlistFunc = func(_ context.Context, _ uint64, _ int64) error { return nil }
				m.DeleteDropFunc = func(_ context.Context, _ uint64) error { return nil }
			},
			userID:     123,
			channelID:  456,
			charName:   "Test Character",
			wantCharID: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &collectiontest.MockStore{}
			if tt.setup != nil {
				tt.setup(store)
			}

			char, err := collection.Claim(t.Context(), store, tt.userID, tt.channelID, tt.charName)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, 1, store.RollbackCalls)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantCharID, char.ID)
			assert.Equal(t, 1, store.CommitCalls)
		})
	}
}

func TestExchange(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(m *collectiontest.MockStore)
		userID     uint64
		charID     int64
		wantErr    error
		wantCharID int64
	}{
		{
			name: "success",
			setup: func(m *collectiontest.MockStore) {
				m.GetCharacterByIDFunc = func(_ context.Context, _ int64) (catalog.Character, error) {
					return catalog.Character{ID: 1, Name: "Char1", Image: "img1"}, nil
				}
				m.GetOwnedCharacterFunc = func(_ context.Context, _ uint64, _ int64) (collection.OwnedCharacter, error) {
					return collection.OwnedCharacter{
						Character: collection.Character{ID: 1, Name: "Char1"},
						Date:      time.Now(),
						Source:    "ROLL",
						UserID:    123,
					}, nil
				}
				m.RemoveFromCollectionFunc = func(_ context.Context, _ uint64, _ int64) error { return nil }
				m.AddTokensFunc = func(_ context.Context, _ uint64, _ int32) (collection.User, error) {
					return collection.User{UserID: 123, Tokens: 6}, nil
				}
			},
			userID:     123,
			charID:     1,
			wantCharID: 1,
		},
		{
			name: "not_owned",
			setup: func(m *collectiontest.MockStore) {
				m.GetCharacterByIDFunc = func(_ context.Context, _ int64) (catalog.Character, error) {
					return catalog.Character{ID: 1}, nil
				}
				m.GetOwnedCharacterFunc = func(_ context.Context, _ uint64, _ int64) (collection.OwnedCharacter, error) {
					return collection.OwnedCharacter{}, collection.ErrNotFound
				}
			},
			userID:  123,
			charID:  1,
			wantErr: collection.ErrUserDoesNotOwnCharacter,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &collectiontest.MockStore{}
			if tt.setup != nil {
				tt.setup(store)
			}

			char, err := collection.Exchange(t.Context(), store, tt.userID, tt.charID)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, 1, store.RollbackCalls)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantCharID, char.ID)
			assert.Equal(t, 1, store.CommitCalls)
		})
	}
}

func TestGive(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(m *collectiontest.MockStore)
		from        uint64
		to          uint64
		charID      int64
		wantErr     bool
		errContains string
		wantSource  string
		wantUserID  uint64
	}{
		{
			name: "success",
			setup: func(m *collectiontest.MockStore) {
				m.GetOwnedCharacterFunc = func(_ context.Context, userID uint64, _ int64) (collection.OwnedCharacter, error) {
					if userID == 123 {
						return collection.OwnedCharacter{
							Character: collection.Character{ID: 1, Name: "Char1", Image: "img1"},
						}, nil
					}
					return collection.OwnedCharacter{}, collection.ErrNotFound
				}
				m.GiveCharacterFunc = func(_ context.Context, _, _ uint64, _ int64) (collection.OwnedCharacter, error) {
					return collection.OwnedCharacter{
						Character: collection.Character{ID: 1, Name: "Char1", Image: "img1"},
						Source:    "TRADE",
						UserID:    456,
					}, nil
				}
				m.RemoveFromWishlistFunc = func(_ context.Context, _ uint64, _ int64) error { return nil }
			},
			from:       123,
			to:         456,
			charID:     1,
			wantSource: "TRADE",
			wantUserID: 456,
		},
		{
			name: "not_owned",
			setup: func(m *collectiontest.MockStore) {
				m.GetOwnedCharacterFunc = func(_ context.Context, _ uint64, _ int64) (collection.OwnedCharacter, error) {
					return collection.OwnedCharacter{}, collection.ErrNotFound
				}
			},
			from:        123,
			to:          456,
			charID:      1,
			wantErr:     true,
			errContains: "does not own",
		},
		{
			name: "target_already_owns",
			setup: func(m *collectiontest.MockStore) {
				m.GetOwnedCharacterFunc = func(_ context.Context, _ uint64, _ int64) (collection.OwnedCharacter, error) {
					return collection.OwnedCharacter{
						Character: collection.Character{ID: 1, Name: "Char1"},
					}, nil
				}
			},
			from:        123,
			to:          456,
			charID:      1,
			wantErr:     true,
			errContains: "already owns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &collectiontest.MockStore{}
			if tt.setup != nil {
				tt.setup(store)
			}

			char, err := collection.Give(t.Context(), store, tt.from, tt.to, tt.charID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Equal(t, 1, store.RollbackCalls)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantSource, char.Source)
			assert.Equal(t, tt.wantUserID, char.UserID)
			assert.Equal(t, 1, store.CommitCalls)
		})
	}
}

func TestTransferTokens(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *collectiontest.MockStore)
		from    uint64
		to      uint64
		amount  int32
		wantErr error
	}{
		{
			name: "success",
			setup: func(m *collectiontest.MockStore) {
				m.SpendTokensFunc = func(_ context.Context, _ uint64, _ int32) (collection.User, error) {
					return collection.User{UserID: 1, Tokens: 50}, nil
				}
				m.AddTokensFunc = func(_ context.Context, _ uint64, _ int32) (collection.User, error) {
					return collection.User{UserID: 2, Tokens: 50}, nil
				}
			},
			from:   1,
			to:     2,
			amount: 50,
		},
		{
			name: "insufficient_funds",
			setup: func(m *collectiontest.MockStore) {
				m.SpendTokensFunc = func(_ context.Context, _ uint64, _ int32) (collection.User, error) {
					return collection.User{}, collection.ErrInsufficientTokens
				}
			},
			from:    1,
			to:      2,
			amount:  100,
			wantErr: collection.ErrInsufficientTokens,
		},
		{
			name:    "invalid_amount_zero",
			from:    1,
			to:      2,
			amount:  0,
			wantErr: collection.ErrInvalidAmount,
		},
		{
			name:    "invalid_amount_negative",
			from:    1,
			to:      2,
			amount:  -5,
			wantErr: collection.ErrInvalidAmount,
		},
		{
			name:    "same_user",
			from:    1,
			to:      1,
			amount:  50,
			wantErr: collection.ErrSameUserTransfer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &collectiontest.MockStore{}
			if tt.setup != nil {
				tt.setup(store)
			}

			err := collection.TransferTokens(t.Context(), store, tt.from, tt.to, tt.amount)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, 1, store.CommitCalls)
		})
	}
}
