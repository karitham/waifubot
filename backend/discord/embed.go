package discord

import (
	"fmt"
	"io"
	"strings"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
)

func rollEmbed(char collection.MediaCharacter, usersWanting string) corde.Embed {
	return newCharEmbed(char, fmt.Sprintf(
		"You got %s (%s)\n⭐ Rarity: %s | ❤️ %d favorites%s\n\n🎲 Standard Roll\nID: %d",
		char.Name,
		char.MediaTitle,
		char.Rarity(),
		char.Favorites,
		usersWanting,
		char.ID,
	))
}

func seriesRollEmbed(char collection.MediaCharacter, cost int32) corde.Embed {
	return newCharEmbed(char, fmt.Sprintf(
		"You got %s (%s)\n⭐ Rarity: %s | ❤️ %d favorites\n\n🎯 Series Roll | Cost: %d tokens\nID: %d",
		char.Name,
		char.MediaTitle,
		char.Rarity(),
		char.Favorites,
		cost,
		char.ID,
	))
}

// dropMessage creates a drop message with the character embed and optional image attachment.
// If image is nil, the message is built without an image attachment.
func dropMessage(char collection.MediaCharacter, image io.Reader) corde.Message {
	parts := strings.Fields(char.Name)
	initials := strings.Builder{}
	for i, part := range parts {
		if i != 0 {
			initials.WriteRune('.')
		}

		initials.WriteRune([]rune(part)[0])
	}

	msg := corde.Message{
		Embeds: []corde.Embed{{
			Title: "Character Drop!",
			Description: fmt.Sprintf(
				"A character has appeared! Use `/claim <name>` to add them to your collection.\n\n**Hint:** %s\n\n⭐ Rarity: %s",
				initials.String(),
				char.Rarity(),
			),
			Footer: corde.Footer{IconURL: AnilistIconURL, Text: "View on Anilist"},
			Color:  collection.GradientColor(char.Favorites),
		}},
	}

	if image != nil {
		msg.Embeds[0].Image = corde.Image{URL: "attachment://image.png"}
		msg.Attachments = []corde.Attachment{{
			Filename: "image.png",
			Body:     image,
		}}
	}

	return msg
}

func claimEmbed(char collection.Character, rarity string) corde.Embed {
	return corde.NewEmbed().
		Title(char.Name).
		URL(fmt.Sprintf("https://anilist.co/character/%d", char.ID)).
		Color(collection.GradientColor(char.Favorites)).
		Footer(corde.Footer{IconURL: AnilistIconURL, Text: "View on Anilist"}).
		Thumbnail(corde.Image{URL: char.Image}).
		Descriptionf(
			"You got %s (%s)\n⭐ Rarity: %s | ❤️ %d favorites\nID: %d",
			char.Name,
			char.MediaTitle,
			rarity,
			char.Favorites,
			char.ID,
		).
		Embed()
}

func newCharEmbed(char collection.MediaCharacter, description string) corde.Embed {
	return corde.NewEmbed().
		Title(char.Name).
		URL(fmt.Sprintf("https://anilist.co/character/%d", char.ID)).
		Color(collection.GradientColor(char.Favorites)).
		Footer(corde.Footer{IconURL: AnilistIconURL, Text: "View on Anilist"}).
		Thumbnail(corde.Image{URL: char.ImageURL}).
		Description(description).
		Embed()
}
