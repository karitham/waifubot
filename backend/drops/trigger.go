package drops

import (
	"context"
	"fmt"
)

// Trigger records an activity on the channel. When the channel's count
// reaches the threshold, it resets the count and stores a new drop, replacing
// any previous one. It reports whether a drop was created.
func Trigger(ctx context.Context, store Store, channel ChannelID, threshold int64) (Drop, bool, error) {
	count, err := store.IncrementActivity(ctx, channel)
	if err != nil {
		return Drop{}, false, fmt.Errorf("failed to record channel activity: %w", err)
	}

	if count < threshold {
		return Drop{}, false, nil
	}

	// The drop consumes the accumulated activity.
	if err := store.ResetActivity(ctx, channel); err != nil {
		return Drop{}, false, fmt.Errorf("failed to reset channel activity: %w", err)
	}

	char, err := store.RandomActiveChar(ctx, DropWeightExponent)
	if err != nil {
		return Drop{}, false, fmt.Errorf("failed to pick a drop character: %w", err)
	}

	drop := Drop{Channel: channel, Character: char}
	if err := store.SetDrop(ctx, drop); err != nil {
		return Drop{}, false, fmt.Errorf("failed to store the drop: %w", err)
	}

	return drop, true, nil
}
