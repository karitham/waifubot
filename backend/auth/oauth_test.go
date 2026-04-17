package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

const (
	testClientID     = "test_client_id"
	testClientSecret = "test_client_secret"
	testRedirectURI  = "https://example.com/auth/callback"
	testState        = "random_state_123"
	testCode         = "auth_code_abc"
	testAccessToken  = "test_token"
)

func testConfig(tokenURL, userURL string) OAuthConfig {
	return OAuthConfig{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
		RedirectURI:  testRedirectURI,
	}
}

func TestAuthorizeURL(t *testing.T) {
	client := NewOAuthClient(OAuthConfig{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
		RedirectURI:  testRedirectURI,
	})

	gotURL := client.AuthorizeURL(testState)

	parsed, err := url.Parse(gotURL)
	if err != nil {
		t.Fatalf("failed to parse authorize URL: %v", err)
	}

	if parsed.Scheme != "https" {
		t.Errorf("scheme = %q, want https", parsed.Scheme)
	}
	if parsed.Host != "discord.com" {
		t.Errorf("host = %q, want discord.com", parsed.Host)
	}
	if parsed.Path != "/oauth2/authorize" {
		t.Errorf("path = %q, want /oauth2/authorize", parsed.Path)
	}

	q := parsed.Query()
	if got := q.Get("response_type"); got != "code" {
		t.Errorf("response_type = %q, want code", got)
	}
	if got := q.Get("client_id"); got != testClientID {
		t.Errorf("client_id = %q, want %q", got, testClientID)
	}
	if got := q.Get("scope"); got != "identify" {
		t.Errorf("scope = %q, want identify", got)
	}
	if got := q.Get("redirect_uri"); got != testRedirectURI {
		t.Errorf("redirect_uri = %q, want %q", got, testRedirectURI)
	}
	if got := q.Get("state"); got != testState {
		t.Errorf("state = %q, want %q", got, testState)
	}
}

// setupMockServer creates an httptest server that mocks both Discord token and user endpoints.
// The token handler validates Basic auth and returns a test access token.
// The user handler validates Bearer auth and returns a test Discord user.
func setupMockServer(tokenHandler, userHandler http.HandlerFunc) *httptest.Server {
	mux := http.NewServeMux()
	if tokenHandler != nil {
		mux.HandleFunc("/api/v10/oauth2/token", tokenHandler)
	}
	if userHandler != nil {
		mux.HandleFunc("/api/v10/users/@me", userHandler)
	}
	return httptest.NewServer(mux)
}

// validTokenHandler returns a handler that validates the token exchange request and returns a test token.
func validTokenHandler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if r.Method != http.MethodPost {
			t.Errorf("token request method = %q, want POST", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Validate Basic Auth
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Basic ") {
			t.Errorf("missing Basic auth header")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))
		if err != nil {
			t.Errorf("failed to decode Basic auth: %v", err)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		parts := strings.SplitN(string(decoded), ":", 2)
		if len(parts) != 2 || parts[0] != testClientID || parts[1] != testClientSecret {
			t.Errorf("Basic auth credentials = %q, want %q:%q", string(decoded), testClientID, testClientSecret)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Validate content type
		if ct := r.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
			t.Errorf("content-type = %q, want application/x-www-form-urlencoded", ct)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		// Validate body params
		if err := r.ParseForm(); err != nil {
			t.Errorf("failed to parse form: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if got := r.FormValue("grant_type"); got != "authorization_code" {
			t.Errorf("grant_type = %q, want authorization_code", got)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if got := r.FormValue("code"); got != testCode {
			t.Errorf("code = %q, want %q", got, testCode)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: testAccessToken,
			TokenType:   "Bearer",
		})
	}
}

// validUserHandler returns a handler that validates the Bearer token and returns a test user.
func validUserHandler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if r.Method != http.MethodGet {
			t.Errorf("user request method = %q, want GET", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		auth := r.Header.Get("Authorization")
		want := "Bearer " + testAccessToken
		if auth != want {
			t.Errorf("Authorization header = %q, want %q", auth, want)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(DiscordUser{
			ID:       "123456789",
			Username: "testuser",
			Avatar:   "abc123",
		})
	}
}

func TestExchange_Success(t *testing.T) {
	srv := setupMockServer(validTokenHandler(t), validUserHandler(t))
	defer srv.Close()

	client := NewOAuthClient(OAuthConfig{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
		RedirectURI:  testRedirectURI,
	})
	// Override httpClient to point at test server
	client.httpClient = srv.Client()

	// Redirect client requests to test server by replacing URLs via a transport wrapper.
	// We use a custom RoundTripper to rewrite URLs.
	client.httpClient.Transport = &urlRewriter{
		base:     srv.URL,
		original: client.httpClient.Transport,
	}

	user, err := client.Exchange(context.Background(), testCode)
	if err != nil {
		t.Fatalf("Exchange() error = %v", err)
	}

	if user.ID != "123456789" {
		t.Errorf("user.ID = %q, want 123456789", user.ID)
	}
	if user.Username != "testuser" {
		t.Errorf("user.Username = %q, want testuser", user.Username)
	}
	if user.Avatar != "abc123" {
		t.Errorf("user.Avatar = %q, want abc123", user.Avatar)
	}
}

func TestExchange_TokenEndpointError(t *testing.T) {
	srv := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"error": "server_error"}`))
	}, nil)
	defer srv.Close()

	client := NewOAuthClient(OAuthConfig{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
		RedirectURI:  testRedirectURI,
	})
	client.httpClient = srv.Client()
	client.httpClient.Transport = &urlRewriter{
		base:     srv.URL,
		original: client.httpClient.Transport,
	}

	_, err := client.Exchange(context.Background(), testCode)
	if err == nil {
		t.Fatal("Exchange() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "exchange code") {
		t.Errorf("error = %q, want to contain 'exchange code'", err.Error())
	}
	if !strings.Contains(err.Error(), "502") {
		t.Errorf("error = %q, want to contain '502'", err.Error())
	}
}

func TestExchange_InvalidCode(t *testing.T) {
	srv := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid_grant"}`))
	}, nil)
	defer srv.Close()

	client := NewOAuthClient(OAuthConfig{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
		RedirectURI:  testRedirectURI,
	})
	client.httpClient = srv.Client()
	client.httpClient.Transport = &urlRewriter{
		base:     srv.URL,
		original: client.httpClient.Transport,
	}

	_, err := client.Exchange(context.Background(), "bad_code")
	if err == nil {
		t.Fatal("Exchange() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "exchange code") {
		t.Errorf("error = %q, want to contain 'exchange code'", err.Error())
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error = %q, want to contain '401'", err.Error())
	}
}

func TestExchange_DiscordAPIError(t *testing.T) {
	srv := setupMockServer(validTokenHandler(t), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	})
	defer srv.Close()

	client := NewOAuthClient(OAuthConfig{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
		RedirectURI:  testRedirectURI,
	})
	client.httpClient = srv.Client()
	client.httpClient.Transport = &urlRewriter{
		base:     srv.URL,
		original: client.httpClient.Transport,
	}

	_, err := client.Exchange(context.Background(), testCode)
	if err == nil {
		t.Fatal("Exchange() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "fetch user") {
		t.Errorf("error = %q, want to contain 'fetch user'", err.Error())
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error = %q, want to contain '500'", err.Error())
	}
}

// urlRewriter is a http.RoundTripper that rewrites requests to Discord's API
// to point at the test server instead.
type urlRewriter struct {
	base     string
	original http.RoundTripper
}

func (r *urlRewriter) RoundTrip(req *http.Request) (*http.Response, error) {
	target, err := url.Parse(r.base)
	if err != nil {
		return nil, err
	}
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	transport := r.original
	if transport == nil {
		transport = http.DefaultTransport
	}
	return transport.RoundTrip(req)
}
