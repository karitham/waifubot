package collection

import (
	"context"
	"errors"
	"testing"

	"github.com/Karitham/corde"
	"go.uber.org/mock/gomock"

	"github.com/karitham/waifubot/storage/collectionstore"
	"github.com/karitham/waifubot/storage/guildstore"
)

func TestCharacterHolders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		guildID     corde.Snowflake
		charID      int64
		setupMocks  func(store *MockProfileStore, coll *MockCollectionQuerier, guild *MockGuildQuerier)
		wantName    string
		wantHolders []corde.Snowflake
		wantErr     bool
		errMsg      string
	}{
		{
			name:    "success",
			guildID: 123,
			charID:  456,
			setupMocks: func(store *MockProfileStore, coll *MockCollectionQuerier, guild *MockGuildQuerier) {
				store.EXPECT().CollectionStore().Return(coll).AnyTimes()
				store.EXPECT().GuildStore().Return(guild).AnyTimes()
				coll.EXPECT().GetByID(gomock.Any(), int64(456)).Return(collectionstore.Character{
					ID:    456,
					Name:  "TestChar",
					Image: "img",
				}, nil)
				guild.EXPECT().GetGuildMembers(gomock.Any(), uint64(123)).Return([]int64{1, 2, 3}, nil)
				guild.EXPECT().UsersOwningCharInGuild(gomock.Any(), guildstore.UsersOwningCharInGuildParams{
					CharacterID: 456,
					GuildID:     123,
				}).Return([]uint64{1, 2}, nil)
			},
			wantName:    "TestChar",
			wantHolders: []corde.Snowflake{1, 2},
			wantErr:     false,
		},
		{
			name:    "guildID 0",
			guildID: 0,
			charID:  456,
			setupMocks: func(store *MockProfileStore, coll *MockCollectionQuerier, guild *MockGuildQuerier) {
				// No mocks needed
			},
			wantErr: true,
			errMsg:  "this command can only be used in servers",
		},
		{
			name:    "GetByID error",
			guildID: 123,
			charID:  456,
			setupMocks: func(store *MockProfileStore, coll *MockCollectionQuerier, guild *MockGuildQuerier) {
				store.EXPECT().CollectionStore().Return(coll)
				coll.EXPECT().GetByID(gomock.Any(), int64(456)).Return(collectionstore.Character{}, errors.New("get error"))
			},
			wantErr: true,
			errMsg:  "no one in this server has 456",
		},
		{
			name:    "GetGuildMembers error",
			guildID: 123,
			charID:  456,
			setupMocks: func(store *MockProfileStore, coll *MockCollectionQuerier, guild *MockGuildQuerier) {
				store.EXPECT().CollectionStore().Return(coll)
				store.EXPECT().GuildStore().Return(guild)
				coll.EXPECT().GetByID(gomock.Any(), int64(456)).Return(collectionstore.Character{
					ID:    456,
					Name:  "TestChar",
					Image: "img",
				}, nil)
				guild.EXPECT().GetGuildMembers(gomock.Any(), uint64(123)).Return(nil, errors.New("members error"))
			},
			wantErr: true,
			errMsg:  "failed to fetch guild members: members error",
		},
		{
			name:    "empty members",
			guildID: 123,
			charID:  456,
			setupMocks: func(store *MockProfileStore, coll *MockCollectionQuerier, guild *MockGuildQuerier) {
				store.EXPECT().CollectionStore().Return(coll)
				store.EXPECT().GuildStore().Return(guild)
				coll.EXPECT().GetByID(gomock.Any(), int64(456)).Return(collectionstore.Character{
					ID:    456,
					Name:  "TestChar",
					Image: "img",
				}, nil)
				guild.EXPECT().GetGuildMembers(gomock.Any(), uint64(123)).Return([]int64{}, nil)
			},
			wantErr: true,
			errMsg:  "guild members not indexed yet, please try again later",
		},
		{
			name:    "UsersOwningCharInGuild error",
			guildID: 123,
			charID:  456,
			setupMocks: func(store *MockProfileStore, coll *MockCollectionQuerier, guild *MockGuildQuerier) {
				store.EXPECT().CollectionStore().Return(coll).AnyTimes()
				store.EXPECT().GuildStore().Return(guild).AnyTimes()
				coll.EXPECT().GetByID(gomock.Any(), int64(456)).Return(collectionstore.Character{
					ID:    456,
					Name:  "TestChar",
					Image: "img",
				}, nil)
				guild.EXPECT().GetGuildMembers(gomock.Any(), uint64(123)).Return([]int64{1, 2, 3}, nil)
				guild.EXPECT().UsersOwningCharInGuild(gomock.Any(), guildstore.UsersOwningCharInGuildParams{
					CharacterID: 456,
					GuildID:     123,
				}).Return(nil, errors.New("holders error"))
			},
			wantErr: true,
			errMsg:  "failed to fetch character holders: holders error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMockProfileStore(ctrl)
			coll := NewMockCollectionQuerier(ctrl)
			guild := NewMockGuildQuerier(ctrl)
			tt.setupMocks(store, coll, guild)

			gotName, gotHolders, err := CharacterHolders(context.Background(), store, tt.guildID, tt.charID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CharacterHolders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err.Error() != tt.errMsg {
					t.Errorf("CharacterHolders() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}
			if gotName != tt.wantName {
				t.Errorf("CharacterHolders() gotName = %v, want %v", gotName, tt.wantName)
			}
			if len(gotHolders) != len(tt.wantHolders) {
				t.Errorf("CharacterHolders() gotHolders len = %v, want %v", len(gotHolders), len(tt.wantHolders))
			} else {
				for i, h := range gotHolders {
					if h != tt.wantHolders[i] {
						t.Errorf("CharacterHolders() gotHolders[%d] = %v, want %v", i, h, tt.wantHolders[i])
					}
				}
			}
		})
	}
}
