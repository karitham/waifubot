package drops

import (
	"context"
	"fmt"
)

// Trigger records an interaction on the channel. When the channel's count
// reaches the threshold, it resets the count and stores a new drop, replacing
// any previous one. It reports whether a drop was created.
//
// The whole trigger is one transaction and the increment holds the channel's
// counter row, so concurrent triggers serialize, the count is consumed
// exactly once, and a failure after the reset rolls the count back.
func Trigger(ctx context.Context, store DB, channel ChannelID, threshold int64) (Drop, bool, error) {
	tx, err := store.Begin(ctx)
	if err != nil {
		return Drop{}, false, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	count, err := tx.IncrementInteraction(ctx, channel)
	if err != nil {
		return Drop{}, false, fmt.Errorf("failed to record interaction: %w", err)
	}

	if count < threshold {
		if err := tx.Commit(ctx); err != nil {
			return Drop{}, false, fmt.Errorf("failed to commit interaction: %w", err)
		}
		return Drop{}, false, nil
	}

	// The drop consumes the accumulated interactions.
	if err := tx.ResetInteraction(ctx, channel); err != nil {
		return Drop{}, false, fmt.Errorf("failed to reset interactions: %w", err)
	}

	char, err := tx.RandomActiveChar(ctx, DropWeightExponent)
	if err != nil {
		return Drop{}, false, fmt.Errorf("failed to pick a drop character: %w", err)
	}

	drop := Drop{Channel: channel, Character: char}
	if err := tx.SetDrop(ctx, drop); err != nil {
		return Drop{}, false, fmt.Errorf("failed to store the drop: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Drop{}, false, fmt.Errorf("failed to commit drop: %w", err)
	}
	return drop, true, nil
}
