package collection

import (
	"context"
	"errors"
	"testing"

	"github.com/Karitham/corde"
	"go.uber.org/mock/gomock"

	"github.com/karitham/waifubot/storage/userstore"
)

func TestTransferTokens(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name       string
		from       corde.Snowflake
		to         corde.Snowflake
		amount     int32
		setupMocks func(*MockProfileStore, *MockProfileStore, *MockUserQuerier)
		wantErr    bool
		wantErrIs  error
	}{
		{
			name:   "successful transfer",
			from:   1,
			to:     2,
			amount: 50,
			setupMocks: func(store, tx *MockProfileStore, user *MockUserQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().UserStore().Return(user).AnyTimes()
				user.EXPECT().UpdateTokens(gomock.Any(), userstore.UpdateTokensParams{
					UserID: 1,
					Tokens: -50,
				}).Return(userstore.User{}, nil)
				user.EXPECT().UpdateTokens(gomock.Any(), userstore.UpdateTokensParams{
					UserID: 2,
					Tokens: 50,
				}).Return(userstore.User{}, nil)
				tx.EXPECT().Commit(gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name:       "zero amount",
			from:       1,
			to:         2,
			amount:     0,
			setupMocks: func(store, tx *MockProfileStore, user *MockUserQuerier) {},
			wantErr:    true,
			wantErrIs:  ErrInvalidAmount,
		},
		{
			name:       "negative amount",
			from:       1,
			to:         2,
			amount:     -10,
			setupMocks: func(store, tx *MockProfileStore, user *MockUserQuerier) {},
			wantErr:    true,
			wantErrIs:  ErrInvalidAmount,
		},
		{
			name:   "insufficient funds - source goes below zero",
			from:   1,
			to:     2,
			amount: 100,
			setupMocks: func(store, tx *MockProfileStore, user *MockUserQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().UserStore().Return(user).AnyTimes()
				user.EXPECT().UpdateTokens(gomock.Any(), userstore.UpdateTokensParams{
					UserID: 1,
					Tokens: -100,
				}).Return(userstore.User{Tokens: -1}, nil)
				tx.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr:   true,
			wantErrIs: ErrInsufficientTokens,
		},
		{
			name:   "Tx error",
			from:   1,
			to:     2,
			amount: 50,
			setupMocks: func(store, tx *MockProfileStore, user *MockUserQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(nil, errors.New("tx error"))
			},
			wantErr: true,
		},
		{
			name:   "recipient update error triggers rollback",
			from:   1,
			to:     2,
			amount: 50,
			setupMocks: func(store, tx *MockProfileStore, user *MockUserQuerier) {
				store.EXPECT().Tx(gomock.Any()).Return(tx, nil)
				tx.EXPECT().UserStore().Return(user).AnyTimes()
				user.EXPECT().UpdateTokens(gomock.Any(), userstore.UpdateTokensParams{
					UserID: 1,
					Tokens: -50,
				}).Return(userstore.User{}, nil)
				user.EXPECT().UpdateTokens(gomock.Any(), userstore.UpdateTokensParams{
					UserID: 2,
					Tokens: 50,
				}).Return(userstore.User{}, errors.New("something went wrong"))
				tx.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr: true,
		},
		{
			name:       "same user transfer",
			from:       1,
			to:         1,
			amount:     50,
			setupMocks: func(store, tx *MockProfileStore, user *MockUserQuerier) {},
			wantErr:    true,
			wantErrIs:  ErrSameUserTransfer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMockProfileStore(ctrl)
			tx := NewMockProfileStore(ctrl)
			user := NewMockUserQuerier(ctrl)
			tt.setupMocks(store, tx, user)

			err := TransferTokens(context.Background(), store, tt.from, tt.to, tt.amount)

			if tt.wantErr {
				if err == nil {
					t.Errorf("TransferTokens() expected error, got nil")
					return
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("TransferTokens() error = %v, want error is %v", err, tt.wantErrIs)
				}
			} else {
				if err != nil {
					t.Errorf("TransferTokens() unexpected error = %v", err)
				}
			}
		})
	}
}
