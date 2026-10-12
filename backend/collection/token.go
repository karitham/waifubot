package collection

import (
	"context"
	"errors"
	"fmt"
	"math"
)

// TokenAmount is a positive token amount to spend or grant. Both operations
// accept only positive amounts: a negative spend would grant tokens and a
// negative grant would debit them, so the type represents amounts above zero.
// The zero value is invalid; code outside this package constructs amounts
// with NewTokenAmount.
type TokenAmount struct {
	amount int32
}

// NewTokenAmount converts an input into a TokenAmount. It returns
// ErrInvalidAmount when n is zero, negative, or above math.MaxInt32.
func NewTokenAmount(n int) (TokenAmount, error) {
	if n <= 0 || n > math.MaxInt32 {
		return TokenAmount{}, ErrInvalidAmount
	}
	return TokenAmount{amount: int32(n)}, nil
}

// Valid reports whether the amount is positive. The zero value is not valid.
func (a TokenAmount) Valid() bool { return a.amount > 0 }

// Int32 returns the amount in the representation used by storage.
func (a TokenAmount) Int32() int32 { return a.amount }

// TransferTokens transfers tokens between users.
func TransferTokens(ctx context.Context, store DB, from, to UserID, amount TokenAmount) error {
	if !amount.Valid() {
		return ErrInvalidAmount
	}
	if from == to {
		return ErrSameUserTransfer
	}

	return withUsers(ctx, store, []UserID{from, to}, func(tx Tx) error {
		if _, err := tx.SpendTokens(ctx, from, amount); err != nil {
			if errors.Is(err, ErrInsufficientTokens) {
				return err
			}
			return fmt.Errorf("failed to update source token count: %w", err)
		}

		if _, err := tx.AddTokens(ctx, to, amount); err != nil {
			return err
		}
		return nil
	})
}
