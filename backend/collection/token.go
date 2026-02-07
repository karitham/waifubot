package collection

import (
	"context"
	"errors"
	"fmt"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/storage/userstore"
)

var (
	ErrInsufficientTokens = errors.New("insufficient tokens")
	ErrInvalidAmount      = errors.New("amount must be positive")
	ErrSameUserTransfer   = errors.New("cannot transfer to yourself")
)

func TransferTokens(ctx context.Context, store Store, from, to corde.Snowflake, amount int32) (err error) {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if from == to {
		return ErrSameUserTransfer
	}

	tx, err := store.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	fromUser, err := tx.UserStore().UpdateTokens(ctx, userstore.UpdateTokensParams{
		Tokens: -amount,
		UserID: uint64(from),
	})
	if err != nil {
		return fmt.Errorf("failed to update source token count: %w", err)
	}

	if fromUser.Tokens < 0 {
		return ErrInsufficientTokens
	}

	_, err = tx.UserStore().UpdateTokens(ctx, userstore.UpdateTokensParams{
		Tokens: amount,
		UserID: uint64(to),
	})
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
