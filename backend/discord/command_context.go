package discord

import (
	"context"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/drops"
	"github.com/karitham/waifubot/guild"
)

// CommandContext provides the interaction data needed by command handlers.
type CommandContext interface {
	UserID() collection.UserID
	Username() string
	GuildID() guild.ID
	ChannelID() drops.ChannelID
	AvatarPNG() string
	FirstResolvedUser() (corde.User, bool)
	OptString(key string) (string, error)
	OptInt(key string) (int, error)
	OptInt64(key string) (int64, error)
	OptUser(key string) (corde.User, error)
}

type slashCommandCtx struct {
	i    *corde.Interaction[corde.SlashCommandInteractionData]
	user corde.User
}

func newSlashCommandCtx(i *corde.Interaction[corde.SlashCommandInteractionData]) *slashCommandCtx {
	return &slashCommandCtx{i: i, user: callerUser(i.Member, i.User)}
}

// callerUser resolves the invoking user. Discord sends member in guilds and
// user in DMs; a DM has no member, so reading Member.User alone would collapse
// every DM caller to user 0.
func callerUser(member corde.Member, user *corde.User) corde.User {
	if member.User.ID != 0 {
		return member.User
	}
	if user != nil {
		return *user
	}
	return corde.User{}
}

// callerUserID is the caller identity for interaction types that do not use
// CommandContext, such as autocomplete.
func callerUserID(member corde.Member, user *corde.User) collection.UserID {
	return collection.UserID(callerUser(member, user).ID)
}

func (c *slashCommandCtx) UserID() collection.UserID { return collection.UserID(c.user.ID) }
func (c *slashCommandCtx) Username() string          { return c.user.Username }
func (c *slashCommandCtx) GuildID() guild.ID         { return guild.ID(c.i.GuildID) }
func (c *slashCommandCtx) ChannelID() drops.ChannelID {
	return drops.ChannelID(c.i.ChannelID)
}
func (c *slashCommandCtx) AvatarPNG() string { return c.user.AvatarPNG() }
func (c *slashCommandCtx) FirstResolvedUser() (corde.User, bool) {
	if len(c.i.Data.Resolved.Users) == 0 {
		return corde.User{}, false
	}
	return c.i.Data.Resolved.Users.First(), true
}

func (c *slashCommandCtx) OptString(key string) (string, error) { return c.i.Data.Options.String(key) }

func (c *slashCommandCtx) OptInt(key string) (int, error) { return c.i.Data.Options.Int(key) }

func (c *slashCommandCtx) OptInt64(key string) (int64, error) { return c.i.Data.Options.Int64(key) }

func (c *slashCommandCtx) OptUser(key string) (corde.User, error) { return c.i.Data.OptionsUser(key) }

// wrapCtx adapts a CommandContext-accepting handler into the middleware chain's signature.
func wrapCtx(
	handler func(ctx context.Context, w corde.ResponseWriter, cmd CommandContext),
) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	return func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
		cmd := newSlashCommandCtx(i)
		if !cmd.UserID().Valid() {
			w.Respond(Privf("Could not identify your Discord user."))
			return
		}
		handler(ctx, w, cmd)
	}
}

// MockCommandContext is a test double for CommandContext.
type MockCommandContext struct {
	UserIDVal            collection.UserID
	UsernameVal          string
	GuildIDVal           guild.ID
	ChannelIDVal         drops.ChannelID
	AvatarPNGVal         string
	FirstResolvedUserVal corde.User
	HasResolvedUser      bool
	OptStringVals        map[string]string
	OptIntVals           map[string]int
	OptInt64Vals         map[string]int64
	OptUserVals          map[string]corde.User
	ErrVal               error
}

func (m *MockCommandContext) UserID() collection.UserID { return m.UserIDVal }
func (m *MockCommandContext) Username() string          { return m.UsernameVal }
func (m *MockCommandContext) GuildID() guild.ID         { return m.GuildIDVal }
func (m *MockCommandContext) ChannelID() drops.ChannelID {
	return m.ChannelIDVal
}
func (m *MockCommandContext) AvatarPNG() string { return m.AvatarPNGVal }
func (m *MockCommandContext) FirstResolvedUser() (corde.User, bool) {
	return m.FirstResolvedUserVal, m.HasResolvedUser
}

func (m *MockCommandContext) OptString(key string) (string, error) {
	v, ok := m.OptStringVals[key]
	if !ok {
		return "", m.ErrVal
	}
	return v, nil
}

func (m *MockCommandContext) OptInt(key string) (int, error) {
	v, ok := m.OptIntVals[key]
	if !ok {
		return 0, m.ErrVal
	}
	return v, nil
}

func (m *MockCommandContext) OptInt64(key string) (int64, error) {
	v, ok := m.OptInt64Vals[key]
	if !ok {
		return 0, m.ErrVal
	}
	return v, nil
}

func (m *MockCommandContext) OptUser(key string) (corde.User, error) {
	v, ok := m.OptUserVals[key]
	if !ok {
		return corde.User{}, m.ErrVal
	}
	return v, nil
}

var _ CommandContext = (*MockCommandContext)(nil)
