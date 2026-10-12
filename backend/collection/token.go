package collection

import (
	"context"
	"errors"
	"fmt"
)

// TransferTokens transfers tokens between users.
func TransferTokens(ctx context.Context, store DB, from, to UserID, amount int32) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if from == to {
		return ErrSameUserTransfer
	}

	return withTx(ctx, store, func(tx Tx) error {
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
