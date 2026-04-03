package discord

import (
	"context"
	"fmt"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
)

// ListHandler handles the /list command.
type ListHandler struct {
	store collection.Store
}

// listOptions holds the parsed options for the list command.
type listOptions struct {
	targetUserID   uint64
	targetUsername string
	targetAvatar   string
}

func parseListOptions(cmd CommandContext) listOptions {
	if user, ok := cmd.FirstResolvedUser(); ok {
		return listOptions{
			targetUserID:   uint64(user.ID),
			targetUsername: user.Username,
			targetAvatar:   user.AvatarPNG(),
		}
	}
	return listOptions{}
}

// List displays a user's character collection.
func (h *ListHandler) List(ctx context.Context, w corde.ResponseWriter, cmd CommandContext) {
	opts := parseListOptions(cmd)

	userID := cmd.UserID()
	username := cmd.Username()
	avatar := cmd.AvatarPNG()
	if opts.targetUserID != 0 {
		userID = opts.targetUserID
		username = opts.targetUsername
		avatar = opts.targetAvatar
	}

	chars, err := collection.Characters(ctx, h.store, userID)
	if err != nil {
		w.Respond(rspErr("An error occurred dialing the database, please try again later"))
		return
	}

	if len(chars) == 0 {
		w.Respond(rspErr("No characters in collection"))
		return
	}

	embed := corde.NewEmbed().
		Titlef("%s's List", username).
		Thumbnail(corde.Image{URL: avatar}).
		URL(fmt.Sprintf("https://waifugui.karitham.dev/#/list/%d", userID))

	if len(chars) > 18 {
		chars = chars[:18]
	}

	for _, c := range chars {
		embed.FieldInline(c.Name, fmt.Sprintf("%d — %s", c.ID, c.Date.Format("02/01")))
	}

	w.Respond(corde.NewResp().Embeds(embed).Ephemeral())
}
