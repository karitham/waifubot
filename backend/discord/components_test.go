package discord

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Karitham/corde"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscordMessageCleaner_ClearComponents(t *testing.T) {
	var method, path, auth, contentType, body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		auth = r.Header.Get("authorization")
		contentType = r.Header.Get("content-type")
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		body = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cleaner := discordMessageCleaner{
		api:      srv.URL,
		appID:    corde.Snowflake(123),
		botToken: "bot-token",
		client:   srv.Client(),
	}

	err := cleaner.ClearComponents("interaction-token", "Sold Rem for 1 token")

	require.NoError(t, err)
	assert.Equal(t, http.MethodPatch, method)
	assert.Equal(t, "/webhooks/123/interaction-token/messages/@original", path)
	assert.Equal(t, "Bot bot-token", auth)
	assert.Equal(t, "application/json", contentType)
	assert.JSONEq(t, `{"content":"Sold Rem for 1 token","components":[]}`, body)
}

func TestDiscordMessageCleaner_ClearComponentsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Unknown Message"}`))
	}))
	defer srv.Close()

	cleaner := discordMessageCleaner{api: srv.URL, appID: 123, client: srv.Client()}

	err := cleaner.ClearComponents("interaction-token", "Sold Rem for 1 token")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 404")
	assert.Contains(t, err.Error(), "Unknown Message")
}
