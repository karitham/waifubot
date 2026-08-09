package discord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Karitham/corde"
	"github.com/stretchr/testify/assert"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/discord/cordetest"
)

func TestHoldersHandler_Holders(t *testing.T) {
	tests := []struct {
		name           string
		cmd            CommandContext
		guildOps       *mockGuildQuerier
		catalog        *cordetest.MockCatalogStore
		wantContent    string
		wantNotContent string
	}{
		{
			name: "missing character ID",
			cmd: &MockCommandContext{
				GuildIDVal: 1,
				OptIntVals: map[string]int{},
				ErrVal:     errors.New("option not found"),
			},
			wantContent: "select a character",
		},
		{
			name: "no holders found",
			cmd: &MockCommandContext{
				GuildIDVal: 1,
				OptIntVals: map[string]int{"id": 42},
			},
			guildOps: &mockGuildQuerier{
				IsGuildIndexedFunc: func(ctx context.Context, guildID uint64) (collection.GuildIndexStatus, error) {
					return collection.GuildIndexStatus{Status: collection.IndexingCompleted}, nil
				},
			},
			catalog: &cordetest.MockCatalogStore{
				GetCharacterByIDFunc: func(ctx context.Context, charID int64) (catalog.Character, error) {
					return catalog.Character{ID: 42, Name: "Sakura"}, nil
				},
				GetCharacterHoldersInGuildFunc: func(ctx context.Context, guildID uint64, charID int64) ([]uint64, error) {
					return nil, nil
				},
			},
			wantContent: "No one in this server has",
		},
		{
			name: "holders found",
			cmd: &MockCommandContext{
				GuildIDVal: 1,
				OptIntVals: map[string]int{"id": 42},
			},
			guildOps: &mockGuildQuerier{
				IsGuildIndexedFunc: func(ctx context.Context, guildID uint64) (collection.GuildIndexStatus, error) {
					return collection.GuildIndexStatus{Status: collection.IndexingCompleted}, nil
				},
			},
			catalog: &cordetest.MockCatalogStore{
				GetCharacterByIDFunc: func(ctx context.Context, charID int64) (catalog.Character, error) {
					return catalog.Character{ID: 42, Name: "Sakura"}, nil
				},
				GetCharacterHoldersInGuildFunc: func(ctx context.Context, guildID uint64, charID int64) ([]uint64, error) {
					return []uint64{123, 456}, nil
				},
			},
			wantContent: "<@123>",
		},
		{
			name: "store error",
			cmd: &MockCommandContext{
				GuildIDVal: 1,
				OptIntVals: map[string]int{"id": 42},
			},
			guildOps: &mockGuildQuerier{
				IsGuildIndexedFunc: func(ctx context.Context, guildID uint64) (collection.GuildIndexStatus, error) {
					return collection.GuildIndexStatus{}, errors.New("db error")
				},
			},
			wantContent: "Failed to check holders",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &HoldersHandler{
				guildOps: tt.guildOps,
				catalog:  tt.catalog,
			}

			h.Holders(t.Context(), w, tt.cmd)

			assert.True(t, w.RespondCalled)
			if tt.wantContent != "" {
				w.AssertContains(t, tt.wantContent)
			}
			if tt.wantNotContent != "" {
				data := w.LastRespond.InteractionRespData()
				assert.NotContains(t, data.Content, tt.wantNotContent)
			}
		})
	}
}

func TestHoldersHandler_Autocomplete(t *testing.T) {
	tests := []struct {
		name             string
		input            corde.JsonRaw
		searchResult     []catalog.Character
		searchErr        error
		wantAutocomplete bool
		wantSearchCalled bool
		wantChoices      int
	}{
		{
			name:             "results found",
			input:            corde.JsonRaw(`"sakura"`),
			searchResult:     []catalog.Character{{ID: 1, Name: "Sakura"}, {ID: 2, Name: "Sakura Kinomoto"}},
			wantAutocomplete: true,
			wantSearchCalled: true,
			wantChoices:      2,
		},
		{
			name:             "no results",
			input:            corde.JsonRaw(`"xyz"`),
			searchResult:     nil,
			wantAutocomplete: true,
			wantSearchCalled: true,
			wantChoices:      0,
		},
		{
			name:             "store error",
			input:            corde.JsonRaw(`"test"`),
			searchErr:        errors.New("search failed"),
			wantAutocomplete: false,
			wantSearchCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var searchCalled bool
			mockStore := &cordetest.MockCatalogStore{
				SearchGlobalCharactersFunc: func(ctx context.Context, term string) ([]catalog.Character, error) {
					searchCalled = true
					return tt.searchResult, tt.searchErr
				},
			}

			w := &cordetest.MockResponseWriter{}
			i := cordetest.AutocompleteInteraction(1, 1, 1, "test", corde.OptionsInteractions{"id": tt.input})
			h := &HoldersHandler{catalog: mockStore}

			h.Autocomplete(t.Context(), w, i)

			assert.Equal(t, tt.wantSearchCalled, searchCalled)
			assert.Equal(t, tt.wantAutocomplete, w.AutocompleteCalled)

			if tt.wantAutocomplete {
				choices := w.Choices()
				assert.Len(t, choices, tt.wantChoices)
			}
		})
	}
}

// mockGuildQuerier implements guild.GuildQuerier for testing.
type mockGuildQuerier struct {
	IsGuildIndexedFunc func(ctx context.Context, guildID uint64) (collection.GuildIndexStatus, error)
}

func (m *mockGuildQuerier) IsGuildIndexed(ctx context.Context, guildID uint64) (collection.GuildIndexStatus, error) {
	return m.IsGuildIndexedFunc(ctx, guildID)
}

func (m *mockGuildQuerier) StartIndexingJob(ctx context.Context, guildID uint64) error {
	return nil
}

func (m *mockGuildQuerier) CompleteIndexingJob(ctx context.Context, guildID uint64) error {
	return nil
}

func (m *mockGuildQuerier) UpsertGuildMembers(ctx context.Context, guildID uint64, memberIDs []uint64, indexedAt time.Time) error {
	return nil
}

func (m *mockGuildQuerier) DeleteGuildMembersNotIn(ctx context.Context, guildID uint64, memberIDs []uint64) error {
	return nil
}
