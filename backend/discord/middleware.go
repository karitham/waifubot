package discord

import (
	"context"
	"log/slog"

	"github.com/Karitham/corde"
	"github.com/karitham/waifubot/storage/interactionstore"
)

// interaction middleware — increments interaction count and checks for drop trigger.
func interact[T corde.InteractionDataConstraint](inter interactionstore.Store, interact func(ctx context.Context, count int64, i *corde.Interaction[T])) func(func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T])) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
	return func(next func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T])) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
		return func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
			go func() {
				ctx := context.Background()

				err := inter.Increment(ctx, i.ChannelID)
				if err != nil {
					slog.Debug("failed to increment interaction count", "error", err)
				}

				count, err := inter.Get(ctx, i.ChannelID)
				if err != nil {
					slog.Error("failed to get interaction count", "error", err)
					return
				}

				interact(ctx, count, i)
			}()

			next(ctx, w, i)
		}
	}
}

func wrap[T corde.InteractionDataConstraint](
	next func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]),
	fns ...func(func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T])) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]),
) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
	for i := len(fns) - 1; i >= 0; i-- {
		next = fns[i](next)
	}
	return next
}
