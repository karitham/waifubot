package discord

import (
	"github.com/Karitham/corde"
)

// Privf creates a private (ephemeral) formatted response
func Privf(format string, args ...any) *corde.RespB {
	return corde.NewResp().Contentf(format, args...).Ephemeral()
}
