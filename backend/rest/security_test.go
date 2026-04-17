package rest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/karitham/waifubot/auth"
	"github.com/karitham/waifubot/rest/api"
)

type mockSessionStore struct {
	sessions map[string]auth.Session
	err      error
}

func (m *mockSessionStore) CreateSession(_ context.Context, _ uint64) (auth.Session, error) {
	return auth.Session{}, nil
}

func (m *mockSessionStore) GetSession(_ context.Context, token string) (auth.Session, error) {
	if m.err != nil {
		return auth.Session{}, m.err
	}
	s, ok := m.sessions[token]
	if !ok {
		return auth.Session{}, errors.New("session not found")
	}
	if s.ExpiresAt.Before(time.Now()) {
		return auth.Session{}, errors.New("session expired")
	}
	return s, nil
}

func (m *mockSessionStore) DeleteSession(_ context.Context, _ string) error {
	return nil
}

func (m *mockSessionStore) DeleteExpiredSessions(_ context.Context) error {
	return nil
}

func TestSecurityHandler_HandleBearerAuth(t *testing.T) {
	validSession := auth.Session{
		Token:     "valid-token",
		UserID:    42,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}

	expiredSession := auth.Session{
		Token:     "expired-token",
		UserID:    99,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		CreatedAt: time.Now().Add(-48 * time.Hour),
	}

	tests := []struct {
		name    string
		token   string
		setup   func(m *mockSessionStore)
		wantErr bool
		wantUID uint64
	}{
		{
			name:    "valid session",
			token:   "valid-token",
			setup:   func(m *mockSessionStore) { m.sessions = map[string]auth.Session{"valid-token": validSession} },
			wantErr: false,
			wantUID: 42,
		},
		{
			name:    "empty token",
			token:   "",
			setup:   func(m *mockSessionStore) {},
			wantErr: true,
			wantUID: 0,
		},
		{
			name:    "session not found",
			token:   "unknown-token",
			setup:   func(m *mockSessionStore) { m.sessions = map[string]auth.Session{} },
			wantErr: true,
			wantUID: 0,
		},
		{
			name:    "store error",
			token:   "valid-token",
			setup:   func(m *mockSessionStore) { m.err = errors.New("db error") },
			wantErr: true,
			wantUID: 0,
		},
		{
			name:    "expired session",
			token:   "expired-token",
			setup:   func(m *mockSessionStore) { m.sessions = map[string]auth.Session{"expired-token": expiredSession} },
			wantErr: true,
			wantUID: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockSessionStore{}
			tt.setup(m)

			h := NewSecurityHandler(m)
			ba := api.BearerAuth{Token: tt.token}

			ctx, err := h.HandleBearerAuth(context.Background(), "TestOp", ba)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, ctx)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantUID, UserIDFromContext(ctx))
		})
	}
}

func TestUserIDFromContext(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want uint64
	}{
		{
			name: "with user id",
			ctx:  context.WithValue(context.Background(), userIDKey{}, uint64(42)),
			want: 42,
		},
		{
			name: "without user id",
			ctx:  context.Background(),
			want: 0,
		},
		{
			name: "wrong type",
			ctx:  context.WithValue(context.Background(), userIDKey{}, "not-a-uint64"),
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, UserIDFromContext(tt.ctx))
		})
	}
}
