package guild

import (
	"context"
	"testing"
	"time"

	"github.com/Karitham/corde"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"

	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/guildstore"
)

// MockStore is a mock implementation of the Store interface
type MockStore struct {
	TxFunc         func(ctx context.Context) (storage.Store, error)
	CommitFunc     func(ctx context.Context) error
	RollbackFunc   func(ctx context.Context) error
	GuildStoreFunc func() guildstore.Querier

	storage.Store
}

func (m *MockStore) Tx(ctx context.Context) (storage.Store, error) {
	if m.TxFunc != nil {
		return m.TxFunc(ctx)
	}
	return m, nil
}

func (m *MockStore) Commit(ctx context.Context) error {
	if m.CommitFunc != nil {
		return m.CommitFunc(ctx)
	}
	return nil
}

func (m *MockStore) Rollback(ctx context.Context) error {
	if m.RollbackFunc != nil {
		return m.RollbackFunc(ctx)
	}
	return nil
}

func (m *MockStore) GuildStore() guildstore.Querier {
	if m.GuildStoreFunc != nil {
		return m.GuildStoreFunc()
	}
	return &MockGuildStore{}
}

// MockGuildStore is a mock implementation of guilds.Querier
type MockGuildStore struct {
	GetGuildMembersFunc         func(ctx context.Context, guildID uint64) ([]int64, error)
	UsersOwningCharInGuildFunc  func(ctx context.Context, arg guildstore.UsersOwningCharInGuildParams) ([]uint64, error)
	CompleteIndexingJobFunc     func(ctx context.Context, guildID uint64) error
	DeleteGuildMembersFunc      func(ctx context.Context, guildID uint64) error
	DeleteGuildMembersNotInFunc func(ctx context.Context, arg guildstore.DeleteGuildMembersNotInParams) error
	GetIndexingStatusFunc       func(ctx context.Context, guildID uint64) (guildstore.GetIndexingStatusRow, error)
	IsGuildIndexedFunc          func(ctx context.Context, guildID uint64) (guildstore.IsGuildIndexedRow, error)
	StartIndexingJobFunc        func(ctx context.Context, guildID uint64) error
	UpsertGuildMembersFunc      func(ctx context.Context, arg guildstore.UpsertGuildMembersParams) error
}

func (m *MockGuildStore) GetGuildMembers(ctx context.Context, guildID uint64) ([]int64, error) {
	if m.GetGuildMembersFunc != nil {
		return m.GetGuildMembersFunc(ctx, guildID)
	}
	return []int64{}, nil
}

func (m *MockGuildStore) UsersOwningCharInGuild(ctx context.Context, arg guildstore.UsersOwningCharInGuildParams) ([]uint64, error) {
	if m.UsersOwningCharInGuildFunc != nil {
		return m.UsersOwningCharInGuildFunc(ctx, arg)
	}
	return []uint64{}, nil
}

func (m *MockGuildStore) CompleteIndexingJob(ctx context.Context, guildID uint64) error {
	if m.CompleteIndexingJobFunc != nil {
		return m.CompleteIndexingJobFunc(ctx, guildID)
	}
	return nil
}

func (m *MockGuildStore) DeleteGuildMembers(ctx context.Context, guildID uint64) error {
	if m.DeleteGuildMembersFunc != nil {
		return m.DeleteGuildMembersFunc(ctx, guildID)
	}
	return nil
}

func (m *MockGuildStore) DeleteGuildMembersNotIn(ctx context.Context, arg guildstore.DeleteGuildMembersNotInParams) error {
	if m.DeleteGuildMembersNotInFunc != nil {
		return m.DeleteGuildMembersNotInFunc(ctx, arg)
	}
	return nil
}

func (m *MockGuildStore) GetIndexingStatus(ctx context.Context, guildID uint64) (guildstore.GetIndexingStatusRow, error) {
	if m.GetIndexingStatusFunc != nil {
		return m.GetIndexingStatusFunc(ctx, guildID)
	}
	return guildstore.GetIndexingStatusRow{}, nil
}

func (m *MockGuildStore) IsGuildIndexed(ctx context.Context, guildID uint64) (guildstore.IsGuildIndexedRow, error) {
	if m.IsGuildIndexedFunc != nil {
		return m.IsGuildIndexedFunc(ctx, guildID)
	}
	return guildstore.IsGuildIndexedRow{}, nil
}

func (m *MockGuildStore) StartIndexingJob(ctx context.Context, guildID uint64) error {
	if m.StartIndexingJobFunc != nil {
		return m.StartIndexingJobFunc(ctx, guildID)
	}
	return nil
}

func (m *MockGuildStore) UpsertGuildMembers(ctx context.Context, arg guildstore.UpsertGuildMembersParams) error {
	if m.UpsertGuildMembersFunc != nil {
		return m.UpsertGuildMembersFunc(ctx, arg)
	}
	return nil
}

func TestNewIndexer(t *testing.T) {
	store := &MockStore{}
	botToken := "test_token"
	indexer := NewIndexer(store, botToken)

	if indexer.store != store {
		t.Errorf("NewIndexer() store = %v, want %v", indexer.store, store)
	}
	if indexer.botToken != botToken {
		t.Errorf("NewIndexer() botToken = %v, want %v", indexer.botToken, botToken)
	}
	if indexer.httpClient == nil {
		t.Error("NewIndexer() httpClient should not be nil")
	}
}

func TestIndexGuildIfNeeded(t *testing.T) {
	tests := []struct {
		name    string
		guildID corde.Snowflake
		setup   func(*MockStore)
		wantErr bool
	}{
		{
			name:    "guild already indexed",
			guildID: 123,
			setup: func(ms *MockStore) {
				ms.GuildStoreFunc = func() guildstore.Querier {
					return &MockGuildStore{
						IsGuildIndexedFunc: func(ctx context.Context, guildID uint64) (guildstore.IsGuildIndexedRow, error) {
							return guildstore.IsGuildIndexedRow{
								Status:    guildstore.IndexingStatusCompleted,
								UpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
							}, nil
						},
					}
				}
			},
			wantErr: false,
		},
		{
			name:    "guild not indexed, start indexing",
			guildID: 123,
			setup: func(ms *MockStore) {
				ms.TxFunc = func(ctx context.Context) (storage.Store, error) {
					return ms, nil
				}
				ms.GuildStoreFunc = func() guildstore.Querier {
					return &MockGuildStore{
						IsGuildIndexedFunc: func(ctx context.Context, guildID uint64) (guildstore.IsGuildIndexedRow, error) {
							return guildstore.IsGuildIndexedRow{}, nil // not indexed
						},
						StartIndexingJobFunc: func(ctx context.Context, guildID uint64) error {
							return nil
						},
					}
				}
				ms.CommitFunc = func(ctx context.Context) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:    "tx error",
			guildID: 123,
			setup: func(ms *MockStore) {
				ms.TxFunc = func(ctx context.Context) (storage.Store, error) {
					return nil, assert.AnError
				}
			},
			wantErr: true,
		},
		{
			name:    "is guild indexed error",
			guildID: 123,
			setup: func(ms *MockStore) {
				ms.TxFunc = func(ctx context.Context) (storage.Store, error) {
					return ms, nil
				}
				ms.GuildStoreFunc = func() guildstore.Querier {
					return &MockGuildStore{
						IsGuildIndexedFunc: func(ctx context.Context, guildID uint64) (guildstore.IsGuildIndexedRow, error) {
							return guildstore.IsGuildIndexedRow{}, assert.AnError
						},
					}
				}
				ms.RollbackFunc = func(ctx context.Context) error {
					return nil
				}
			},
			wantErr: true,
		},
		{
			name:    "start indexing job error",
			guildID: 123,
			setup: func(ms *MockStore) {
				ms.TxFunc = func(ctx context.Context) (storage.Store, error) {
					return ms, nil
				}
				ms.GuildStoreFunc = func() guildstore.Querier {
					return &MockGuildStore{
						IsGuildIndexedFunc: func(ctx context.Context, guildID uint64) (guildstore.IsGuildIndexedRow, error) {
							return guildstore.IsGuildIndexedRow{}, nil
						},
						StartIndexingJobFunc: func(ctx context.Context, guildID uint64) error {
							return assert.AnError
						},
					}
				}
				ms.RollbackFunc = func(ctx context.Context) error {
					return nil
				}
			},
			wantErr: true,
		},
		{
			name:    "commit error",
			guildID: 123,
			setup: func(ms *MockStore) {
				ms.TxFunc = func(ctx context.Context) (storage.Store, error) {
					return ms, nil
				}
				ms.GuildStoreFunc = func() guildstore.Querier {
					return &MockGuildStore{
						IsGuildIndexedFunc: func(ctx context.Context, guildID uint64) (guildstore.IsGuildIndexedRow, error) {
							return guildstore.IsGuildIndexedRow{}, nil
						},
						StartIndexingJobFunc: func(ctx context.Context, guildID uint64) error {
							return nil
						},
					}
				}
				ms.CommitFunc = func(ctx context.Context) error {
					return assert.AnError
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &MockStore{}
			tt.setup(store)
			indexer := NewIndexer(store, "test_token")

			err := indexer.IndexGuildIfNeeded(context.Background(), tt.guildID)
			if (err != nil) != tt.wantErr {
				t.Errorf("IndexGuildIfNeeded() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
