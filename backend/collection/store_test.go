package collection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertIndexingStatus(t *testing.T) {
	tests := []struct {
		input string
		want  IndexingStatus
	}{
		{"completed", IndexingCompleted},
		{"in_progress", IndexingInProgress},
		{"pending", IndexingPending},
		{"", IndexingPending},
		{"unknown_status", IndexingPending},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ConvertIndexingStatus(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no change needed", "Naruto Uzumaki", "Naruto Uzumaki"},
		{"collapse spaces", "Naruto  Uzumaki", "Naruto Uzumaki"},
		{"collapse tabs", "Naruto\tUzumaki", "Naruto Uzumaki"},
		{"leading and trailing", "  Naruto  ", "Naruto"},
		{"multiple whitespace types", " \t Naruto \t Uzumaki \n ", "Naruto Uzumaki"},
		{"empty string", "", ""},
		{"only whitespace", "  \t \n  ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeName(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
