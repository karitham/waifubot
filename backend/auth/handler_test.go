package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

// --- Mocks ---

type mockSessionStore struct {
	sessions map[string]Session
	createFn func(ctx context.Context, userID uint64) (Session, error)
	deleteFn func(ctx context.Context, token string) error
	err      error
}

func (m *mockSessionStore) CreateSession(ctx context.Context, userID uint64) (Session, error) {
	if m.createFn != nil {
		return m.createFn(ctx, userID)
	}
	token := fmt.Sprintf("tok_%d", userID)
	s := Session{
		Token:     token,
		UserID:    userID,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
	}
	if m.sessions != nil {
		m.sessions[token] = s
	}
	return s, m.err
}

func (m *mockSessionStore) GetSession(ctx context.Context, token string) (Session, error) {
	if m.sessions != nil {
		if s, ok := m.sessions[token]; ok {
			return s, nil
		}
	}
	return Session{}, fmt.Errorf("session not found")
}

func (m *mockSessionStore) DeleteSession(ctx context.Context, token string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, token)
	}
	delete(m.sessions, token)
	return m.err
}

func (m *mockSessionStore) DeleteExpiredSessions(ctx context.Context) error {
	return nil
}

type mockOAuthClient struct {
	authorizeURL string
	exchangeUser DiscordUser
	exchangeErr  error
}

func (m *mockOAuthClient) AuthorizeURL(state string) string {
	if m.authorizeURL != "" {
		return m.authorizeURL + "?state=" + state
	}
	return "https://discord.com/oauth2/authorize?state=" + state
}

func (m *mockOAuthClient) Exchange(ctx context.Context, code string) (DiscordUser, error) {
	return m.exchangeUser, m.exchangeErr
}

type mockUserStore struct {
	users     map[uint64]uint64
	createErr error
	idOffset  uint64 // offset to simulate internal IDs differing from Discord IDs
}

func newMockUserStore() *mockUserStore {
	return &mockUserStore{users: make(map[uint64]uint64)}
}

func (m *mockUserStore) GetOrCreateUser(ctx context.Context, discordID uint64, username, avatar string) (uint64, error) {
	if m.createErr != nil {
		return 0, m.createErr
	}
	if uid, ok := m.users[discordID]; ok {
		return uid, nil
	}
	// Simulate internal ID differing from Discord ID
	internalID := discordID + m.idOffset
	m.users[discordID] = internalID
	return internalID, nil
}

// --- Helpers ----

func setupHandler(sessions SessionStore, users UserStore, oauth OAuthExchanger) *Handler {
	if oauth == nil {
		oauth = &mockOAuthClient{}
	}
	return NewHandler(sessions, users, oauth, "https://frontend.example.com")
}

func setupRouter(h *Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Route("/auth", h.Routes)
	return r
}

// --- Tests ---

func TestHandler_Login(t *testing.T) {
	sessions := &mockSessionStore{}
	users := newMockUserStore()
	oauth := &mockOAuthClient{}
	h := setupHandler(sessions, users, oauth)
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code, "should redirect to Discord")

	loc := w.Header().Get("Location")
	assert.Contains(t, loc, "discord.com/oauth2/authorize", "redirect URL should contain Discord authorize path")
	assert.Contains(t, loc, "state=", "redirect URL should contain state parameter")

	var stateCookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == oauthStateCookie {
			stateCookie = c
			break
		}
	}
	assert.NotNil(t, stateCookie, "oauth_state cookie should be set")
	if stateCookie != nil {
		assert.Equal(t, 300, stateCookie.MaxAge)
		assert.True(t, stateCookie.HttpOnly)
		assert.Equal(t, http.SameSiteLaxMode, stateCookie.SameSite)
		assert.Equal(t, "/auth/callback", stateCookie.Path)
		// The state in the URL should match the cookie value
		assert.Contains(t, loc, "state="+stateCookie.Value)
	}
}

func TestHandler_Callback(t *testing.T) {
	validState := "abcdef0123456789"

	tests := []struct {
		name          string
		cookieState   string // value of oauth_state cookie; empty means no cookie
		queryState    string
		queryCode     string
		exchangeUser  DiscordUser
		exchangeErr   error
		createUserErr error
		createSession func(ctx context.Context, userID uint64) (Session, error)
		wantCode      int
		wantLocation  string // substring to check in Location header
	}{
		{
			name:         "valid exchange",
			cookieState:  validState,
			queryState:   validState,
			queryCode:    "valid_code",
			exchangeUser: DiscordUser{ID: "123456789", Username: "testuser", Avatar: "avatar1"},
			createSession: func(_ context.Context, userID uint64) (Session, error) {
				return Session{Token: "ses_tok_123", UserID: userID, ExpiresAt: time.Now().Add(7 * 24 * time.Hour)}, nil
			},
			wantCode:     http.StatusFound,
			wantLocation: "#token=ses_tok_123",
		},
		{
			name:         "invalid state",
			cookieState:  validState,
			queryState:   "wrong_state",
			queryCode:    "valid_code",
			exchangeUser: DiscordUser{ID: "123456789", Username: "testuser", Avatar: "avatar1"},
			wantCode:     http.StatusBadRequest,
		},
		{
			name:         "missing state cookie",
			cookieState:  "", // no cookie
			queryState:   validState,
			queryCode:    "valid_code",
			exchangeUser: DiscordUser{ID: "123456789", Username: "testuser", Avatar: "avatar1"},
			wantCode:     http.StatusBadRequest,
		},
		{
			// Test state with different length - constant-time compare should reject all mismatches
			name:         "invalid state - different length",
			cookieState:  validState,
			queryState:   "short",
			queryCode:    "valid_code",
			exchangeUser: DiscordUser{ID: "123456789", Username: "testuser", Avatar: "avatar1"},
			wantCode:     http.StatusBadRequest,
		},
		{
			// Test state same length but all zeros - should reject (not equal to valid state)
			name:         "invalid state - zeros mismatched",
			cookieState:  validState,
			queryState:   "0000000000000000",
			queryCode:    "valid_code",
			exchangeUser: DiscordUser{ID: "123456789", Username: "testuser", Avatar: "avatar1"},
			wantCode:     http.StatusBadRequest,
		},
		{
			// Test empty query state - should reject
			name:         "empty state query",
			cookieState:  validState,
			queryState:   "",
			queryCode:    "valid_code",
			exchangeUser: DiscordUser{ID: "123456789", Username: "testuser", Avatar: "avatar1"},
			wantCode:     http.StatusBadRequest,
		},
		{
			name:        "oauth exchange failure",
			cookieState: validState,
			queryState:  validState,
			queryCode:   "valid_code",
			exchangeErr: fmt.Errorf("discord down"),
			wantCode:    http.StatusBadGateway,
		},
		{
			name:          "user creation failure",
			cookieState:   validState,
			queryState:    validState,
			queryCode:     "valid_code",
			exchangeUser:  DiscordUser{ID: "123456789", Username: "testuser", Avatar: "avatar1"},
			createUserErr: fmt.Errorf("db error"),
			wantCode:      http.StatusInternalServerError,
		},
		{
			name:         "session creation failure",
			cookieState:  validState,
			queryState:   validState,
			queryCode:    "valid_code",
			exchangeUser: DiscordUser{ID: "123456789", Username: "testuser", Avatar: "avatar1"},
			createSession: func(_ context.Context, _ uint64) (Session, error) {
				return Session{}, fmt.Errorf("session db error")
			},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessions := &mockSessionStore{
				sessions: make(map[string]Session),
			}
			if tt.createSession != nil {
				sessions.createFn = tt.createSession
			}

			users := newMockUserStore()
			users.createErr = tt.createUserErr

			oauth := &mockOAuthClient{
				exchangeUser: tt.exchangeUser,
				exchangeErr:  tt.exchangeErr,
			}

			h := setupHandler(sessions, users, oauth)
			r := setupRouter(h)

			url := "/auth/callback?code=" + tt.queryCode + "&state=" + tt.queryState
			req := httptest.NewRequest(http.MethodGet, url, nil)
			if tt.cookieState != "" {
				req.AddCookie(&http.Cookie{
					Name:  oauthStateCookie,
					Value: tt.cookieState,
				})
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantLocation != "" {
				loc := w.Header().Get("Location")
				assert.Contains(t, loc, tt.wantLocation)
			}

			// On success, the oauth_state cookie should be cleared
			if tt.wantCode == http.StatusFound && tt.cookieState != "" {
				var cleared *http.Cookie
				for _, c := range w.Result().Cookies() {
					if c.Name == oauthStateCookie {
						cleared = c
						break
					}
				}
				if cleared != nil {
					assert.Equal(t, -1, cleared.MaxAge, "oauth_state cookie should be cleared")
				}
			}
		})
	}

	// Verify the user was created with the correct discord ID
	t.Run("creates user with correct discord ID", func(t *testing.T) {
		sessions := &mockSessionStore{
			sessions: make(map[string]Session),
			createFn: func(_ context.Context, userID uint64) (Session, error) {
				return Session{Token: "tok", UserID: userID, ExpiresAt: time.Now().Add(24 * time.Hour)}, nil
			},
		}
		users := newMockUserStore()
		oauth := &mockOAuthClient{
			exchangeUser: DiscordUser{ID: "999888777", Username: "newuser", Avatar: "av"},
		}

		h := setupHandler(sessions, users, oauth)
		r := setupRouter(h)

		req := httptest.NewRequest(http.MethodGet, "/auth/callback?code=abc&state="+validState, nil)
		req.AddCookie(&http.Cookie{Name: oauthStateCookie, Value: validState})

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusFound, w.Code)
		_, exists := users.users[uint64(999888777)]
		assert.True(t, exists, "user should be created with discord ID 999888777")
	})
}

// Test idempotent user handling - simulates a race where user already exists.
// This verifies GetOrCreateUser handles concurrent creation safely.
func TestHandler_IdempotentUserCreation(t *testing.T) {
	const testState = "teststate12345678"
	sessions := &mockSessionStore{
		sessions: make(map[string]Session),
		createFn: func(_ context.Context, userID uint64) (Session, error) {
			return Session{Token: "tok", UserID: userID, ExpiresAt: time.Now().Add(24 * time.Hour)}, nil
		},
	}
	users := newMockUserStore()
	// Pre-populate: user already exists (simulates race condition from concurrent request)
	users.users[111] = 111

	oauth := &mockOAuthClient{
		exchangeUser: DiscordUser{ID: "111", Username: "testuser", Avatar: "av"},
	}

	h := setupHandler(sessions, users, oauth)
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/auth/callback?code=abc&state="+testState, nil)
	req.AddCookie(&http.Cookie{Name: oauthStateCookie, Value: testState})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should succeed - existing user returned
	assert.Equal(t, http.StatusFound, w.Code, "idempotent user creation should succeed when user exists")
}

func TestHandler_Logout(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		deleteErr  error
		wantCode   int
		wantBody   string
	}{
		{
			name:       "valid token",
			authHeader: "Bearer valid_token",
			wantCode:   http.StatusOK,
			wantBody:   "logged out",
		},
		{
			name:     "missing auth header",
			wantCode: http.StatusUnauthorized,
		},
		{
			name:       "malformed auth header",
			authHeader: "Basic abc123",
			wantCode:   http.StatusUnauthorized,
		},
		{
			name:       "session delete error",
			authHeader: "Bearer valid_token",
			deleteErr:  fmt.Errorf("db error"),
			wantCode:   http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessions := &mockSessionStore{
				sessions: map[string]Session{
					"valid_token": {Token: "valid_token", UserID: 1},
				},
				deleteFn: func(_ context.Context, _ string) error {
					return tt.deleteErr
				},
			}
			users := newMockUserStore()

			h := setupHandler(sessions, users, nil)
			r := setupRouter(h)

			req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestHandler_Me(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		sessions   map[string]Session
		wantCode   int
		wantUserID string
	}{
		{
			name:       "valid session",
			authHeader: "Bearer valid_token",
			sessions: map[string]Session{
				"valid_token": {
					Token:     "valid_token",
					UserID:    42,
					ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
					CreatedAt: time.Now(),
				},
			},
			wantCode:   http.StatusOK,
			wantUserID: "42",
		},
		{
			name:     "missing auth header",
			wantCode: http.StatusUnauthorized,
		},
		{
			name:       "expired session",
			authHeader: "Bearer expired_token",
			sessions:   map[string]Session{}, // token not in map → not found
			wantCode:   http.StatusUnauthorized,
		},
		{
			name:       "invalid token",
			authHeader: "Bearer nonexistent",
			sessions:   map[string]Session{},
			wantCode:   http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessions := &mockSessionStore{sessions: tt.sessions}
			users := newMockUserStore()

			h := setupHandler(sessions, users, nil)
			r := setupRouter(h)

			req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantUserID != "" {
				var body map[string]string
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
				assert.Equal(t, tt.wantUserID, body["user_id"])
				assert.NotEmpty(t, body["expires_at"])
			}
		})
	}
}

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		want    string
		wantErr bool
	}{
		{"valid bearer", "Bearer abc123", "abc123", false},
		{"missing header", "", "", true},
		{"wrong scheme", "Basic abc123", "", true},
		{"empty token", "Bearer ", "", true},
		{"lowercase bearer", "bearer abc123", "abc123", false},
		{"mixed case bearer", "BeArEr abc123", "abc123", false},
		{"no space", "Bearerabc", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			got, err := ExtractBearerToken(req)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGenerateCSRFState(t *testing.T) {
	state, err := generateCSRFState()
	assert.NoError(t, err)
	assert.Len(t, state, 32, "hex-encoded 16 bytes = 32 chars")

	// Should be hex
	for _, c := range state {
		assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'),
			"character %q should be hex", c)
	}

	// Should generate different values
	other, err := generateCSRFState()
	assert.NoError(t, err)
	assert.NotEqual(t, state, other, "consecutive states should differ")
}

func TestClearStateCookie(t *testing.T) {
	w := httptest.NewRecorder()
	clearStateCookie(w)

	var cookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == oauthStateCookie {
			cookie = c
			break
		}
	}
	assert.NotNil(t, cookie)
	assert.Equal(t, -1, cookie.MaxAge)
	assert.Equal(t, "/auth/callback", cookie.Path)
}

// Verify the Handler callback redirects to the correct frontend URL
func TestHandler_Callback_FrontendRedirect(t *testing.T) {
	const customFrontend = "https://waifugui.karitham.dev"
	const state = "test_state_abc"

	sessions := &mockSessionStore{
		sessions: make(map[string]Session),
		createFn: func(_ context.Context, userID uint64) (Session, error) {
			return Session{Token: "mytoken", UserID: userID, ExpiresAt: time.Now().Add(24 * time.Hour)}, nil
		},
	}
	users := newMockUserStore()
	oauth := &mockOAuthClient{
		exchangeUser: DiscordUser{ID: "111", Username: "user", Avatar: "av"},
	}

	h := NewHandler(sessions, users, oauth, customFrontend)
	r := chi.NewRouter()
	r.Route("/auth", h.Routes)

	req := httptest.NewRequest(http.MethodGet, "/auth/callback?code=abc&state="+state, nil)
	req.AddCookie(&http.Cookie{Name: oauthStateCookie, Value: state})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	loc := w.Header().Get("Location")
	assert.True(t, strings.HasPrefix(loc, customFrontend), "redirect should start with frontend URL, got: %s", loc)
	assert.Contains(t, loc, "#token=mytoken")
}
