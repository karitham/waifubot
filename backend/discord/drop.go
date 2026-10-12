package discord

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
)

func (r *Router) drop(ctx context.Context, channelID corde.Snowflake) {
	logger := slog.With("channel_id", uint64(channelID))

	char, err := collection.CreateDrop(ctx, r.Store, collection.ChannelID(channelID))
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

	_, err = r.mux.CreateMessage(channelID, dropMessage(char, img))
	if err != nil {
		logger.Error("failed to create drop message", "error", err, "character_id", char.ID, "character_name", char.Name)
		return
	}
}
