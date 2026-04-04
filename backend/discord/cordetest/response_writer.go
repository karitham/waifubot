package cordetest

import "github.com/Karitham/corde"

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
