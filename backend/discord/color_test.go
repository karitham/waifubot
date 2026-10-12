package discord

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGradientHex(t *testing.T) {
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
