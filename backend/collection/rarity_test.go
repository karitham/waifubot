package collection_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/karitham/waifubot/collection"
)

func TestRarityFromFavorites(t *testing.T) {
	tests := []struct {
		favorites int
		want      collection.RarityTier
	}{
		{0, collection.RarityCommon},
		{50, collection.RarityCommon},
		{99, collection.RarityCommon},
		{100, collection.RarityUncommon},
		{500, collection.RarityUncommon},
		{999, collection.RarityUncommon},
		{1000, collection.RarityRare},
		{3000, collection.RarityRare},
		{4999, collection.RarityRare},
		{5000, collection.RarityLegendary},
		{10000, collection.RarityLegendary},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := collection.RarityFromFavorites(tt.favorites)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRarityTierString(t *testing.T) {
	tests := []struct {
		tier collection.RarityTier
		want string
	}{
		{collection.RarityCommon, "Common"},
		{collection.RarityUncommon, "Uncommon"},
		{collection.RarityRare, "Rare"},
		{collection.RarityLegendary, "Legendary"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.tier.String())
		})
	}
}

func TestRarityTierColor(t *testing.T) {
	tests := []struct {
		tier collection.RarityTier
		want uint32
	}{
		{collection.RarityCommon, 0x99AAB5},
		{collection.RarityUncommon, 0x57F287},
		{collection.RarityRare, 0x5865F2},
		{collection.RarityLegendary, 0xFEE75C},
	}

	for _, tt := range tests {
		t.Run(tt.tier.String(), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.tier.Color())
		})
	}
}
