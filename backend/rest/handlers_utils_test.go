package rest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEncodeDecodePageToken(t *testing.T) {
	t.Run("encodes and decodes page token correctly", func(t *testing.T) {
		offset := 20
		encoded := encodePageToken(offset)
		assert.NotEmpty(t, encoded)

		decoded, err := decodePageToken(encoded)
		assert.NoError(t, err)
		assert.Equal(t, offset, decoded)
	})

	t.Run("returns error for invalid token", func(t *testing.T) {
		_, err := decodePageToken("invalid-token")
		assert.Error(t, err)
	})
}

func TestEncodeDecodeCollectionPageToken(t *testing.T) {
	t.Run("encodes and decodes token correctly", func(t *testing.T) {
		token := collectionPageToken{
			LastDate: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			LastID:   42,
			LastName: "Rem",
			OrderBy:  "date",
			Search:   "test",
		}
		encoded := encodeCollectionPageToken(token)
		assert.NotEmpty(t, encoded)

		decoded, err := decodeCollectionPageToken(encoded)
		assert.NoError(t, err)
		assert.Equal(t, token.LastID, decoded.LastID)
		assert.Equal(t, token.LastName, decoded.LastName)
	})

	t.Run("returns error for invalid token", func(t *testing.T) {
		_, err := decodeCollectionPageToken("invalid")
		assert.Error(t, err)
	})
}

func TestNormalizeAnilistURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"username", "https://anilist.co/user/username"},
		{"https://anilist.co/user/username", "https://anilist.co/user/username"},
		{"http://anilist.co/user/username", "http://anilist.co/user/username"},
		{"  username  ", "https://anilist.co/user/username"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeAnilistURL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
