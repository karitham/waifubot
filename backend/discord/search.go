package discord

import (
	"context"
	"log/slog"
	"regexp"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
)

func (b *Bot) search(m *corde.Mux) {
	t := trace[corde.SlashCommandInteractionData]
	i := interact(b.InterStore, onInteraction[corde.SlashCommandInteractionData](b))
	idx := indexMiddleware[corde.SlashCommandInteractionData](b)

	m.SlashCommand("char", wrap(b.SearchChar, t, i, idx))
	m.SlashCommand("user", wrap(b.SearchUser, t, i, idx))
	m.SlashCommand("manga", wrap(b.SearchManga, t, i, idx))
	m.SlashCommand("anime", wrap(b.SearchAnime, t, i, idx))
}

func (b *Bot) SearchAnime(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	search, _ := i.Data.Options.String("name")

	media, err := b.AnimeService.Anime(ctx, search)
	if err != nil {
		slog.ErrorContext(ctx, "error with anime service", "error", err)
		w.Respond(rspErr("Error searching for this anime, either it doesn't exist or something went wrong"))
		return
	}

	if len(media) == 0 {
		w.Respond(rspErr("No anime found with that name"))
		return
	}

	w.Respond(mediaEmbed(media[0]))
}

func (b *Bot) SearchManga(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	search, _ := i.Data.Options.String("name")

	media, err := b.AnimeService.Manga(ctx, search)
	if err != nil {
		slog.ErrorContext(ctx, "error with anime service", "error", err)
		w.Respond(rspErr("Error searching for this manga, either it doesn't exist or something went wrong"))
		return
	}

	if len(media) == 0 {
		w.Respond(rspErr("No manga found with that name"))
		return
	}

	w.Respond(mediaEmbed(media[0]))
}

func (b *Bot) SearchUser(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	search, _ := i.Data.Options.String("name")

	users, err := b.AnimeService.User(ctx, search)
	if err != nil {
		slog.ErrorContext(ctx, "error with user service", "error", err)
		w.Respond(rspErr("Error searching for this user, either it doesn't exist or something went wrong"))
		return
	}

	if len(users) == 0 {
		w.Respond(rspErr("No user found with that name"))
		return
	}

	w.Respond(userEmbed(users[0]))
}

func (b *Bot) SearchChar(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	search, _ := i.Data.Options.String("name")

	characters, err := b.AnimeService.Character(ctx, search)
	if err != nil {
		slog.ErrorContext(ctx, "error with char service", "error", err)
		w.Respond(rspErr("Error searching for this character, either it doesn't exist or something went wrong"))
		return
	}

	if len(characters) == 0 {
		w.Respond(rspErr("No character found with that name"))
		return
	}

	w.Respond(charEmbed(characters[0]))
}

func mediaEmbed(m collection.Media) *corde.EmbedB {
	return applyEmbedOpt(corde.NewEmbed().
		Title(m.Title).
		URL(m.URL).
		Color(m.CoverImageColor).
		ImageURL(m.BannerImageURL).
		Thumbnail(corde.Image{URL: m.CoverImageURL}).
		Description(m.Description),
		anilistFooter,
		sanitizeDescOpt,
	)
}

func userEmbed(u collection.TrackerUser) *corde.EmbedB {
	return applyEmbedOpt(corde.NewEmbed().
		Title(u.Name).
		URL(u.URL).
		Color(AnilistColor).
		ImageURL(u.ImageURL).
		Description(u.About),
		anilistFooter,
		sanitizeDescOpt,
	)
}

func charEmbed(c collection.MediaCharacter) *corde.EmbedB {
	return applyEmbedOpt(corde.NewEmbed().
		Title(c.Name).
		Color(AnilistColor).
		URL(c.URL).
		Thumbnail(corde.Image{URL: c.ImageURL}).
		Description(c.Description),
		anilistFooter,
		sanitizeDescOpt,
	)
}

func anilistFooter(b *corde.EmbedB) *corde.EmbedB {
	return b.Footer(corde.Footer{
		Text:    "View on Anilist",
		IconURL: AnilistIconURL,
	})
}

func applyEmbedOpt(b *corde.EmbedB, opts ...func(*corde.EmbedB) *corde.EmbedB) *corde.EmbedB {
	for _, opt := range opts {
		b = opt(b)
	}
	return b
}

func sanitizeDescOpt(b *corde.EmbedB) *corde.EmbedB {
	return b.Description(sanitizeDesc(b.Embed().Description))
}

var descRegexp = regexp.MustCompile(`!~|~!`)

func sanitizeDesc(desc string) string {
	return descRegexp.ReplaceAllString(desc, "||")
}
