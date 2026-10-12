package discord

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGradientHex(t *testing.T) {
	tests := []struct {
		name    string
		val     int
		wantHex string
	}{
		{"zero", 0, "#969696"},
		{"one", 1, "#969696"},
		{"uncommon stop", 100, "#2ECC71"},
		{"rare stop", 1000, "#3498DB"},
		{"legendary stop", 5000, "#F1C40F"},
		{"max stop", 15000, "#E67E22"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gradientHex(tt.val)
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
