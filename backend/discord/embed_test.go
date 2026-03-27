package discord

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/karitham/waifubot/collection"
)

func TestRollEmbed(t *testing.T) {
	for _, tt := range []struct {
		name         string
		char         collection.MediaCharacter
		rarity       string
		usersWanting string
	}{
		{
			name: "basic roll",
			char: collection.MediaCharacter{
				ID:         123,
				Name:       "Test Character",
				ImageURL:   "https://example.com/image.png",
				URL:        "https://anilist.co/character/123",
				MediaTitle: "Test Anime",
				Favorites:  150,
			},
			rarity:       "Uncommon",
			usersWanting: "",
		},
		{
			name: "with users wanting",
			char: collection.MediaCharacter{
				ID:         456,
				Name:       "Popular Character",
				ImageURL:   "https://example.com/popular.png",
				URL:        "https://anilist.co/character/456",
				MediaTitle: "Popular Anime",
				Favorites:  2500,
			},
			rarity:       "Rare",
			usersWanting: "\n\nWanted by: @user1, @user2",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			embed := rollEmbed(tt.char, tt.usersWanting)

			assert.Equal(t, tt.char.Name, embed.Title)
			assert.Equal(t, tt.char.URL, embed.URL)
			assert.Equal(t, collection.GradientColor(tt.char.Favorites), embed.Color)
			assert.Contains(t, embed.Description, tt.char.Name)
			assert.Contains(t, embed.Description, tt.char.MediaTitle)
			assert.Contains(t, embed.Description, tt.rarity)
			assert.Contains(t, embed.Description, "Standard Roll")
			assert.Contains(t, embed.Description, tt.char.URL)
			assert.NotEmpty(t, embed.Footer.Text)
			assert.Equal(t, AnilistIconURL, embed.Footer.IconURL)
		})
	}
}

func TestSeriesRollEmbed(t *testing.T) {
	for _, tt := range []struct {
		name string
		char collection.MediaCharacter
		cost int32
	}{
		{
			name: "series roll",
			char: collection.MediaCharacter{
				ID:         789,
				Name:       "Series Character",
				ImageURL:   "https://example.com/series.png",
				URL:        "https://anilist.co/character/789",
				MediaTitle: "Series Anime",
				Favorites:  500,
			},
			cost: 5,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			embed := seriesRollEmbed(tt.char, tt.cost)

			assert.Equal(t, tt.char.Name, embed.Title)
			assert.Equal(t, tt.char.URL, embed.URL)
			assert.Equal(t, collection.GradientColor(tt.char.Favorites), embed.Color)
			assert.Contains(t, embed.Description, tt.char.Name)
			assert.Contains(t, embed.Description, tt.char.MediaTitle)
			assert.Contains(t, embed.Description, "Series Roll")
			assert.Contains(t, embed.Description, "Cost: 5 tokens")
			assert.NotEmpty(t, embed.Footer.Text)
			assert.Equal(t, AnilistIconURL, embed.Footer.IconURL)
		})
	}
}

func TestClaimEmbed(t *testing.T) {
	for _, tt := range []struct {
		name   string
		char   collection.Character
		rarity string
	}{
		{
			name: "basic claim",
			char: collection.Character{
				ID:         999,
				Name:       "Claimed Character",
				Image:      "https://example.com/claimed.png",
				MediaTitle: "Claimed Anime",
				Favorites:  75,
			},
			rarity: "Common",
		},
		{
			name: "legendary claim",
			char: collection.Character{
				ID:         999,
				Name:       "Legend Claimed",
				Image:      "https://example.com/legend.png",
				MediaTitle: "Legend Anime",
				Favorites:  10000,
			},
			rarity: "Legendary",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			embed := claimEmbed(tt.char, tt.rarity)

			assert.Equal(t, tt.char.Name, embed.Title)
			assert.Equal(t, "https://anilist.co/character/999", embed.URL)
			assert.Equal(t, collection.GradientColor(tt.char.Favorites), embed.Color)
			assert.Contains(t, embed.Description, tt.char.Name)
			assert.Contains(t, embed.Description, tt.char.MediaTitle)
			assert.Contains(t, embed.Description, tt.rarity)
			assert.NotEmpty(t, embed.Footer.Text)
			assert.Equal(t, AnilistIconURL, embed.Footer.IconURL)
		})
	}
}

// Mock image fetcher for testing

type mockImageFetcher struct {
	image io.Reader
	err   error
}

func (m *mockImageFetcher) Fetch(ctx context.Context, url string) (io.ReadCloser, error) {
	if m.err != nil {
		return nil, m.err
	}
	return io.NopCloser(m.image), nil
}

type errorImageFetcher struct {
	err error
}

func (e *errorImageFetcher) Fetch(ctx context.Context, url string) (io.ReadCloser, error) {
	return nil, e.err
}

func TestBuildDropMessage(t *testing.T) {
	for _, tt := range []struct {
		name      string
		char      collection.MediaCharacter
		image     io.Reader
		wantTitle string
		wantImage bool
	}{
		{
			name: "without hint but with image",
			char: collection.MediaCharacter{
				ID:         222,
				Name:       "Character With Name",
				ImageURL:   "https://example.com/nohint.png",
				URL:        "https://anilist.co/character/222",
				MediaTitle: "No Hint Anime",
				Favorites:  1200,
			},
			image:     strings.NewReader("fake image data"),
			wantTitle: "Character Drop!",
			wantImage: true,
		},
		{
			name: "without image (nil)",
			char: collection.MediaCharacter{
				ID:         333,
				Name:       "Character With Name",
				ImageURL:   "https://example.com/noimage.png",
				URL:        "https://anilist.co/character/333",
				MediaTitle: "No Image Anime",
				Favorites:  450,
			},
			image:     nil,
			wantTitle: "Character Drop!",
			wantImage: false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			msg := dropMessage(tt.char, tt.image)

			require.Len(t, msg.Embeds, 1)
			embed := msg.Embeds[0]

			assert.Equal(t, tt.wantTitle, embed.Title)
			assert.Equal(t, tt.char.URL, embed.URL)
			assert.Equal(t, collection.GradientColor(tt.char.Favorites), embed.Color)
			assert.Contains(t, embed.Description, "A character has appeared!")
			assert.Contains(t, embed.Description, "/claim")

			assert.Contains(t, embed.Description, "C.W.N")

			assert.Contains(t, embed.Description, tt.char.Rarity().String())
			assert.Equal(t, AnilistIconURL, embed.Footer.IconURL)
			assert.Equal(t, "View on Anilist", embed.Footer.Text)

			if tt.wantImage {
				assert.Equal(t, "attachment://image.png", embed.Image.URL)
				require.Len(t, msg.Attachments, 1)
				assert.Equal(t, "image.png", msg.Attachments[0].Filename)
			} else {
				assert.Empty(t, embed.Image.URL)
				assert.Empty(t, msg.Attachments)
			}
		})
	}
}

func TestFetchCharacterImage(t *testing.T) {
	t.Run("successful fetch", func(t *testing.T) {
		fetcher := &mockImageFetcher{image: strings.NewReader("fake image")}

		body, err := fetcher.Fetch(t.Context(), "https://example.com/image.png")
		require.NoError(t, err)

		defer body.Close()

		data, err := io.ReadAll(body)
		require.NoError(t, err)
		assert.Equal(t, "fake image", string(data))
	})

	t.Run("fetch error", func(t *testing.T) {
		expectedErr := errors.New("network error")
		fetcher := &errorImageFetcher{err: expectedErr}

		body, err := fetcher.Fetch(t.Context(), "https://example.com/image.png")
		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, body)
	})
}

func TestHttpImageFetcher(t *testing.T) {
	t.Run("non-200 status code", func(t *testing.T) {
		fetcher := httpImageFetcher{doer: &errorHTTPClient{statusCode: http.StatusNotFound}}
		body, err := fetcher.Fetch(t.Context(), "https://example.com/image.png")

		assert.Nil(t, body)
		var httpErr *httpError
		assert.ErrorAs(t, err, &httpErr)
		assert.Equal(t, http.StatusNotFound, httpErr.StatusCode)
		assert.Contains(t, httpErr.URL, "example.com")
	})

	t.Run("invalid request", func(t *testing.T) {
		fetcher := httpImageFetcher{}
		body, err := fetcher.Fetch(t.Context(), "://invalid-url")

		assert.Nil(t, body)
		assert.Error(t, err)
	})
}

type errorHTTPClient struct {
	statusCode int
}

func (e *errorHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: e.statusCode,
		Body:       io.NopCloser(strings.NewReader("")),
	}, nil
}
