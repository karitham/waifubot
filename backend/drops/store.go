// Package drops models channel drops: a character waits in a channel to be
// claimed, and channel activity makes drops appear.
package drops

import (
	"context"
	"errors"
	"strings"

	"github.com/karitham/waifubot/catalog"
)

// ChannelID identifies a channel that can hold a drop.
type ChannelID uint64

// DropWeightExponent controls how much favorites skew drop probability in the
// weighted-random formula. Higher values amplify the advantage of high-favorites
// (rarer) characters. 0 = uniform random, 1 = linear, 2 = quadratic.
const DropWeightExponent = 2.0

var (
	// ErrNoDropInChannel is returned when there is no drop in the channel.
	ErrNoDropInChannel = errors.New("no drop in this channel")
	// ErrWrongCharacterName is returned when the character name is wrong.
	ErrWrongCharacterName = errors.New("wrong character name")
)

// Drop is a character waiting to be claimed in a channel.
type Drop struct {
	Channel   ChannelID
	Character catalog.Character
}

// Matches reports whether the given name names the dropped character,
// ignoring case and whitespace differences.
func (d Drop) Matches(name string) bool {
	return strings.EqualFold(sanitizeName(d.Character.Name), sanitizeName(name))
}

// sanitizeName collapses whitespace in a character name to a single space.
func sanitizeName(name string) string {
	return strings.Join(strings.Fields(name), " ")
}

// Store is the pool-level drops surface: the activity counter, the drop
// write, and the pick a new drop is chosen from. It is implemented by
// storage/pg.
type Store interface {
	// IncrementActivity records an interaction on the channel and returns
	// the channel's new count.
	IncrementActivity(ctx context.Context, channel ChannelID) (int64, error)
	// ResetActivity clears the channel's activity count.
	ResetActivity(ctx context.Context, channel ChannelID) error
	// RandomActiveChar returns a random active character to drop, weighted
	// by favorites^weightExponent. Excludes the default AniList placeholder
	// image since drop images are embedded publicly.
	RandomActiveChar(ctx context.Context, weightExponent float64) (catalog.Character, error)
	// SetDrop stores the drop as the channel's current one, replacing any
	// previous drop. The character must already be in the catalog; the FK
	// enforces it.
	SetDrop(ctx context.Context, drop Drop) error
}

// TxWriter holds the drop writes that run inside the claim transaction.
type TxWriter interface {
	// GetDropForUpdate locks the channel's drop for the rest of the
	// transaction. It returns ErrNoDropInChannel when the channel has none.
	GetDropForUpdate(ctx context.Context, channel ChannelID) (Drop, error)
	// DeleteDrop removes the channel's drop.
	DeleteDrop(ctx context.Context, channel ChannelID) error
}
