package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/retrypolicy"
)

// OAuthConfig holds Discord OAuth2 credentials and settings.
type OAuthConfig struct {
	ClientID     string // Discord APP_ID
	ClientSecret string // OAUTH_CLIENT_SECRET
	RedirectURI  string // e.g. "https://waifuapi.karitham.dev/auth/callback"
}

// DiscordUser represents the user info returned by Discord's /users/@me endpoint.
type DiscordUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

// OAuthClient handles Discord OAuth2 token exchange and user lookup.
type OAuthClient struct {
	cfg        OAuthConfig
	httpClient *http.Client
}

// NewOAuthClient creates a new OAuth client with a 5-second HTTP timeout.
func NewOAuthClient(cfg OAuthConfig) *OAuthClient {
	return &OAuthClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// AuthorizeURL builds the Discord OAuth2 authorization redirect URL.
// The "identify" scope is sufficient because we only need the user's ID, username, and avatar
// from the /users/@me endpoint. We don't need guilds, messages, or any other permissions.
func (c *OAuthClient) AuthorizeURL(state string) string {
	v := url.Values{}
	v.Set("response_type", "code")
	v.Set("client_id", c.cfg.ClientID)
	v.Set("scope", "identify")
	v.Set("redirect_uri", c.cfg.RedirectURI)
	v.Set("state", state)
	return "https://discord.com/oauth2/authorize?" + v.Encode()
}

// tokenResponse represents the JSON response from Discord's token endpoint.
// Note: We intentionally ignore expires_in because we issue our own session tokens
// via the session store, not using Discord's access token directly.
// Discord's token is only used as a temporary credential to fetch user info.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// Exchange performs the full OAuth2 code exchange:
// 1. POST to discord.com/api/v10/oauth2/token with the authorization code
// 2. GET discord.com/api/v10/users/@me with the access token
// 3. Return the DiscordUser
func (c *OAuthClient) Exchange(ctx context.Context, code string) (DiscordUser, error) {
	accessToken, err := c.exchangeCode(ctx, code)
	if err != nil {
		return DiscordUser{}, fmt.Errorf("exchange code: %w", err)
	}

	user, err := c.fetchUser(ctx, accessToken)
	if err != nil {
		return DiscordUser{}, fmt.Errorf("fetch user: %w", err)
	}

	return user, nil
}

// newRetryPolicy creates a retry policy with exponential backoff for transient errors.
func newRetryPolicy() failsafe.Executor[string] {
	rp := retrypolicy.NewBuilder[string]().
		WithMaxRetries(3).
		WithBackoff(100*time.Millisecond, 400*time.Millisecond).
		Build()
	return failsafe.With(rp)
}

// exchangeCode POSTs to Discord's token endpoint to exchange an authorization code for an access token.
// It retries with exponential backoff on transient errors (5xx status codes).
func (c *OAuthClient) exchangeCode(ctx context.Context, code string) (string, error) {
	v := url.Values{}
	v.Set("grant_type", "authorization_code")
	v.Set("code", code)
	v.Set("redirect_uri", c.cfg.RedirectURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://discord.com/api/v10/oauth2/token", strings.NewReader(v.Encode()))
	if err != nil {
		return "", fmt.Errorf("create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Discord requires HTTP Basic Auth with client_id:client_secret for confidential clients.
	creds := base64.StdEncoding.EncodeToString([]byte(c.cfg.ClientID + ":" + c.cfg.ClientSecret))
	req.Header.Set("Authorization", "Basic "+creds)

	executor := newRetryPolicy()
	result, err := executor.Get(func() (string, error) {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("token request: %w", err)
		}
		defer resp.Body.Close()

		// Retry on 5xx errors
		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			body, _ := io.ReadAll(resp.Body)
			return "", fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, body)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return "", fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, body)
		}

		var tok tokenResponse
		if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
			return "", fmt.Errorf("decode token response: %w", err)
		}

		return tok.AccessToken, nil
	})

	return result, err
}

// fetchUser calls Discord's /users/@me endpoint with the given access token.
// It retries with exponential backoff on transient errors (5xx status codes).
func (c *OAuthClient) fetchUser(ctx context.Context, accessToken string) (DiscordUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://discord.com/api/v10/users/@me", nil)
	if err != nil {
		return DiscordUser{}, fmt.Errorf("create user request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	rp := retrypolicy.NewBuilder[DiscordUser]().
		WithMaxRetries(3).
		WithBackoff(100*time.Millisecond, 400*time.Millisecond).
		Build()
	executor := failsafe.With(rp)

	result, err := executor.Get(func() (DiscordUser, error) {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return DiscordUser{}, fmt.Errorf("user request: %w", err)
		}
		defer resp.Body.Close()

		// Retry on 5xx errors
		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			body, _ := io.ReadAll(resp.Body)
			return DiscordUser{}, fmt.Errorf("user endpoint returned %d: %s", resp.StatusCode, body)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return DiscordUser{}, fmt.Errorf("user endpoint returned %d: %s", resp.StatusCode, body)
		}

		var user DiscordUser
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			return DiscordUser{}, fmt.Errorf("decode user response: %w", err)
		}

		return user, nil
	})

	return result, err
}
