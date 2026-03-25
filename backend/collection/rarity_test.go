package collection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRarityFromFavorites(t *testing.T) {
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
			got := RarityFromFavorites(tt.favorites)
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

func TestRarityTierColor(t *testing.T) {
	tests := []struct {
		tier RarityTier
		want uint32
	}{
		{RarityCommon, 0x969696},
		{RarityUncommon, 0x2ECC71},
		{RarityRare, 0x3498DB},
		{RarityLegendary, 0xF1C40F},
	}

	for _, tt := range tests {
		t.Run(tt.tier.String(), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.tier.Color())
		})
	}
}

func TestGetRarityHex(t *testing.T) {
	tests := []struct {
		val     int
		wantHex string
	}{
		{0, "#969696"},
		{1, "#969696"},
		{100, "#2ECC71"},
		{1000, "#3498DB"},
		{5000, "#F1C40F"},
		{15000, "#E67E22"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := getRarityHex(tt.val)
			assert.Equal(t, tt.wantHex, got)

			// Verify format is "#RRGGBB"
			assert.Len(t, got, 7)
			assert.Equal(t, "#", got[:1])
			for _, c := range got[1:] {
				assert.True(t, (c >= '0' && c <= '9') || (c >= 'A' && c <= 'F'), "got=%s", got)
			}
		})
	}
}
