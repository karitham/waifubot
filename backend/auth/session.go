package auth

import (
	"context"
	"time"
)

// Session represents an authenticated user session.
type Session struct {
	Token     string
	UserID    uint64
	ExpiresAt time.Time
	CreatedAt time.Time
}

// SessionStore persists and retrieves sessions.
type SessionStore interface {
	CreateSession(ctx context.Context, userID uint64) (Session, error)
	GetSession(ctx context.Context, token string) (Session, error)
	DeleteSession(ctx context.Context, token string) error
	DeleteExpiredSessions(ctx context.Context) error
}
