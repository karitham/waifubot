package rest

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// OAuthConfig holds the Discord OAuth application credentials and the URL
// policy for the login flow. Set via rest.New and populated from flags.
//
// AllowedOrigins is the list of origins the browser may be redirected to
// after Discord auth completes. An empty list disables the login flow.
//
// Origin capture happens on the login request (where the Origin/Referer
// header is set by the browser) and is stored alongside the OAuth state,
// so the callback can use it even though the callback's own Origin/Referer
// is discord.com.
type OAuthConfig struct {
	ClientID       string
	ClientSecret   string
	RedirectURL    string
	AllowedOrigins []string
}

// stateStore is a short-lived CSRF token store for the Discord OAuth flow.
// States are single-use and expire after 10 minutes. The store is in-memory
// and not shared across instances; in a multi-instance deployment swap for
// the authstore or a Redis-backed equivalent.
type stateStore struct {
	mu     sync.Mutex
	states map[string]stateEntry
	ttl    time.Duration
}

type stateEntry struct {
	expiresAt time.Time
	origin    string
}

func newStateStore() *stateStore {
	return &stateStore{
		states: make(map[string]stateEntry),
		ttl:    10 * time.Minute,
	}
}

// New generates a state token and stores it with an expiry and the
// requesting browser's origin. The origin is used in the callback to
// redirect back to the frontend that initiated login.
func (s *stateStore) New(origin string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	state := base64.RawURLEncoding.EncodeToString(b)
	s.states[state] = stateEntry{
		expiresAt: time.Now().Add(s.ttl),
		origin:    origin,
	}
	s.gc()
	return state
}

// Consume validates and removes a state token. Returns the origin captured
// at login (may be empty if none was provided) and true only if the state
// existed, was not expired, and has not been consumed before.
func (s *stateStore) Consume(state string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.states[state]
	if !ok {
		return "", false
	}
	delete(s.states, state)
	if time.Now().After(entry.expiresAt) {
		return "", false
	}
	return entry.origin, true
}

func (s *stateStore) gc() {
	now := time.Now()
	for k, v := range s.states {
		if now.After(v.expiresAt) {
			delete(s.states, k)
		}
	}
}

// originFromRequest extracts the browser's origin from the request. Priority:
//  1. ?origin= query param (the frontend knows its own origin; passing it
//     explicitly is robust to referrer-policy stripping that often nukes
//     the Referer header on cross-origin top-level navigations).
//  2. Origin header (sent on cross-origin CORS requests, including top-level
//     navigations in modern browsers).
//  3. Referer header (legacy fallback; some referrer policies strip it).
//  4. "" — caller decides what to do (e.g., fall back to allowlist default).
func originFromRequest(r *http.Request) string {
	if o := r.URL.Query().Get("origin"); o != "" {
		return o
	}
	if o := r.Header.Get("Origin"); o != "" {
		return o
	}
	if ref := r.Header.Get("Referer"); ref != "" {
		u, err := url.Parse(ref)
		if err == nil && u.Scheme != "" && u.Host != "" {
			return u.Scheme + "://" + u.Host
		}
	}
	return ""
}

// isAllowedOrigin reports whether origin is in the allowlist. Exact match
// only — no wildcard, no scheme downgrade, no subdomain tricks.
func isAllowedOrigin(origin string, allowed []string) bool {
	for _, a := range allowed {
		if a == origin {
			return true
		}
	}
	return false
}

// discordAuthorizeURL builds the URL the user is redirected to in order to
// authorize the app with Discord.
func discordAuthorizeURL(cfg OAuthConfig, state string) string {
	v := url.Values{
		"client_id":     {cfg.ClientID},
		"redirect_uri":  {cfg.RedirectURL},
		"response_type": {"code"},
		"scope":         {"identify"},
		"state":         {state},
	}
	return "https://discord.com/api/oauth2/authorize?" + v.Encode()
}

// discordTokenRequestForm returns the form-encoded body for exchanging an
// authorization code for an access token.
func discordTokenRequestForm(cfg OAuthConfig, code string) string {
	v := url.Values{
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {cfg.RedirectURL},
	}
	return v.Encode()
}

// splitAvatarHash handles the avatar value Discord returns. When the user has
// no custom avatar, the field is empty and the default avatar URL is used.
func splitAvatarHash(avatar string) (hash string, hasCustom bool) {
	avatar = strings.TrimSpace(avatar)
	return avatar, avatar != ""
}
