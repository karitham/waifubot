package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/karitham/waifubot/storage/sessionstore"
)

const (
	tokenBytes = 32
	sessionTTL = 7 * 24 * time.Hour
)

// PgSessionStore implements SessionStore backed by Postgres via sqlc.
type PgSessionStore struct {
	q sessionstore.Querier
}

// NewPgSessionStore wraps a sqlc Querier for session persistence.
func NewPgSessionStore(q sessionstore.Querier) *PgSessionStore {
	return &PgSessionStore{q: q}
}

func (s *PgSessionStore) CreateSession(ctx context.Context, userID uint64) (Session, error) {
	token, err := generateToken()
	if err != nil {
		return Session{}, fmt.Errorf("generate token: %w", err)
	}

	expiresAt := time.Now().Add(sessionTTL)

	row, err := s.q.CreateSession(ctx, sessionstore.CreateSessionParams{
		Token:  token,
		UserID: userID,
		ExpiresAt: pgtype.Timestamp{
			Time:  expiresAt.UTC(),
			Valid: true,
		},
	})
	if err != nil {
		return Session{}, fmt.Errorf("insert session: %w", err)
	}

	return Session{
		Token:     row.Token,
		UserID:    row.UserID,
		CreatedAt: row.CreatedAt.Time,
		ExpiresAt: row.ExpiresAt.Time,
	}, nil
}

func (s *PgSessionStore) GetSession(ctx context.Context, token string) (Session, error) {
	row, err := s.q.GetSession(ctx, token)
	if err != nil {
		return Session{}, fmt.Errorf("get session: %w", err)
	}

	if row.ExpiresAt.Time.Before(time.Now()) {
		return Session{}, errors.New("session expired")
	}

	return Session{
		Token:     row.Token,
		UserID:    row.UserID,
		CreatedAt: row.CreatedAt.Time,
		ExpiresAt: row.ExpiresAt.Time,
	}, nil
}

func (s *PgSessionStore) DeleteSession(ctx context.Context, token string) error {
	return s.q.DeleteSession(ctx, token)
}

func (s *PgSessionStore) DeleteExpiredSessions(ctx context.Context) error {
	_, err := s.q.DeleteExpiredSessions(ctx)
	return err
}

func generateToken() (string, error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}
