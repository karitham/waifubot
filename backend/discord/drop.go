package discord

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage/dropstore"
)

func (b *Bot) drop(ctx context.Context, channelID corde.Snowflake) {
	logger := slog.With("channel_id", uint64(channelID))

	char, err := b.AnimeService.RandomChar(ctx)
	if err != nil {
		logger.Error("failed to get random character for drop", "error", err)
		return
	}

	err = b.DropStore.Set(ctx, channelID, dropstore.Drop{
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

	_, err = b.mux.CreateMessage(channelID, dropMessage(char, img))
	if err != nil {
		logger.Error("failed to create drop message", "error", err, "character_id", char.ID, "character_name", char.Name)
		return
	}
}

func (b *Bot) claim(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	logger := slog.With("user_id", uint64(i.Member.User.ID), "channel_id", uint64(i.ChannelID), "guild_id", uint64(i.GuildID))

	name, err := i.Data.Options.String("name")
	if err != nil || name == "" {
		logger.Debug("claim: no name provided")
		w.Respond(rspErr("enter a name to claim the character"))
		return
	}

	char, err := collection.Claim(ctx, b.Store, uint64(i.Member.User.ID), uint64(i.ChannelID), collection.SanitizeName(name))
	if err != nil {
		logger.Debug("failed to claim", "error", err)
		switch {
		case errors.Is(err, collection.ErrNoDropInChannel):
			w.Respond(rspErr("No character to claim in this channel. Wait for the next drop!"))
		case errors.Is(err, collection.ErrWrongCharacterName):
			w.Respond(rspErr("Wrong name! Check the hint and try again."))
		case errors.Is(err, collection.ErrAlreadyOwned):
			w.Respond(rspErr("You already have this character in your collection!"))
		default:
			w.Respond(rspErr("Failed to claim character"))
		}
		return
	}

	logger.Info("user claimed character",
		"character_id", char.ID,
		"character_name", char.Name)

	rarity := collection.RarityFromFavorites(char.Favorites)

	w.Respond(claimEmbed(char, rarity.String()))
}
