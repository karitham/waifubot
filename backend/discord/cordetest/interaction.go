package cordetest

import (
	"github.com/Karitham/corde"
)

// AutocompleteInteraction creates an autocomplete interaction with the given options.
// It sets reasonable defaults for Member, GuildID, ChannelID, and other fields.
func AutocompleteInteraction(
	userID uint64,
	guildID uint64,
	channelID uint64,
	username string,
	opts corde.OptionsInteractions,
) *corde.Interaction[corde.AutocompleteInteractionData] {
	return &corde.Interaction[corde.AutocompleteInteractionData]{
		ID:            corde.Snowflake(userID),
		ApplicationID: 1,
		Type:          corde.INTERACTION_TYPE_APPLICATION_COMMAND_AUTOCOMPLETE,
		GuildID:       corde.Snowflake(guildID),
		ChannelID:     corde.Snowflake(channelID),
		Member: corde.Member{
			User: corde.User{
				ID:       corde.Snowflake(userID),
				Username: username,
			},
		},
		InnerInteractionType: corde.AutocompleteInteraction,
		Data: corde.AutocompleteInteractionData{
			ID:      1,
			Name:    "test",
			Type:    1,
			Options: opts,
		},
	}
}
