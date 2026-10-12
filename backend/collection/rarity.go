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

// rarityThresholds maps popularity thresholds to tiers, highest first.
var rarityThresholds = []struct {
	min  int
	tier RarityTier
}{
	{5000, RarityLegendary},
	{1000, RarityRare},
	{100, RarityUncommon},
}

// RarityFromPopularity classifies a popularity count into a tier.
func RarityFromPopularity(popularity int) RarityTier {
	for _, t := range rarityThresholds {
		if popularity >= t.min {
			return t.tier
		}
	}
	return RarityCommon
}
