package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/karitham/waifubot/auth"
	"github.com/karitham/waifubot/rest/api"
)

type userIDKey struct{}

// SecurityHandler implements ogen's generated SecurityHandler interface for bearer auth.
type SecurityHandler struct {
	sessions auth.SessionStore
}

// NewSecurityHandler creates a SecurityHandler backed by the given session store.
func NewSecurityHandler(sessions auth.SessionStore) *SecurityHandler {
	return &SecurityHandler{sessions: sessions}
}

// HandleBearerAuth validates the bearer token and injects the user ID into context.
func (s *SecurityHandler) HandleBearerAuth(ctx context.Context, operationName api.OperationName, t api.BearerAuth) (context.Context, error) {
	// The token already comes extracted from the Authorization header by ogen,
	// just validate it's not empty.
	token := t.GetToken()
	if token == "" {
		return nil, fmt.Errorf("empty bearer token")
	}

	session, err := s.sessions.GetSession(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid session")
	}

	return context.WithValue(ctx, userIDKey{}, session.UserID), nil
}

// ExtractBearerTokenFromRequest extracts the bearer token from an HTTP request.
// This is a convenience wrapper around auth.ExtractBearerToken for REST handlers.
func ExtractBearerTokenFromRequest(r *http.Request) (string, error) {
	return auth.ExtractBearerToken(r)
}

// UserIDFromContext extracts the authenticated user ID from context.
// Returns 0 if no user ID is present (should only happen in unauthenticated handlers).
func UserIDFromContext(ctx context.Context) uint64 {
	v, _ := ctx.Value(userIDKey{}).(uint64)
	return v
}
