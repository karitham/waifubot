package discord

import (
	"context"
	"errors"
	"testing"

	"github.com/Karitham/corde"
	"github.com/stretchr/testify/assert"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/discord/cordetest"
)

func TestExtractAutocompleteInput(t *testing.T) {
	tests := []struct {
		name     string
		opts     corde.OptionsInteractions
		key      string
		expected string
	}{
		{
			name:     "string value present",
			opts:     corde.OptionsInteractions{"name": corde.JsonRaw(`"sakura"`)},
			key:      "name",
			expected: "sakura",
		},
		{
			name:     "int value present (no string)",
			opts:     corde.OptionsInteractions{"id": corde.JsonRaw(`12345`)},
			key:      "id",
			expected: "12345",
		},
		{
			name:     "key not found",
			opts:     corde.OptionsInteractions{"other": corde.JsonRaw(`"value"`)},
			key:      "name",
			expected: "",
		},
		{
			name:     "empty options map",
			opts:     corde.OptionsInteractions{},
			key:      "name",
			expected: "",
		},
		{
			name:     "string takes priority over int for same key",
			opts:     corde.OptionsInteractions{"id": corde.JsonRaw(`"hello"`)},
			key:      "id",
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractAutocompleteInput(tt.opts, tt.key)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestAutocomplete(t *testing.T) {
	tests := []struct {
		name             string
		input            corde.JsonRaw
		searchResult     []string
		searchErr        error
		wantAutocomplete bool
		wantSearchCalled bool
		wantChoices      int
		checkChoices     func(t *testing.T, choices []corde.Choice[any])
		wantSearchInput  string
	}{
		{
			name:             "empty input sends empty response",
			input:            corde.JsonRaw(`""`),
			wantAutocomplete: true,
			wantSearchCalled: false,
			wantChoices:      0,
		},
		{
			name:             "search error returns without responding",
			input:            corde.JsonRaw(`"test"`),
			searchErr:        errors.New("search failed"),
			wantAutocomplete: false,
			wantSearchCalled: true,
		},
		{
			name:             "results formatted into choices",
			input:            corde.JsonRaw(`"abc"`),
			searchResult:     []string{"alpha", "bravo", "charlie"},
			wantAutocomplete: true,
			wantSearchCalled: true,
			wantChoices:      3,
			checkChoices: func(t *testing.T, choices []corde.Choice[any]) {
				t.Helper()
				assert.Equal(t, "alpha", choices[0].Name)
				assert.Equal(t, int64('a'), choices[0].Value)
				assert.Equal(t, "bravo", choices[1].Name)
				assert.Equal(t, int64('b'), choices[1].Value)
				assert.Equal(t, "charlie", choices[2].Name)
				assert.Equal(t, int64('c'), choices[2].Value)
			},
		},
		{
			name:             "nil results sends empty choices",
			input:            corde.JsonRaw(`"xyz"`),
			searchResult:     nil,
			wantAutocomplete: true,
			wantSearchCalled: true,
			wantChoices:      0,
		},
		{
			name:             "empty slice sends empty choices",
			input:            corde.JsonRaw(`"query"`),
			searchResult:     []string{},
			wantAutocomplete: true,
			wantSearchCalled: true,
			wantChoices:      0,
		},
		{
			name:             "int fallback passes stringified int",
			input:            corde.JsonRaw(`42`),
			searchResult:     []string{"result"},
			wantAutocomplete: true,
			wantSearchCalled: true,
			wantChoices:      1,
			wantSearchInput:  "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var searchCalled bool
			var gotInput string
			searchFn := func(ctx context.Context, input string) ([]string, error) {
				searchCalled = true
				gotInput = input
				return tt.searchResult, tt.searchErr
			}
			formatFn := func(s string) (string, int64) {
				return s, int64(s[0])
			}

			w := &cordetest.MockResponseWriter{}
			i := cordetest.AutocompleteInteraction(1, 1, 1, "test", corde.OptionsInteractions{"id": tt.input})

			autocomplete(t.Context(), w, i, "id", searchFn, formatFn)

			assert.Equal(t, tt.wantAutocomplete, w.AutocompleteCalled)
			assert.Equal(t, tt.wantSearchCalled, searchCalled)

			if tt.wantAutocomplete {
				choices := w.Choices()
				assert.Len(t, choices, tt.wantChoices)
				if tt.checkChoices != nil {
					tt.checkChoices(t, choices)
				}
			}
			if tt.wantSearchInput != "" {
				assert.Equal(t, tt.wantSearchInput, gotInput)
			}
		})
	}
}

func TestFormatCharacterChoice(t *testing.T) {
	tests := []struct {
		name      string
		char      catalog.Character
		wantLabel string
		wantValue int64
	}{
		{
			name:      "normal character",
			char:      catalog.Character{ID: 42, Name: "Sakura"},
			wantLabel: "Sakura (42)",
			wantValue: 42,
		},
		{
			name:      "zero ID",
			char:      catalog.Character{ID: 0, Name: "Unknown"},
			wantLabel: "Unknown (0)",
			wantValue: 0,
		},
		{
			name:      "large ID",
			char:      catalog.Character{ID: 9223372036854775807, Name: "MaxID"},
			wantLabel: "MaxID (9223372036854775807)",
			wantValue: 9223372036854775807,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label, value := formatCharacterChoice(tt.char)
			assert.Equal(t, tt.wantLabel, label)
			assert.Equal(t, tt.wantValue, value)
		})
	}
}

func TestFormatMediaChoice(t *testing.T) {
	tests := []struct {
		name      string
		media     collection.Media
		wantLabel string
		wantValue int64
	}{
		{
			name:      "ANIME type uppercase preserved",
			media:     collection.Media{ID: 1, Title: "Naruto", Type: "ANIME"},
			wantLabel: "Naruto (ANIME)",
			wantValue: 1,
		},
		{
			name:      "manga type lowercased then uppercased",
			media:     collection.Media{ID: 2, Title: "One Piece", Type: "manga"},
			wantLabel: "One Piece (MANGA)",
			wantValue: 2,
		},
		{
			name:      "mixed case properly uppercased",
			media:     collection.Media{ID: 3, Title: "Bleach", Type: "AnImE"},
			wantLabel: "Bleach (ANIME)",
			wantValue: 3,
		},
		{
			name:      "OVA type uppercased",
			media:     collection.Media{ID: 4, Title: "Hellsing Ultimate", Type: "ova"},
			wantLabel: "Hellsing Ultimate (OVA)",
			wantValue: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label, value := formatMediaChoice(tt.media)
			assert.Equal(t, tt.wantLabel, label)
			assert.Equal(t, tt.wantValue, value)
		})
	}
}
