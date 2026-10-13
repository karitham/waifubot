package discord

import (
	"context"
	"slices"

	"github.com/Karitham/corde"
)

// slashHandler is the handler shape for one slash command route.
type slashHandler = func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData])

// slashMiddleware wraps a slash command handler with another layer.
type slashMiddleware = func(next slashHandler) slashHandler

// onInteraction middleware runs fn in the background so the interaction
// response is not delayed.
func onInteraction(fn func(ctx context.Context, i *corde.Interaction[corde.SlashCommandInteractionData])) slashMiddleware {
	return func(next slashHandler) slashHandler {
		return func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
			go fn(context.Background(), i)
			next(ctx, w, i)
		}
	}
}

func wrap[T corde.InteractionDataConstraint](
	next func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]),
	fns ...func(func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T])) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]),
) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
	for _, fn := range slices.Backward(fns) {
		next = fn(next)
	}
	return next
}
