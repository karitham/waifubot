package collection

// RarityTier represents the rarity classification of a character.
type RarityTier int

const (
	RarityCommon    RarityTier = iota // Gray
	RarityUncommon                    // Green
	RarityRare                        // Blue
	RarityLegendary                   // Gold
)

// String returns the display name of the rarity tier.
func (r RarityTier) String() string {
	switch r {
	case RarityCommon:
		return "Common"
	case RarityUncommon:
		return "Uncommon"
	case RarityRare:
		return "Rare"
	case RarityLegendary:
		return "Legendary"
	default:
		return "Unknown"
	}
}

// Color returns the Discord embed color for this tier.
func (r RarityTier) Color() uint32 {
	switch r {
	case RarityCommon:
		return 0x99AAB5
	case RarityUncommon:
		return 0x57F287
	case RarityRare:
		return 0x5865F2
	case RarityLegendary:
		return 0xFEE75C
	default:
		return 0x99AAB5
	}
}

// RarityFromFavorites classifies a favorites count into a tier.
func RarityFromFavorites(favorites int) RarityTier {
	switch {
	case favorites >= 5000:
		return RarityLegendary
	case favorites >= 1000:
		return RarityRare
	case favorites >= 100:
		return RarityUncommon
	default:
		return RarityCommon
	}
}

// Rarity returns the rarity tier for this character.
func (c MediaCharacter) Rarity() RarityTier {
	return RarityFromFavorites(c.Favorites)
}
