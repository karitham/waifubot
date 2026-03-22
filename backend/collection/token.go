package collection

import (
	"context"
	"errors"
	"fmt"
)

// TransferTokens transfers tokens between users.
func TransferTokens(ctx context.Context, store Store, from, to UserID, amount int32) (err error) {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if from == to {
		return ErrSameUserTransfer
	}

	tx, err := store.WithTx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	_, err = tx.SpendTokens(ctx, from, amount)
	if err != nil {
		if errors.Is(err, ErrInsufficientTokens) {
			return err
		}
		return fmt.Errorf("failed to update source token count: %w", err)
	}

	_, err = tx.AddTokens(ctx, to, amount)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
