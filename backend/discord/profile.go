package discord

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
)

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
	logger := slog.With("user_id", uint64(i.Member.User.ID), "guild_id", uint64(i.GuildID))

	user := i.Member.User
	if len(i.Data.Resolved.Users) > 0 {
		user = i.Data.Resolved.Users.First()
	}

	data, err := collection.UserProfile(ctx, b.Store, user.ID)
	if err != nil {
		logger.Error("error getting profile", "error", err, "target_user_id", uint64(user.ID))
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
		).
		Field("Collection", fmt.Sprintf("[View Collection](https://waifugui.karitham.dev/#/list/%s)", user.ID.String())).
		Field("Wishlist", fmt.Sprintf("[View Wishlist](https://waifugui.karitham.dev/#/wishlist/%s)", user.ID.String()))
	if data.Favorite.Image != "" {
		resp.Thumbnail(corde.Image{URL: data.Favorite.Image})
	}

	w.Respond(resp)
}

func (b *Bot) profileEditFavorite(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	logger := slog.With("user_id", uint64(i.Member.User.ID), "guild_id", uint64(i.GuildID))

	optID, _ := i.Data.Options.Int64("id")
	if err := collection.SetFavorite(ctx, b.Store, i.Member.User.ID, optID); err != nil {
		logger.Error("error setting favorite character", "error", err, "character_id", optID)
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

	chars, err := collection.SearchCharacters(ctx, b.Store, i.Member.User.ID, id)
	if err != nil {
		slog.Error("Error getting user's characters", "error", err, "user", i.Member.User.ID)
		return
	}

	resp := corde.NewResp()
	for _, c := range chars {
		resp.Choice(fmt.Sprintf("%s (%d)", c.Name, c.ID), c.ID)
	}

	w.Autocomplete(resp)
}

func (b *Bot) profileEditAnilistURL(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	anilistURL, _ := i.Data.Options.String("url")
	if err := collection.SetAnilistURL(ctx, b.Store, i.Member.User.ID, anilistURL); err != nil {
		w.Respond(corde.NewResp().Content(err.Error()).Ephemeral())
		return
	}

	w.Respond(corde.NewResp().Contentf("Anilist URL set as %s", anilistURL).Ephemeral())
}

func (b *Bot) profileEditQuote(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	quote, _ := i.Data.Options.String("value")
	if err := collection.SetQuote(ctx, b.Store, i.Member.User.ID, quote); err != nil {
		w.Respond(corde.NewResp().Content(err.Error()).Ephemeral())
		return
	}

	w.Respond(corde.NewResp().Content("Quote set").Ephemeral())
}
