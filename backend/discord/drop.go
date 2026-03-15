package discord

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

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
	})
	if err != nil {
		logger.Error("failed to set channel character", "error", err, "character_id", char.ID, "character_name", char.Name)
		return
	}

	logger.Debug("dropped character", "character_id", char.ID, "character_name", char.Name)

	msg, cleanup := DropEmbed(ctx, char)
	defer cleanup()
	_, err = b.mux.CreateMessage(channelID, msg)
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

	char, err := collection.Claim(ctx, b.Store, uint64(i.Member.User.ID), uint64(i.ChannelID), sanitizeName(name))
	if err != nil {
		logger.Debug("failed to claim", "error", err)
		switch {
		case errors.Is(err, collection.ErrNoDropInChannel):
			w.Respond(rspErr("No character to claim"))
		case errors.Is(err, collection.ErrWrongCharacterName):
			w.Respond(rspErr("Wrong!"))
		default:
			w.Respond(rspErr("Failed to claim character"))
		}
		return
	}

	logger.Info("user claimed character",
		"character_id", char.ID,
		"character_name", char.Name)

	w.Respond(corde.NewEmbed().
		Title(char.Name).
		URL(fmt.Sprintf("https://anilist.co/character/%d", char.ID)).
		Footer(corde.Footer{IconURL: AnilistIconURL, Text: "View on Anilist"}).
		Thumbnail(corde.Image{URL: char.Image}).
		Descriptionf("Congratulations!\n%s added to your collection!\nID: %d", char.Name, char.ID),
	)
}

func sanitizeName(name string) string {
	return strings.Join(strings.Fields(name), " ")
}

func DropEmbed(ctx context.Context, char collection.MediaCharacter) (corde.Message, func()) {
	logger := slog.With("character_id", char.ID, "character_name", char.Name, "image_url", char.ImageURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, char.ImageURL, nil)
	if err != nil {
		logger.Error("failed to create image request for drop embed", "error", err)
		return corde.Message{}, func() {}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error("failed to fetch image for drop embed", "error", err)
		return corde.Message{}, func() {}
	}

	parts := strings.Fields(char.Name)
	initials := strings.Builder{}
	for i, part := range parts {
		if i != 0 {
			initials.WriteRune('.')
		}
		initials.WriteRune([]rune(part)[0])
	}

	return corde.Message{
			Embeds: []corde.Embed{{
				Title:       "Character Drop!",
				Description: "Can you guess who this is?\nUse `/claim name` to add them to your collection.\n\n**Hint:** " + initials.String(),
				Image:       corde.Image{URL: "attachment://image.png"},
			}},
			Attachments: []corde.Attachment{{
				Filename: "image.png",
				Body:     resp.Body,
			}},
		}, func() {
			_ = resp.Body.Close()
		}
}
