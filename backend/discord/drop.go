package discord

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage/dropstore"
)

func (r *Router) drop(ctx context.Context, channelID corde.Snowflake) {
	logger := slog.With("channel_id", uint64(channelID))

	catChar, err := r.Store.RandomActiveChar(ctx, collection.DropWeightExponent)
	if err != nil {
		logger.Error("failed to get random character for drop", "error", err)
		return
	}

	char := collection.MediaCharacter{
		ID:         catChar.ID,
		Name:       catChar.Name,
		ImageURL:   catChar.Image,
		MediaTitle: catChar.MediaTitle,
		Favorites:  catChar.Favorites,
	}

	err = r.DropStore.Set(ctx, channelID, dropstore.Drop{
		ID:         char.ID,
		Name:       char.Name,
		ImageURL:   char.ImageURL,
		MediaTitle: char.MediaTitle,
		Favorites:  char.Favorites,
	})
	if err != nil {
		logger.Error("failed to set channel character", "error", err, "character_id", char.ID, "character_name", char.Name)
		return
	}

	logger.Debug("dropped character", "character_id", char.ID, "character_name", char.Name)

	img, err := httpImageFetcher{doer: http.DefaultClient}.Fetch(ctx, char.ImageURL)
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
