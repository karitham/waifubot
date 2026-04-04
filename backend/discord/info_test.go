package discord

import (
	"testing"

	"github.com/Karitham/corde"
	"github.com/karitham/waifubot/discord/cordetest"
	"github.com/stretchr/testify/assert"
)

func TestInfoHandler_Info(t *testing.T) {
	tests := []struct {
		name  string
		cmd   CommandContext
		check func(t *testing.T, w *cordetest.MockResponseWriter)
	}{
		{
			name: "basic call",
			cmd:  &MockCommandContext{},
			check: func(t *testing.T, w *cordetest.MockResponseWriter) {
				t.Helper()
				assert.True(t, w.RespondCalled, "Respond should have been called")

				data := w.LastRespond.InteractionRespData()
				assert.NotNil(t, data)
				assert.Len(t, data.Embeds, 1)
				assert.Equal(t, "Info", data.Embeds[0].Title)

				desc := data.Embeds[0].Description
				assert.Contains(t, desc, "Version: go")
				assert.Contains(t, desc, "Runtime: ")
				assert.Contains(t, desc, "Goroutines: ")
				assert.Contains(t, desc, "Memory: ")
				assert.Contains(t, desc, "GC Pauses: ")
			},
		},
		{
			name: "ephemeral flag",
			cmd:  &MockCommandContext{},
			check: func(t *testing.T, w *cordetest.MockResponseWriter) {
				t.Helper()
				assert.True(t, w.RespondCalled, "Respond should have been called")

				data := w.LastRespond.InteractionRespData()
				assert.NotNil(t, data)
				assert.True(t, data.Flags&corde.RESPONSE_FLAGS_EPHEMERAL != 0, "response should be ephemeral")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cordetest.MockResponseWriter{}
			h := &InfoHandler{}
			h.Info(t.Context(), w, tt.cmd)
			tt.check(t, w)
		})
	}
}
