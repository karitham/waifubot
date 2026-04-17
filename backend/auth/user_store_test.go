package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/collection/collectiontest"
	"github.com/stretchr/testify/assert"
)

// --- Mock UserRepository (wrapper around collectiontest.MockStore) ---

type mockUserRepository struct {
	*collectiontest.MockStore
	users map[collection.UserID]collection.User
}

func newMockUserRepository() *mockUserRepository {
	m := &mockUserRepository{
		MockStore: &collectiontest.MockStore{},
		users:     make(map[collection.UserID]collection.User),
	}
	// Wrap the methods to use our in-memory user map
	m.MockStore.GetUserFunc = func(ctx context.Context, userID collection.UserID) (collection.User, error) {
		if u, ok := m.users[userID]; ok {
			return u, nil
		}
		return collection.User{}, collection.ErrNotFound
	}
	m.MockStore.CreateUserFunc = func(ctx context.Context, userID collection.UserID) error {
		// Default: create user with auto-generated internal ID
		m.users[userID] = collection.User{
			UserID: userID + 1000, // internal ID differs from Discord ID
			Date:   time.Now(),
		}
		return nil
	}
	m.MockStore.UpdateDiscordInfoFunc = func(ctx context.Context, userID collection.UserID, username, avatar string, lastUpdated time.Time) error {
		if u, ok := m.users[userID]; ok {
			u.DiscordUsername = username
			u.DiscordAvatar = avatar
			u.LastUpdated = lastUpdated
			m.users[userID] = u
		}
		return nil
	}
	return m
}

// --- Tests ---

func TestUserStoreAdapter_GetOrCreateUser(t *testing.T) {
	tests := []struct {
		name             string
		preExistingUsers map[uint64]collection.User // map of discordID -> User
		getUserErr       error
		createUserErr    error
		wantUserID       uint64
		wantErr          bool
	}{
		{
			name: "user exists - returns existing ID, updates Discord info",
			preExistingUsers: map[uint64]collection.User{
				12345: {UserID: 12345, DiscordUsername: "olduser", DiscordAvatar: "oldavatar"},
			},
			wantUserID: 12345,
			wantErr:    false,
		},
		{
			name:             "user doesn't exist - creates new user, returns new ID",
			preExistingUsers: map[uint64]collection.User{},
			wantUserID:       12345, // user_id = discordID (no separate internal ID)
			wantErr:          false,
		},
		{
			name:       "get user fails with non-not-found error - returns error",
			getUserErr: errors.New("database connection error"),
			wantErr:    true,
		},
		{
			name:          "create user fails with non-duplicate error - returns error",
			createUserErr: errors.New("insert failed"),
			wantErr:       true,
		},
		// Note: "GetUser fails after CreateUser" is no longer a valid test case
		// since we don't fetch after create - we return discordID directly.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockUserRepository()
			if tt.preExistingUsers != nil {
				repo.users = tt.preExistingUsers
			}
			if tt.getUserErr != nil {
				repo.MockStore.GetUserFunc = func(_ context.Context, _ collection.UserID) (collection.User, error) {
					return collection.User{}, tt.getUserErr
				}
			}
			if tt.createUserErr != nil {
				repo.MockStore.CreateUserFunc = func(_ context.Context, _ collection.UserID) error {
					return tt.createUserErr
				}
			}

			adapter := NewUserStoreAdapter(repo)
			got, err := adapter.GetOrCreateUser(context.Background(), 12345, "newuser", "newavatar")

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantUserID, got)
		})
	}
}

// Test concurrent creation (duplicate key) scenario
func TestUserStoreAdapter_GetOrCreateUser_DuplicateKey(t *testing.T) {
	repo := newMockUserRepository()

	// Simulate: user doesn't exist initially, then duplicate key error on create,
	// then Fetch returns existing user
	repo.MockStore.CreateUserFunc = func(ctx context.Context, userID collection.UserID) error {
		// First call - user doesn't exist, create fails with duplicate key
		// Simulating race condition: another request created user between GetUser and CreateUser
		return &pgconn.PgError{Code: "23505"} // unique_violation
	}
	// After duplicate error, GetUser should find the existing user
	repo.MockStore.GetUserFunc = func(ctx context.Context, userID collection.UserID) (collection.User, error) {
		return collection.User{
			UserID:          999,
			DiscordUsername: "concurrent_user",
			DiscordAvatar:   "avatar",
			Date:            time.Now(),
		}, nil
	}

	adapter := NewUserStoreAdapter(repo)
	got, err := adapter.GetOrCreateUser(context.Background(), 12345, "newuser", "newavatar")

	assert.NoError(t, err)
	assert.Equal(t, uint64(999), got, "should return existing user ID after duplicate key error")
}

// Test that Discord info gets updated when user exists
func TestUserStoreAdapter_GetOrCreateUser_UpdatesDiscordInfo(t *testing.T) {
	repo := newMockUserRepository()
	repo.users[12345] = collection.User{UserID: 12345, DiscordUsername: "olduser", DiscordAvatar: "oldavatar"}

	var updateCalled bool
	var updateUsername, updateAvatar string
	repo.MockStore.UpdateDiscordInfoFunc = func(_ context.Context, _ collection.UserID, username, avatar string, _ time.Time) error {
		updateCalled = true
		updateUsername = username
		updateAvatar = avatar
		return nil
	}

	adapter := NewUserStoreAdapter(repo)
	_, err := adapter.GetOrCreateUser(context.Background(), 12345, "newuser", "newavatar")

	assert.NoError(t, err)
	assert.True(t, updateCalled, "UpdateDiscordInfo should be called for existing user")
	assert.Equal(t, "newuser", updateUsername)
	assert.Equal(t, "newavatar", updateAvatar)
}

// Test that existing user returns its ID
func TestUserStoreAdapter_GetOrCreateUser_FetchesAfterCreate(t *testing.T) {
	repo := newMockUserRepository()
	// Pre-populate to simulate user exists
	repo.users[12345] = collection.User{UserID: 12345, DiscordUsername: "test", DiscordAvatar: "av"}

	adapter := NewUserStoreAdapter(repo)
	got, err := adapter.GetOrCreateUser(context.Background(), 12345, "user", "avatar")

	assert.NoError(t, err)
	assert.Equal(t, uint64(12345), got, "should return existing user ID")
}
