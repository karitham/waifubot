package discord

import (
	"context"
	"fmt"

	"github.com/Karitham/corde"
)

func (b *Bot) list(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	user := i.Member.User
	if len(i.Data.Resolved.Members) > 0 {
		user = i.Data.Resolved.Users.First()
	}

	chars, err := b.Store.Chars(ctx, user.ID)
	if err != nil {
		w.Respond(rspErr("An error occurred dialing the database, please try again later"))
		return
	}

	if len(chars) == 0 {
		w.Respond(rspErr("No characters in collection"))
		return
	}

	embed := corde.NewEmbed().
		Titlef("%s's List", user.Username).
		Thumbnail(corde.Image{URL: user.AvatarPNG()}).
		URL(fmt.Sprintf("https://waifugui.karitham.dev/#/list/%s", user.ID.String()))

	if len(chars) > 18 {
		chars = chars[:18]
	}

	for _, c := range chars {
		embed.FieldInline(c.Name, fmt.Sprintf("%d — %s", c.ID, c.Date.Format("02/01")))
	}

	w.Respond(corde.NewResp().Embeds(embed).Ephemeral())
}
