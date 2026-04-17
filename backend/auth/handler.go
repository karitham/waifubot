package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

// OAuthExchanger abstracts the Discord OAuth2 operations needed by the handler.
type OAuthExchanger interface {
	AuthorizeURL(state string) string
	Exchange(ctx context.Context, code string) (DiscordUser, error)
}

// Handler serves /auth/* chi routes for Discord OAuth2 authentication.
type Handler struct {
	sessions    SessionStore
	users       UserStore
	oauth       OAuthExchanger
	frontendURL string
}

// NewHandler creates a Handler with the given dependencies.
func NewHandler(sessions SessionStore, users UserStore, oauth OAuthExchanger, frontendURL string) *Handler {
	return &Handler{
		sessions:    sessions,
		users:       users,
		oauth:       oauth,
		frontendURL: frontendURL,
	}
}

// Routes registers the auth routes onto the given chi router.
func (h *Handler) Routes(r chi.Router) {
	r.Get("/login", h.login)
	r.Get("/callback", h.callback)
	r.Post("/logout", h.logout)
	r.Get("/me", h.me)
}

const oauthStateCookie = "oauth_state"

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	state, err := generateCSRFState()
	if err != nil {
		http.Error(w, "failed to generate state", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    state,
		MaxAge:   300,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/auth/callback",
	})

	http.Redirect(w, r, h.oauth.AuthorizeURL(state), http.StatusFound)
}

func (h *Handler) callback(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(oauthStateCookie)
	if err != nil {
		slog.With("err", err).Warn("missing state cookie in OAuth callback")
		http.Error(w, "missing state cookie", http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" || subtle.ConstantTimeCompare([]byte(state), []byte(cookie.Value)) != 1 {
		slog.Warn("invalid state parameter in OAuth callback")
		http.Error(w, "invalid state parameter", http.StatusBadRequest)
		return
	}

	clearStateCookie(w)

	code := r.URL.Query().Get("code")
	if code == "" {
		slog.Warn("missing code parameter in OAuth callback")
		http.Error(w, "missing code parameter", http.StatusBadRequest)
		return
	}

	discordUser, err := h.oauth.Exchange(r.Context(), code)
	if err != nil {
		slog.With("err", err).Error("OAuth exchange failed")
		http.Error(w, "oauth exchange failed", http.StatusBadGateway)
		return
	}

	discordID, err := strconv.ParseUint(discordUser.ID, 10, 64)
	if err != nil {
		slog.With("err", err, "discord_id", discordUser.ID).Warn("invalid discord user id")
		http.Error(w, "invalid discord user id", http.StatusInternalServerError)
		return
	}

	userID, err := h.users.GetOrCreateUser(r.Context(), discordID, discordUser.Username, discordUser.Avatar)
	if err != nil {
		slog.With("err", err, "discord_id", discordID).Error("failed to create user")
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	session, err := h.sessions.CreateSession(r.Context(), userID)
	if err != nil {
		slog.With("err", err, "user_id", userID).Error("failed to create session")
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	redirectURL := h.frontendURL + "/auth/callback#token=" + session.Token
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	token, err := ExtractBearerToken(r)
	if err != nil {
		slog.With("err", err).Warn("failed to extract bearer token in logout")
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err := h.sessions.DeleteSession(r.Context(), token); err != nil {
		slog.With("err", err).Error("failed to delete session")
		http.Error(w, "failed to delete session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "logged out"})
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	token, err := ExtractBearerToken(r)
	if err != nil {
		slog.With("err", err).Warn("failed to extract bearer token in /me")
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	session, err := h.sessions.GetSession(r.Context(), token)
	if err != nil {
		slog.With("err", err).Warn("invalid session in /me")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"user_id":    strconv.FormatUint(session.UserID, 10),
		"expires_at": session.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

func clearStateCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/auth/callback",
	})
}

func generateCSRFState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// ExtractBearerToken extracts the bearer token from an HTTP request authorization header.
// Returns an error if the header is missing, malformed, or empty.
func ExtractBearerToken(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("malformed authorization header")
	}

	scheme := strings.ToLower(parts[0])
	if scheme != "bearer" {
		return "", fmt.Errorf("invalid authorization scheme")
	}

	if parts[1] == "" {
		return "", fmt.Errorf("empty bearer token")
	}

	return parts[1], nil
}
