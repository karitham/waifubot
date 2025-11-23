package guild

import (
	"context"
	"testing"
	"time"

	"github.com/Karitham/corde"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	guildmocks "github.com/karitham/waifubot/guild/mocks"
	"github.com/karitham/waifubot/storage/guildstore"
	storemocks "github.com/karitham/waifubot/storage/mocks"
)

func TestIndexGuildIfNeeded(t *testing.T) {
	tests := []struct {
		name    string
		guildID corde.Snowflake
		setup   func(*storemocks.MockStorageStore, *guildmocks.MockGuildQuerier)
		wantErr bool
	}{
		{
			name:    "guild already indexed",
			guildID: 123,
			setup: func(ms *storemocks.MockStorageStore, mq *guildmocks.MockGuildQuerier) {
				ms.EXPECT().GuildStore().Return(mq)
				mq.EXPECT().IsGuildIndexed(gomock.Any(), uint64(123)).Return(guildstore.IsGuildIndexedRow{
					Status:    guildstore.IndexingStatusCompleted,
					UpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
				}, nil)
			},
			wantErr: false,
		},
		{
			name:    "guild not indexed, start indexing",
			guildID: 123,
			setup: func(ms *storemocks.MockStorageStore, mq *guildmocks.MockGuildQuerier) {
				ms.EXPECT().GuildStore().Return(mq)
				mq.EXPECT().IsGuildIndexed(gomock.Any(), uint64(123)).Return(guildstore.IsGuildIndexedRow{
					Status:    guildstore.IndexingStatusCompleted,
					UpdatedAt: pgtype.Timestamp{Time: time.Now().Add(-8 * 24 * time.Hour), Valid: true},
				}, nil)
				ms.EXPECT().Tx(gomock.Any()).Return(ms, nil)
				ms.EXPECT().GuildStore().Return(mq).AnyTimes()
				mq.EXPECT().IsGuildIndexed(gomock.Any(), uint64(123)).Return(guildstore.IsGuildIndexedRow{}, nil)
				mq.EXPECT().StartIndexingJob(gomock.Any(), uint64(123)).Return(nil)
				ms.EXPECT().Commit(gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "tx error",
			guildID: 123,
			setup: func(ms *storemocks.MockStorageStore, mq *guildmocks.MockGuildQuerier) {
				ms.EXPECT().GuildStore().Return(mq)
				mq.EXPECT().IsGuildIndexed(gomock.Any(), uint64(123)).Return(guildstore.IsGuildIndexedRow{
					Status:    guildstore.IndexingStatusCompleted,
					UpdatedAt: pgtype.Timestamp{Time: time.Now().Add(-8 * 24 * time.Hour), Valid: true},
				}, nil)
				ms.EXPECT().Tx(gomock.Any()).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name:    "is guild indexed error",
			guildID: 123,
			setup: func(ms *storemocks.MockStorageStore, mq *guildmocks.MockGuildQuerier) {
				ms.EXPECT().GuildStore().Return(mq)
				mq.EXPECT().IsGuildIndexed(gomock.Any(), uint64(123)).Return(guildstore.IsGuildIndexedRow{
					Status:    guildstore.IndexingStatusCompleted,
					UpdatedAt: pgtype.Timestamp{Time: time.Now().Add(-8 * 24 * time.Hour), Valid: true},
				}, nil)
				ms.EXPECT().Tx(gomock.Any()).Return(ms, nil)
				ms.EXPECT().GuildStore().Return(mq).AnyTimes()
				mq.EXPECT().IsGuildIndexed(gomock.Any(), uint64(123)).Return(guildstore.IsGuildIndexedRow{}, assert.AnError)
				mq.EXPECT().StartIndexingJob(gomock.Any(), uint64(123)).Return(assert.AnError)
				ms.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr: true,
		},
		{
			name:    "start indexing job error",
			guildID: 123,
			setup: func(ms *storemocks.MockStorageStore, mq *guildmocks.MockGuildQuerier) {
				ms.EXPECT().GuildStore().Return(mq)
				mq.EXPECT().IsGuildIndexed(gomock.Any(), uint64(123)).Return(guildstore.IsGuildIndexedRow{
					Status:    guildstore.IndexingStatusCompleted,
					UpdatedAt: pgtype.Timestamp{Time: time.Now().Add(-8 * 24 * time.Hour), Valid: true},
				}, nil)
				ms.EXPECT().Tx(gomock.Any()).Return(ms, nil)
				ms.EXPECT().GuildStore().Return(mq).AnyTimes()
				mq.EXPECT().IsGuildIndexed(gomock.Any(), uint64(123)).Return(guildstore.IsGuildIndexedRow{}, nil)
				mq.EXPECT().StartIndexingJob(gomock.Any(), uint64(123)).Return(assert.AnError)
				ms.EXPECT().Rollback(gomock.Any()).Return(nil)
			},
			wantErr: true,
		},
		{
			name:    "commit error",
			guildID: 123,
			setup: func(ms *storemocks.MockStorageStore, mq *guildmocks.MockGuildQuerier) {
				ms.EXPECT().GuildStore().Return(mq)
				mq.EXPECT().IsGuildIndexed(gomock.Any(), uint64(123)).Return(guildstore.IsGuildIndexedRow{
					Status:    guildstore.IndexingStatusCompleted,
					UpdatedAt: pgtype.Timestamp{Time: time.Now().Add(-8 * 24 * time.Hour), Valid: true},
				}, nil)
				ms.EXPECT().Tx(gomock.Any()).Return(ms, nil)
				ms.EXPECT().GuildStore().Return(mq).AnyTimes()
				mq.EXPECT().IsGuildIndexed(gomock.Any(), uint64(123)).Return(guildstore.IsGuildIndexedRow{}, nil)
				mq.EXPECT().StartIndexingJob(gomock.Any(), uint64(123)).Return(nil)
				ms.EXPECT().Commit(gomock.Any()).Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := storemocks.NewMockStorageStore(ctrl)
			querier := guildmocks.NewMockGuildQuerier(ctrl)
			tt.setup(store, querier)
			indexer := NewIndexer(store, "test_token")

			err := indexer.IndexGuildIfNeeded(context.Background(), tt.guildID)
			if (err != nil) != tt.wantErr {
				t.Errorf("IndexGuildIfNeeded() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
