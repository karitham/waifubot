package rest

import (
	"context"
	"errors"

	"github.com/karitham/waifubot/rest/api"
)

type ctxKey int

const (
	ctxUserID ctxKey = iota
	ctxToken
)

var errUnauthorized = errors.New("unauthorized")

func currentUserID(ctx context.Context) (uint64, bool) {
	id, ok := ctx.Value(ctxUserID).(uint64)
	return id, ok
}

func tokenFromContext(ctx context.Context) (string, bool) {
	t, ok := ctx.Value(ctxToken).(string)
	return t, ok
}

// HandleBearerAuth validates the bearer token and attaches the user ID and raw
// token to the request context. Handlers that need the current user call
// currentUserID(ctx); logout uses tokenFromContext to revoke the row.
func (s *Server) HandleBearerAuth(ctx context.Context, _ api.OperationName, t api.BearerAuth) (context.Context, error) {
	row, err := s.authStore.GetTokenWithUser(ctx, t.Token)
	if err != nil {
		return ctx, errUnauthorized
	}
	ctx = context.WithValue(ctx, ctxUserID, row.UserID)
	ctx = context.WithValue(ctx, ctxToken, t.Token)
	return ctx, nil
}
