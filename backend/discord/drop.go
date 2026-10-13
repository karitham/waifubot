package discord

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
)

// drops posts a channel drop once the channel's activity reaches the threshold.
type drops struct {
	store        collection.DB
	interactions InteractionStore
	mux          *corde.Mux
	threshold    int64
	guildID      *corde.Snowflake
}

// trigger resets the channel's counter and posts a drop when the channel
// passed the activity threshold in the scoped guild.
func (d *drops) trigger(ctx context.Context, count int64, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	if count < d.threshold {
		return
	}

	if d.guildID != nil && *d.guildID != i.GuildID {
		return
	}

	_ = d.interactions.Reset(ctx, i.ChannelID)
	d.post(ctx, i.ChannelID)
}

// post creates a drop for the channel and announces it.
func (d *drops) post(ctx context.Context, channelID corde.Snowflake) {
	logger := slog.With("channel_id", uint64(channelID))

	char, err := collection.CreateDrop(ctx, d.store, collection.ChannelID(channelID))
	if err != nil {
		logger.Error("failed to create drop", "error", err)
		return
	}

	logger.Debug("dropped character", "character_id", char.ID, "character_name", char.Name)

	img, err := httpImageFetcher{doer: http.DefaultClient}.Fetch(ctx, char.Image)
	if err != nil {
		logger.Error("failed to fetch image for drop embed", "error", err)
		return
	}
	defer img.Close()

	_, err = d.mux.CreateMessage(channelID, dropMessage(char, img))
	if err != nil {
		logger.Error("failed to create drop message", "error", err, "character_id", char.ID, "character_name", char.Name)
		return
	}
}
