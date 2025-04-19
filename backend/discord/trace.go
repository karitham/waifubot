package discord

import (
	"context"
	"time"

	"github.com/Karitham/corde"
	"github.com/rs/zerolog/log"
)

func trace[T corde.InteractionDataConstraint](next func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T])) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
	return func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[T]) {
		start := time.Now()
		l := log.With().
			Str("route", i.Route).
			Stringer("guild", i.GuildID).
			Stringer("channel", i.ChannelID).
			Stringer("user", i.Member.User.ID).
			Logger()

		ctx = l.WithContext(ctx)
		next(ctx, w, i)

		l.Debug().Str("took", time.Since(start).String()).Send()
	}
}
