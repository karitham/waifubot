package rest

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/karitham/waifubot/storage/authstore"
	"github.com/karitham/waifubot/storage/userstore"
)

// discordTokenResponse is the subset of Discord's /oauth2/token response we use.
type discordTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// discordUser is the subset of Discord's /users/@me response we use.
type discordUser struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	GlobalName string `json:"global_name"`
	Avatar     string `json:"avatar"`
}

const tokenLifetime = 30 * 24 * time.Hour

// HandleAuthLogin redirects the browser to Discord's authorize endpoint with
// a freshly-generated CSRF state. Mounted as a plain chi route, not via ogen.
// The request's Origin/Referer is captured and stored with the state so the
// callback can redirect back to the right frontend.
func (s *Server) HandleAuthLogin(w http.ResponseWriter, r *http.Request) {
	origin := originFromRequest(r)
	if origin == "" {
		http.Error(w, "origin required", http.StatusBadRequest)
		return
	}
	if !isAllowedOrigin(origin, s.oauth.AllowedOrigins) {
		http.Error(w, "origin not allowed", http.StatusForbidden)
		return
	}
	state := s.states.New(origin)
	http.Redirect(w, r, discordAuthorizeURL(s.oauth, state), http.StatusFound)
}

// HandleAuthCallback validates the state, exchanges the code, links the
// Discord user to a local account, issues a bearer token, and redirects
// the browser to the configured FRONTEND_URL with the token in the URL
// fragment so it never reaches server logs.
func (s *Server) HandleAuthCallback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	code := q.Get("code")
	state := q.Get("state")
	if code == "" || state == "" {
		http.Error(w, "missing code or state", http.StatusBadRequest)
		return
	}
	origin, ok := s.states.Consume(state)
	if !ok {
		http.Error(w, "invalid or expired state", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	discordToken, err := s.exchangeDiscordCode(ctx, code)
	if err != nil {
		slog.ErrorContext(ctx, "discord token exchange failed", "error", err)
		http.Error(w, "discord token exchange failed", http.StatusBadGateway)
		return
	}

	dUser, err := s.fetchDiscordUser(ctx, discordToken)
	if err != nil {
		slog.ErrorContext(ctx, "discord user fetch failed", "error", err)
		http.Error(w, "discord user fetch failed", http.StatusBadGateway)
		return
	}

	discordID, err := strconv.ParseUint(dUser.ID, 10, 64)
	if err != nil {
		http.Error(w, "invalid discord id", http.StatusBadGateway)
		return
	}

	displayName := strings.TrimSpace(dUser.GlobalName)
	if displayName == "" {
		displayName = dUser.Username
	}

	// Store the raw avatar hash; the discord_avatar column is varchar(34)
	// (fits a 32-char hash). DiscordAvatarURL builds the full URL on read.
	if _, err := s.userStore.UpsertFromDiscord(ctx, userstore.UpsertFromDiscordParams{
		UserID:          discordID,
		DiscordUsername: displayName,
		DiscordAvatar:   dUser.Avatar,
	}); err != nil {
		slog.ErrorContext(ctx, "user upsert failed", "error", err, "discord_id", discordID)
		http.Error(w, "user upsert failed", http.StatusInternalServerError)
		return
	}

	token, err := newBearerToken()
	if err != nil {
		slog.ErrorContext(ctx, "token generation failed", "error", err)
		http.Error(w, "token generation failed", http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(tokenLifetime)
	if err := s.authStore.CreateToken(ctx, authstore.CreateTokenParams{
		Token:     token,
		UserID:    discordID,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	}); err != nil {
		slog.ErrorContext(ctx, "token store failed", "error", err)
		http.Error(w, "token store failed", http.StatusInternalServerError)
		return
	}

	redirect := origin
	if redirect == "" {
		// Fallback: no origin captured (e.g., same-origin nav with neither
		// Origin nor Referer). Use the first allowlisted origin.
		if len(s.oauth.AllowedOrigins) > 0 {
			redirect = s.oauth.AllowedOrigins[0]
		} else {
			redirect = "/"
		}
	}
	http.Redirect(w, r, redirect+"#token="+token, http.StatusFound)
}

func (s *Server) exchangeDiscordCode(ctx context.Context, code string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", "https://discord.com/api/oauth2/token",
		strings.NewReader(discordTokenRequestForm(s.oauth, code)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("discord token status=%d body=%s", resp.StatusCode, string(body))
	}
	var out discordTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.AccessToken, nil
}

func (s *Server) fetchDiscordUser(ctx context.Context, accessToken string) (*discordUser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://discord.com/api/users/@me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discord user status=%d body=%s", resp.StatusCode, string(body))
	}
	var u discordUser
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
}

func newBearerToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
