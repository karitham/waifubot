package discord

import (
	"context"

	"github.com/Karitham/corde"
)

// CommandContext provides the interaction data needed by command handlers.
type CommandContext interface {
	UserID() uint64
	Username() string
	GuildID() uint64
	ChannelID() uint64
	AvatarPNG() string
	FirstResolvedUser() (corde.User, bool)
	OptString(key string) (string, error)
	OptInt(key string) (int, error)
	OptInt64(key string) (int64, error)
	OptUser(key string) (corde.User, error)
}

type slashCommandCtx struct {
	i *corde.Interaction[corde.SlashCommandInteractionData]
}

func newSlashCommandCtx(i *corde.Interaction[corde.SlashCommandInteractionData]) *slashCommandCtx {
	return &slashCommandCtx{i: i}
}

func (c *slashCommandCtx) UserID() uint64    { return uint64(c.i.Member.User.ID) }
func (c *slashCommandCtx) Username() string  { return c.i.Member.User.Username }
func (c *slashCommandCtx) GuildID() uint64   { return uint64(c.i.GuildID) }
func (c *slashCommandCtx) ChannelID() uint64 { return uint64(c.i.ChannelID) }
func (c *slashCommandCtx) AvatarPNG() string { return c.i.Member.User.AvatarPNG() }
func (c *slashCommandCtx) FirstResolvedUser() (corde.User, bool) {
	if len(c.i.Data.Resolved.Users) == 0 {
		return corde.User{}, false
	}
	return c.i.Data.Resolved.Users.First(), true
}
func (c *slashCommandCtx) OptString(key string) (string, error)   { return c.i.Data.Options.String(key) }
func (c *slashCommandCtx) OptInt(key string) (int, error)         { return c.i.Data.Options.Int(key) }
func (c *slashCommandCtx) OptInt64(key string) (int64, error)     { return c.i.Data.Options.Int64(key) }
func (c *slashCommandCtx) OptUser(key string) (corde.User, error) { return c.i.Data.OptionsUser(key) }

// wrapCtx adapts a CommandContext-accepting handler into the middleware chain's signature.
func wrapCtx(
	handler func(ctx context.Context, w corde.ResponseWriter, cmd CommandContext),
) func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
	return func(ctx context.Context, w corde.ResponseWriter, i *corde.Interaction[corde.SlashCommandInteractionData]) {
		handler(ctx, w, newSlashCommandCtx(i))
	}
}

// MockCommandContext is a test double for CommandContext.
type MockCommandContext struct {
	UserIDVal            uint64
	UsernameVal          string
	GuildIDVal           uint64
	ChannelIDVal         uint64
	AvatarPNGVal         string
	FirstResolvedUserVal corde.User
	HasResolvedUser      bool
	OptStringVals        map[string]string
	OptIntVals           map[string]int
	OptInt64Vals         map[string]int64
	OptUserVals          map[string]corde.User
	ErrVal               error
}

func (m *MockCommandContext) UserID() uint64    { return m.UserIDVal }
func (m *MockCommandContext) Username() string  { return m.UsernameVal }
func (m *MockCommandContext) GuildID() uint64   { return m.GuildIDVal }
func (m *MockCommandContext) ChannelID() uint64 { return m.ChannelIDVal }
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
