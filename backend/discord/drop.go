package discord

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage/collectionstore"
)

func (b *Bot) drop(ctx context.Context, channelID corde.Snowflake) {
	logger := slog.With("channel_id", uint64(channelID))

	char, err := b.AnimeService.RandomChar(ctx)
	if err != nil {
		logger.Error("failed to get random character for drop", "error", err)
		return
	}

	err = b.DropStore.Set(ctx, channelID, char)
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

	// impl claim.
	char, err := b.DropStore.Get(ctx, i.ChannelID)
	if err != nil {
		logger.Debug("failed to get channel character for claim", "error", err)
		w.Respond(rspErr("No character to claim"))
		return
	}

	if !equalStrings(char.Name, name) {
		w.Respond(rspErr("Wrong!"))
		return
	}

	// Ensure user exists
	_, err = b.Store.UserStore().Get(ctx, uint64(i.Member.User.ID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = b.Store.UserStore().Create(ctx, uint64(i.Member.User.ID))
			if err != nil {
				w.Respond(rspErr("Failed to create user"))
				return
			}
		} else {
			w.Respond(rspErr("Failed to get user"))
			return
		}
	}

	// Insert char
	err = b.Store.CollectionStore().Insert(ctx, collectionstore.InsertParams{
		ID:     char.ID,
		UserID: uint64(i.Member.User.ID),
		Image:  char.ImageURL,
		Name:   sanitizeName(char.Name),
		Type:   "CLAIM",
	})
	if err != nil {
		logger.Debug("failed to insert claimed character", "error", err, "character_id", char.ID, "character_name", char.Name)
		w.Respond(rspErr("Already in your collection!"))
		return
	}

	err = b.DropStore.Delete(ctx, i.ChannelID)
	if err != nil {
		logger.Error("failed to remove channel character after claim", "error", err, "character_id", char.ID, "character_name", char.Name)
		return
	}

	logger.Info("user claimed character",
		"character_id", char.ID,
		"character_name", char.Name,
		"media_title", char.MediaTitle)

	w.Respond(corde.NewEmbed().
		Title(char.Name).
		URL(char.URL).
		Footer(corde.Footer{IconURL: AnilistIconURL, Text: "View on Anilist"}).
		Thumbnail(corde.Image{URL: char.ImageURL}).
		Descriptionf("Congratulations!\n%s added to your collection!\nID: %d\nFrom: %s", char.Name, char.ID, char.MediaTitle),
	)
}

func equalStrings(this, that string) bool {
	return strings.EqualFold(sanitizeName(this), sanitizeName(that))
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
