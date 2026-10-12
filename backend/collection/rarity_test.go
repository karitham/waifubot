package collection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRarityFromPopularity(t *testing.T) {
	tests := []struct {
		favorites int
		want      RarityTier
	}{
		{0, RarityCommon},
		{50, RarityCommon},
		{99, RarityCommon},
		{100, RarityUncommon},
		{500, RarityUncommon},
		{999, RarityUncommon},
		{1000, RarityRare},
		{3000, RarityRare},
		{4999, RarityRare},
		{5000, RarityLegendary},
		{10000, RarityLegendary},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := RarityFromPopularity(tt.favorites)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRarityTierString(t *testing.T) {
	tests := []struct {
		tier RarityTier
		want string
	}{
		{RarityCommon, "Common"},
		{RarityUncommon, "Uncommon"},
		{RarityRare, "Rare"},
		{RarityLegendary, "Legendary"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.tier.String())
		})
	}
}
