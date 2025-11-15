package discord

import (
	"context"
	"log/slog"
	"time"

	"github.com/Karitham/corde"
)

func trace[T corde.InteractionDataConstraint](next func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T])) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
	return func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
		start := time.Now()
		l := slog.With(
			"route", i.Route,
			"guild", i.GuildID,
			"channel", i.ChannelID,
			"user", i.Member.User.ID,
		)

		ctx = context.WithValue(ctx, slog.Default().Handler(), l)
		next(ctx, w, i)

		l.Info("request completed", "took", time.Since(start).String())
	}
}
