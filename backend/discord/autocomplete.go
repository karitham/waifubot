package discord

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
)

// extractAutocompleteInput extracts the user's input from an autocomplete option.
// Tries string first, falls back to int conversion if the option is numeric.
// Returns empty string if neither succeeds.
func extractAutocompleteInput(opts corde.OptionsInteractions, key string) string {
	if s, err := opts.String(key); err == nil {
		return s
	}
	if n, err := opts.Int(key); err == nil {
		return fmt.Sprintf("%d", n)
	}
	return ""
}

// autocomplete runs a generic autocomplete handler pipeline.
// It extracts input from the interaction, runs the search function,
// formats results into choices, and sends the response.
//
// On empty input: sends an empty autocomplete response.
// On search error: logs the error and returns without sending a response
// (Discord retains previous suggestions).
func autocomplete[T any](
	ctx context.Context,
	w corde.ResponseWriter,
	i *corde.Interaction[corde.AutocompleteInteractionData],
	optionKey string,
	searchFn func(ctx context.Context, input string) ([]T, error),
	formatFn func(T) (label string, value int64),
) {
	input := extractAutocompleteInput(i.Data.Options, optionKey)
	if input == "" {
		w.Autocomplete(corde.NewResp())
		return
	}

	results, err := searchFn(ctx, input)
	if err != nil {
		slog.Error("autocomplete search failed", "option", optionKey, "input", input, "error", err)
		return
	}

	resp := corde.NewResp()
	for _, r := range results {
		label, value := formatFn(r)
		resp.Choice(label, value)
	}

	w.Autocomplete(resp)
}

func formatCharacterChoice(c catalog.Character) (string, int64) {
	return fmt.Sprintf("%s (%d)", c.Name, c.ID), c.ID
}

func formatMediaChoice(m collection.Media) (string, int64) {
	return fmt.Sprintf("%s (%s)", m.Title, strings.ToTitle(strings.ToLower(m.Type))), m.ID
}
