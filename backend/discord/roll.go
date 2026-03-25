package discord

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
)

func (b *Bot) roll(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	logger := slog.With("user_id", uint64(i.Member.User.ID), "guild_id", uint64(i.GuildID))

	config := collection.Config{
		RollCooldown: b.RollCooldown,
		TokensNeeded: b.TokensNeeded,
	}

	char, err := collection.Roll(ctx, b.Store, b.AnimeService, config, uint64(i.Member.User.ID))

	var cd collection.ErrRollCooldown
	switch {
	case errors.As(err, &cd):
		w.Respond(rspErr(cd.Error()))
		return
	case err != nil:
		logger.Error("error performing roll", "error", err)
		w.Respond(rspErr("An error occurred, please try again later"))
		return
	}

	wantingUsers, err := b.WishlistStore.GetUsersWantingCharacter(ctx, char.ID, uint64(i.GuildID), uint64(i.Member.User.ID))
	if err != nil {
		logger.Error("error getting users wanting character", "error", err)
	}

	w.Respond(corde.NewEmbed().
		Title(char.Name).
		URL(char.URL).
		Color(collection.GradientColor(char.Favorites)).
		Footer(corde.Footer{IconURL: AnilistIconURL, Text: "View on Anilist"}).
		Thumbnail(corde.Image{URL: char.ImageURL}).
		Descriptionf(
			"You got %s (%s)\n⭐ Rarity: %s | ❤️ %d favorites%s\n\nID: %d",
			char.Name,
			char.MediaTitle,
			char.Rarity(),
			char.Favorites,
			formatUsersWantingCharacter(wantingUsers, uint64(i.Member.User.ID)),
			char.ID,
		),
	)
}
