package collection

import (
	"fmt"
	"math"
	"strconv"
)

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

// GradientColor returns the Discord embed color using gradient interpolation based on favorites count.
func GradientColor(favorites int) uint32 {
	hex := getRarityHex(favorites)
	// Parse "#RRGGBB" to uint32 (stored as 0xRRGGBB)
	parsed, err := strconv.ParseInt(hex[1:], 16, 32)
	if err != nil {
		return 0x99AAB5 // fallback to gray
	}
	return uint32(parsed)
}

// Rarity returns the rarity tier for this character.
func (c MediaCharacter) Rarity() RarityTier {
	return RarityFromFavorites(c.Favorites)
}

// RGB represents a color with red, green, and blue components.
type RGB struct {
	R, G, B float64
}

// Threshold represents a value threshold with an associated color.
type Threshold struct {
	Val   float64
	Color RGB
}

// thresholdColors defines the color gradient in log space.
// Values: 1 (Gray), 100 (Green), 1000 (Blue), 5000 (Gold), 15000 (Orange)
var thresholdColors = []Threshold{
	{1, RGB{150, 150, 150}},
	{100, RGB{46, 204, 113}},
	{1000, RGB{52, 152, 219}},
	{5000, RGB{241, 196, 15}},
	{15000, RGB{230, 126, 34}},
}

// rgbToHex converts an RGB color to a hex string "#RRGGBB".
func rgbToHex(c RGB) string {
	r := int(c.R + 0.5)
	g := int(c.G + 0.5)
	b := int(c.B + 0.5)
	if r > 255 {
		r = 255
	}
	if g > 255 {
		g = 255
	}
	if b > 255 {
		b = 255
	}
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

// rgbLerp performs linear interpolation between two RGB colors.
func rgbLerp(c1, c2 RGB, t float64) RGB {
	return RGB{
		R: c1.R + (c2.R-c1.R)*t,
		G: c1.G + (c2.G-c1.G)*t,
		B: c1.B + (c2.B-c1.B)*t,
	}
}

// getRarityHex returns the hex color for a given favorites count using log-space interpolation.
func getRarityHex(favorites int) string {
	if favorites <= 0 {
		return rgbToHex(thresholdColors[0].Color)
	}

	val := float64(favorites)

	// Find the two thresholds to interpolate between
	var i int
	for i = len(thresholdColors) - 1; i > 0; i-- {
		if val >= thresholdColors[i].Val {
			break
		}
	}

	// Handle case where val exceeds max threshold
	if i >= len(thresholdColors)-1 {
		return rgbToHex(thresholdColors[len(thresholdColors)-1].Color)
	}

	// Log-space interpolation
	logVal := math.Log(val)
	logLower := math.Log(thresholdColors[i].Val)
	logUpper := math.Log(thresholdColors[i+1].Val)

	// Avoid division by zero
	if logUpper == logLower {
		return rgbToHex(thresholdColors[i].Color)
	}

	t := (logVal - logLower) / (logUpper - logLower)
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	interpolated := rgbLerp(thresholdColors[i].Color, thresholdColors[i+1].Color, t)
	return rgbToHex(interpolated)
}
