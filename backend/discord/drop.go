package discord

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/drops"
)

// dropTrigger posts a channel drop once the channel's interaction count
// reaches the threshold.
type dropTrigger struct {
	store     drops.DB
	mux       *corde.Mux
	threshold int64
	guildID   *corde.Snowflake
}

// record records a channel interaction and posts the drop it triggers.
func (d *dropTrigger) record(ctx context.Context, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	if d.guildID != nil && *d.guildID != i.GuildID {
		return
	}

	drop, triggered, err := drops.Trigger(ctx, d.store, drops.ChannelID(i.ChannelID), d.threshold)
	if err != nil {
		slog.Error("failed to record interaction", "error", err, "channel_id", uint64(i.ChannelID))
		return
	}
	if !triggered {
		return
	}

	d.post(ctx, drop)
}

// post announces the dropped character.
func (d *dropTrigger) post(ctx context.Context, drop drops.Drop) {
	logger := slog.With(
		"channel_id", uint64(drop.Channel),
		"character_id", drop.Character.ID,
		"character_name", drop.Character.Name,
	)
	logger.Debug("dropped character")

	img, err := httpImageFetcher{doer: http.DefaultClient}.Fetch(ctx, drop.Character.Image)
	if err != nil {
		logger.Error("failed to fetch image for drop embed", "error", err)
		return
	}
	defer img.Close()

	if _, err := d.mux.CreateMessage(corde.Snowflake(drop.Channel), dropMessage(drop.Character, img)); err != nil {
		logger.Error("failed to create drop message", "error", err)
	}
}
