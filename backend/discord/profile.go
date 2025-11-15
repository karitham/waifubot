package discord

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Karitham/corde"
)

type Profile struct {
	User
	CharacterCount int
	Favorite       Character
}

func (b *Bot) profile(m *corde.Mux) {
	m.SlashCommand("view", trace(b.profileView))
	m.Route("edit", func(m *corde.Mux) {
		m.SlashCommand("quote", trace(b.profileEditQuote))
		m.Route("favorite", func(m *corde.Mux) {
			m.SlashCommand("", trace(b.profileEditFavorite))
			m.Autocomplete("id", trace(b.userCollectionAutocomplete))
		})
		m.SlashCommand("anilist", trace(b.profileEditAnilistURL))
	})
}

func (b *Bot) profileView(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	user := i.Member.User
	if len(i.Data.Resolved.Users) > 0 {
		user = i.Data.Resolved.Users.First()
	}

	data, err := b.Store.ProfileOverview(ctx, user.ID)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting user's profile", "error", err)
		w.Respond(corde.NewResp().Content("An error occurred dialing the database, please try again later").Ephemeral())
		return
	}

	anilistURLDesc := ""
	if data.AnilistURL != "" {
		anilistURLDesc = fmt.Sprintf("Find them on [Anilist](%s)", data.AnilistURL)
	}

	resp := corde.NewEmbed().
		Title(user.Username).
		URL(fmt.Sprintf("https://waifugui.karitham.dev/#/list/%s", user.ID.String())).
		Descriptionf(
			"%s\n%s last rolled %s ago and has %d tokens.\nThey have %d characters.\nFavorite: %s\n%s",
			data.Quote,
			user.Username,
			time.Since(data.Date.UTC()).Truncate(time.Second),
			data.Tokens,
			data.CharacterCount,
			data.Favorite.Name,
			anilistURLDesc,
		)
	if data.Favorite.Image != "" {
		resp.Thumbnail(corde.Image{URL: data.Favorite.Image})
	}

	w.Respond(resp)
}

func (b *Bot) profileEditFavorite(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	optID, _ := i.Data.Options.Int64("id")
	err := b.Store.SetUserFavorite(ctx, i.Member.User.ID, optID)
	if err != nil {
		slog.ErrorContext(ctx, "Error setting user's favorite character", "error", err, "user", i.Member.User.ID, "character", optID)
		w.Respond(corde.NewResp().Content("An error occurred setting this character").Ephemeral())
		return
	}

	w.Respond(corde.NewResp().Contentf("Favorite character set as char id %d", optID).Ephemeral())
}

func (b *Bot) userCollectionAutocomplete(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.AutocompleteInteractionData]) {
	id, err := i.Data.Options.String("id")
	if err != nil {
		i, _ := i.Data.Options.Int("id")
		id = strconv.Itoa(i)
	}

	chars, err := b.Store.CharsStartingWith(ctx, i.Member.User.ID, id)
	if err != nil {
		slog.Error("Error getting user's characters", "error", err, "user", i.Member.User.ID)
		return
	}
	if len(chars) > 25 {
		chars = chars[25:]
	}

	resp := corde.NewResp()
	for _, c := range chars {
		resp.Choice(fmt.Sprintf("%s (%d)", c.Name, c.ID), c.ID)
	}

	w.Autocomplete(resp)
}

func (b *Bot) profileEditAnilistURL(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	anilistURL, _ := i.Data.Options.String("url")
	parsedURL, err := url.Parse(anilistURL)
	if err != nil {
		w.Respond(corde.NewResp().Content("Invalid URL").Ephemeral())
		return
	}

	if parsedURL.Host != "anilist.co" {
		w.Respond(corde.NewResp().Content("Invalid Anilist URL").Ephemeral())
		return
	}

	if !strings.HasPrefix(parsedURL.Path, "/user/") {
		w.Respond(corde.NewResp().Content("Invalid Anilist URL").Ephemeral())
		return
	}

	err = b.Store.SetUserAnilistURL(ctx, i.Member.User.ID, anilistURL)
	if err != nil {
		slog.ErrorContext(ctx, "Error setting user's anilist url", "error", err, "user", i.Member.User.ID)
		w.Respond(corde.NewResp().Content("An error occurred setting your anilist url").Ephemeral())
		return
	}

	w.Respond(corde.NewResp().Contentf("Anilist URL set as %s", anilistURL).Ephemeral())
}

func (b *Bot) profileEditQuote(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	quote, _ := i.Data.Options.String("value")
	if len(quote) > 1024 {
		w.Respond(corde.NewResp().Content("Quote is too long").Ephemeral())
		return
	}

	err := b.Store.SetUserQuote(ctx, i.Member.User.ID, quote)
	if err != nil {
		slog.ErrorContext(ctx, "Error setting user's quote", "error", err, "user", i.Member.User.ID, "quote", quote)
		w.Respond(corde.NewResp().Content("An error occurred setting this character").Ephemeral())
		return
	}

	w.Respond(corde.NewResp().Content("Quote set").Ephemeral())
}
