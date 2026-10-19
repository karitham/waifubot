package discord

import (
	"context"
	"testing"

	"github.com/Karitham/corde"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/discord/cordetest"
)

func TestSlashCommandCtxPrefersGuildMember(t *testing.T) {
	i := &corde.Interaction[corde.SlashCommandInteractionData]{
		Member: corde.Member{User: corde.User{ID: 7, Username: "guild-user", Avatar: "guild-avatar"}},
		User:   &corde.User{ID: 9, Username: "dm-user"},
	}

	cmd := newSlashCommandCtx(i)

	assert.Equal(t, collection.UserID(7), cmd.UserID())
	assert.Equal(t, "guild-user", cmd.Username())
	assert.Contains(t, cmd.AvatarPNG(), "guild-avatar")
}

func TestSlashCommandCtxFallsBackToDMUser(t *testing.T) {
	i := &corde.Interaction[corde.SlashCommandInteractionData]{
		User: &corde.User{ID: 42, Username: "dm-user", Avatar: "dm-avatar"},
	}

	cmd := newSlashCommandCtx(i)

	assert.Equal(t, collection.UserID(42), cmd.UserID())
	assert.Equal(t, "dm-user", cmd.Username())
	assert.Contains(t, cmd.AvatarPNG(), "dm-avatar")
}

func TestWrapCtxRejectsMissingCaller(t *testing.T) {
	called := false
	handler := wrapCtx(func(context.Context, corde.ResponseWriter, CommandContext) {
		called = true
	})

	w := &cordetest.MockResponseWriter{}
	handler(t.Context(), w, &corde.Interaction[corde.SlashCommandInteractionData]{})

	assert.False(t, called, "handler must not run without a caller identity")
	assert.True(t, w.RespondCalled)
}

func TestCallerUserIDForAutocompleteFallsBackToDMUser(t *testing.T) {
	id := callerUserID(corde.Member{}, &corde.User{ID: 42})

	require.Equal(t, collection.UserID(42), id)
}
