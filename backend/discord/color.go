package discord

import (
	"fmt"
	"math"
	"strconv"
)

// gradientColor returns the Discord embed color for a popularity count using
// log-space interpolation. The gradient is deliberately continuous where the
// rarity tiers are discrete, so it keeps shifting past the legendary tier.
func gradientColor(popularity int) uint32 {
	hex := gradientHex(popularity)
	// Parse "#RRGGBB" to uint32 (stored as 0xRRGGBB).
	parsed, err := strconv.ParseInt(hex[1:], 16, 32)
	if err != nil {
		return 0x99AAB5 // fallback to gray
	}
	return uint32(parsed)
}

// rgb represents a color with red, green, and blue components.
type rgb struct {
	r, g, b float64
}

// gradientStops defines the color gradient in log space.
// Values: 1 (Gray), 100 (Green), 1000 (Blue), 5000 (Gold), 15000 (Orange)
var gradientStops = []struct {
	val   float64
	color rgb
}{
	{1, rgb{150, 150, 150}},
	{100, rgb{46, 204, 113}},
	{1000, rgb{52, 152, 219}},
	{5000, rgb{241, 196, 15}},
	{15000, rgb{230, 126, 34}},
}

// rgbToHex converts an rgb color to a hex string "#RRGGBB".
func rgbToHex(c rgb) string {
	r := min(int(c.r+0.5), 255)
	g := min(int(c.g+0.5), 255)
	b := min(int(c.b+0.5), 255)
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

// rgbLerp performs linear interpolation between two rgb colors.
func rgbLerp(c1, c2 rgb, t float64) rgb {
	return rgb{
		r: c1.r + (c2.r-c1.r)*t,
		g: c1.g + (c2.g-c1.g)*t,
		b: c1.b + (c2.b-c1.b)*t,
	}
}

// gradientHex returns the hex color for a given popularity count using
// log-space interpolation.
func gradientHex(popularity int) string {
	if popularity <= 0 {
		return rgbToHex(gradientStops[0].color)
	}

	val := float64(popularity)

	// Find the two stops to interpolate between.
	var i int
	for i = len(gradientStops) - 1; i > 0; i-- {
		if val >= gradientStops[i].val {
			break
		}
	}

	// Handle case where val exceeds max stop.
	if i >= len(gradientStops)-1 {
		return rgbToHex(gradientStops[len(gradientStops)-1].color)
	}

	// Log-space interpolation.
	logVal := math.Log(val)
	logLower := math.Log(gradientStops[i].val)
	logUpper := math.Log(gradientStops[i+1].val)

	// Avoid division by zero.
	if logUpper == logLower {
		return rgbToHex(gradientStops[i].color)
	}

	t := (logVal - logLower) / (logUpper - logLower)
	t = min(max(t, 0), 1)

	interpolated := rgbLerp(gradientStops[i].color, gradientStops[i+1].color, t)
	return rgbToHex(interpolated)
}
