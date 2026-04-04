package cordetest

import (
	"strings"
	"testing"

	"github.com/Karitham/corde"
	"github.com/stretchr/testify/assert"
)

// MockResponseWriter implements corde.ResponseWriter for testing.
// It records all calls and stores the last response for assertions.
type MockResponseWriter struct {
	AckCalled            bool
	RespondCalled        bool
	DeferedRespondCalled bool
	UpdateCalled         bool
	DeferedUpdateCalled  bool
	AutocompleteCalled   bool
	ModalCalled          bool
	LastRespond          corde.InteractionResponder
	LastUpdate           corde.InteractionResponder
	LastAutocomplete     corde.InteractionResponder
	LastModal            corde.Modal
}

func (m *MockResponseWriter) Ack() {
	m.AckCalled = true
}

func (m *MockResponseWriter) Respond(r corde.InteractionResponder) {
	m.RespondCalled = true
	m.LastRespond = r
}

func (m *MockResponseWriter) DeferedRespond() {
	m.DeferedRespondCalled = true
}

func (m *MockResponseWriter) Update(r corde.InteractionResponder) {
	m.UpdateCalled = true
	m.LastUpdate = r
}

func (m *MockResponseWriter) DeferedUpdate() {
	m.DeferedUpdateCalled = true
}

func (m *MockResponseWriter) Autocomplete(r corde.InteractionResponder) {
	m.AutocompleteCalled = true
	m.LastAutocomplete = r
}

func (m *MockResponseWriter) Modal(modal corde.Modal) {
	m.ModalCalled = true
	m.LastModal = modal
}

// Choices extracts choices from LastAutocomplete.
// Returns nil if Autocomplete was not called or the responder has no data.
func (m *MockResponseWriter) Choices() []corde.Choice[any] {
	if !m.AutocompleteCalled || m.LastAutocomplete == nil {
		return nil
	}
	data := m.LastAutocomplete.InteractionRespData()
	if data == nil {
		return nil
	}
	return data.Choices
}

// Responded is a shorthand for RespondCalled || DeferedRespondCalled.
func (m *MockResponseWriter) Responded() bool {
	return m.RespondCalled || m.DeferedRespondCalled
}

var _ corde.ResponseWriter = (*MockResponseWriter)(nil)

// AssertContains checks if the response contains the expected string.
// Checks Content first, then embed Title, Description, and Fields.
func (m *MockResponseWriter) AssertContains(t *testing.T, want string) bool {
	t.Helper()
	data := m.LastRespond.InteractionRespData()
	if data == nil {
		return assert.Fail(t, "no response data")
	}
	if data.Content != "" {
		return assert.Contains(t, data.Content, want)
	}
	for _, e := range data.Embeds {
		if strings.Contains(e.Title, want) || strings.Contains(e.Description, want) {
			return true
		}
		for _, f := range e.Fields {
			if strings.Contains(f.Name, want) || strings.Contains(f.Value, want) {
				return true
			}
		}
	}
	return assert.Fail(t, "response does not contain", want)
}
