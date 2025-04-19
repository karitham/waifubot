package discord

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Karitham/corde"
	"github.com/rs/zerolog/log"
)

func (b *Bot) drop(ctx context.Context, channelID corde.Snowflake) {
	char, err := b.AnimeService.RandomChar(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get random character")
		return
	}

	err = b.Inter.SetChannelChar(ctx, channelID, char)
	if err != nil {
		log.Error().Err(err).Msg("failed to set channel character")
		return
	}

	log.Trace().Msgf("dropped %s", char.Name)

	msg, cleanup := DropEmbed(ctx, char)
	defer cleanup()
	_, err = b.mux.CreateMessage(channelID, msg)
	if err != nil {
		log.Error().Err(err).Msg("failed to create message")
		return
	}
}

func (b *Bot) claim(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	name, err := i.Data.Options.String("name")
	if err != nil || name == "" {
		w.Respond(rspErr("enter a name to claim the character"))
		return
	}

	// impl claim.
	char, err := b.Inter.GetChannelChar(ctx, i.ChannelID)
	if err != nil {
		log.Trace().Err(err).Msg("failed to get channel character")
		w.Respond(rspErr("No character to claim"))
		return
	}

	if !equalStrings(char.Name, name) {
		w.Respond(rspErr("Wrong!"))
		return
	}

	err = b.Store.PutChar(ctx, i.Member.User.ID, Character{
		Date:   time.Now(),
		Image:  char.ImageURL,
		Name:   sanitizeName(char.Name),
		Type:   "CLAIM",
		UserID: i.Member.User.ID,
		ID:     char.ID,
	})
	if err != nil {
		log.Debug().Err(err).Msg("failed to put character")
		w.Respond(rspErr("you already have this character!"))
		return
	}

	err = b.Inter.RemoveChannelChar(ctx, i.ChannelID)
	if err != nil {
		log.Error().Err(err).Msg("failed to remove channel character")
		return
	}

	w.Respond(corde.NewEmbed().
		Title(char.Name).
		URL(char.URL).
		Footer(corde.Footer{IconURL: AnilistIconURL, Text: "View them on anilist"}).
		Thumbnail(corde.Image{URL: char.ImageURL}).
		Descriptionf("Congratulations!\nYou just claimed %s (%d)!\nIt appears in :\n- %s", char.Name, char.ID, char.MediaTitle),
	)
}

func equalStrings(this, that string) bool {
	return strings.EqualFold(sanitizeName(this), sanitizeName(that))
}

func sanitizeName(name string) string {
	return strings.Join(strings.Fields(name), " ")
}

func DropEmbed(ctx context.Context, char MediaCharacter) (corde.Message, func()) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, char.ImageURL, nil)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("failed to create request")
		return corde.Message{}, func() {}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("failed to get image")
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
				Description: "Can you guess which character this is?\nUse `/claim name` to claim the character.\n\n**Hint:** " + initials.String(),
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
