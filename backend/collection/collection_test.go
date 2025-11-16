package collection

import (
	"context"
	"testing"
	"time"

	"github.com/Karitham/corde"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"

	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/collectionstore"
	"github.com/karitham/waifubot/storage/guildstore"
	"github.com/karitham/waifubot/storage/userstore"
)

// MockStore is a mock implementation of the Store interface
type MockStore struct {
	TxFunc              func(ctx context.Context) (storage.Store, error)
	CommitFunc          func(ctx context.Context) error
	RollbackFunc        func(ctx context.Context) error
	CollectionStoreFunc func() collectionstore.Querier
	UserStoreFunc       func() userstore.Querier
	GuildStoreFunc      func() guildstore.Querier
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

func (m *MockStore) CollectionStore() collectionstore.Querier {
	if m.CollectionStoreFunc != nil {
		return m.CollectionStoreFunc()
	}
	return &MockCollectionStore{}
}

func (m *MockStore) UserStore() userstore.Querier {
	if m.UserStoreFunc != nil {
		return m.UserStoreFunc()
	}
	return &MockUserStore{}
}

func (m *MockStore) GuildStore() guildstore.Querier {
	if m.GuildStoreFunc != nil {
		return m.GuildStoreFunc()
	}
	return &MockGuildStore{}
}

// MockCollectionStore is a mock implementation of characters.Querier
type MockCollectionStore struct {
	DeleteFunc                  func(ctx context.Context, arg collectionstore.DeleteParams) (collectionstore.Character, error)
	GetFunc                     func(ctx context.Context, arg collectionstore.GetParams) (collectionstore.Character, error)
	GetByIDFunc                 func(ctx context.Context, id int64) (collectionstore.Character, error)
	ListFunc                    func(ctx context.Context, userID uint64) ([]collectionstore.Character, error)
	ListIDsFunc                 func(ctx context.Context, userID uint64) ([]int64, error)
	CountFunc                   func(ctx context.Context, userID uint64) (int64, error)
	InsertFunc                  func(ctx context.Context, arg collectionstore.InsertParams) error
	GiveFunc                    func(ctx context.Context, arg collectionstore.GiveParams) (collectionstore.Character, error)
	SearchCharactersFunc        func(ctx context.Context, arg collectionstore.SearchCharactersParams) ([]collectionstore.Character, error)
	SearchGlobalCharactersFunc  func(ctx context.Context, arg collectionstore.SearchGlobalCharactersParams) ([]collectionstore.SearchGlobalCharactersRow, error)
	UpdateImageNameFunc         func(ctx context.Context, arg collectionstore.UpdateImageNameParams) (collectionstore.Character, error)
	UsersOwningCharFilteredFunc func(ctx context.Context, arg collectionstore.UsersOwningCharFilteredParams) ([]uint64, error)
}

func (m *MockCollectionStore) Delete(ctx context.Context, arg collectionstore.DeleteParams) (collectionstore.Character, error) {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, arg)
	}
	return collectionstore.Character{}, nil
}

func (m *MockCollectionStore) Get(ctx context.Context, arg collectionstore.GetParams) (collectionstore.Character, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, arg)
	}
	return collectionstore.Character{}, nil
}

func (m *MockCollectionStore) GetByID(ctx context.Context, id int64) (collectionstore.Character, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return collectionstore.Character{}, nil
}

func (m *MockCollectionStore) List(ctx context.Context, userID uint64) ([]collectionstore.Character, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, userID)
	}
	return []collectionstore.Character{}, nil
}

func (m *MockCollectionStore) ListIDs(ctx context.Context, userID uint64) ([]int64, error) {
	if m.ListIDsFunc != nil {
		return m.ListIDsFunc(ctx, userID)
	}
	return []int64{}, nil
}

func (m *MockCollectionStore) Count(ctx context.Context, userID uint64) (int64, error) {
	if m.CountFunc != nil {
		return m.CountFunc(ctx, userID)
	}
	return 0, nil
}

func (m *MockCollectionStore) Insert(ctx context.Context, arg collectionstore.InsertParams) error {
	if m.InsertFunc != nil {
		return m.InsertFunc(ctx, arg)
	}
	return nil
}

func (m *MockCollectionStore) Give(ctx context.Context, arg collectionstore.GiveParams) (collectionstore.Character, error) {
	if m.GiveFunc != nil {
		return m.GiveFunc(ctx, arg)
	}
	return collectionstore.Character{}, nil
}

func (m *MockCollectionStore) SearchCharacters(ctx context.Context, arg collectionstore.SearchCharactersParams) ([]collectionstore.Character, error) {
	if m.SearchCharactersFunc != nil {
		return m.SearchCharactersFunc(ctx, arg)
	}
	return []collectionstore.Character{}, nil
}

func (m *MockCollectionStore) SearchGlobalCharacters(ctx context.Context, arg collectionstore.SearchGlobalCharactersParams) ([]collectionstore.SearchGlobalCharactersRow, error) {
	if m.SearchGlobalCharactersFunc != nil {
		return m.SearchGlobalCharactersFunc(ctx, arg)
	}
	return []collectionstore.SearchGlobalCharactersRow{}, nil
}

func (m *MockCollectionStore) UpdateImageName(ctx context.Context, arg collectionstore.UpdateImageNameParams) (collectionstore.Character, error) {
	if m.UpdateImageNameFunc != nil {
		return m.UpdateImageNameFunc(ctx, arg)
	}
	return collectionstore.Character{}, nil
}

func (m *MockCollectionStore) UsersOwningCharFiltered(ctx context.Context, arg collectionstore.UsersOwningCharFilteredParams) ([]uint64, error) {
	if m.UsersOwningCharFilteredFunc != nil {
		return m.UsersOwningCharFilteredFunc(ctx, arg)
	}
	return []uint64{}, nil
}

// MockUserStore is a mock implementation of users.Querier
type MockUserStore struct {
	GetFunc              func(ctx context.Context, userID uint64) (userstore.User, error)
	IncTokensFunc        func(ctx context.Context, userID uint64) error
	UpdateDateFunc       func(ctx context.Context, arg userstore.UpdateDateParams) error
	ConsumeTokensFunc    func(ctx context.Context, arg userstore.ConsumeTokensParams) (userstore.User, error)
	UpdateFavoriteFunc   func(ctx context.Context, arg userstore.UpdateFavoriteParams) error
	UpdateAnilistURLFunc func(ctx context.Context, arg userstore.UpdateAnilistURLParams) error
	UpdateQuoteFunc      func(ctx context.Context, arg userstore.UpdateQuoteParams) error
	CreateFunc           func(ctx context.Context, userID uint64) error
	GetByAnilistFunc     func(ctx context.Context, lower string) (userstore.User, error)
}

func (m *MockUserStore) Get(ctx context.Context, userID uint64) (userstore.User, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, userID)
	}
	return userstore.User{}, nil
}

func (m *MockUserStore) IncTokens(ctx context.Context, userID uint64) error {
	if m.IncTokensFunc != nil {
		return m.IncTokensFunc(ctx, userID)
	}
	return nil
}

func (m *MockUserStore) UpdateDate(ctx context.Context, arg userstore.UpdateDateParams) error {
	if m.UpdateDateFunc != nil {
		return m.UpdateDateFunc(ctx, arg)
	}
	return nil
}

func (m *MockUserStore) ConsumeTokens(ctx context.Context, arg userstore.ConsumeTokensParams) (userstore.User, error) {
	if m.ConsumeTokensFunc != nil {
		return m.ConsumeTokensFunc(ctx, arg)
	}
	return userstore.User{}, nil
}

func (m *MockUserStore) UpdateFavorite(ctx context.Context, arg userstore.UpdateFavoriteParams) error {
	if m.UpdateFavoriteFunc != nil {
		return m.UpdateFavoriteFunc(ctx, arg)
	}
	return nil
}

func (m *MockUserStore) UpdateAnilistURL(ctx context.Context, arg userstore.UpdateAnilistURLParams) error {
	if m.UpdateAnilistURLFunc != nil {
		return m.UpdateAnilistURLFunc(ctx, arg)
	}
	return nil
}

func (m *MockUserStore) UpdateQuote(ctx context.Context, arg userstore.UpdateQuoteParams) error {
	if m.UpdateQuoteFunc != nil {
		return m.UpdateQuoteFunc(ctx, arg)
	}
	return nil
}

func (m *MockUserStore) Create(ctx context.Context, userID uint64) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, userID)
	}
	return nil
}

func (m *MockUserStore) GetByAnilist(ctx context.Context, lower string) (userstore.User, error) {
	if m.GetByAnilistFunc != nil {
		return m.GetByAnilistFunc(ctx, lower)
	}
	return userstore.User{}, nil
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

// MockAnimeService is a mock implementation of AnimeService
type MockAnimeService struct {
	RandomCharFunc func(ctx context.Context, notIn ...int64) (MediaCharacter, error)
}

func (m *MockAnimeService) RandomChar(ctx context.Context, notIn ...int64) (MediaCharacter, error) {
	if m.RandomCharFunc != nil {
		return m.RandomCharFunc(ctx, notIn...)
	}
	return MediaCharacter{}, nil
}

func (m *MockAnimeService) Anime(ctx context.Context, name string) ([]Media, error) {
	return []Media{}, nil
}

func (m *MockAnimeService) Manga(ctx context.Context, name string) ([]Media, error) {
	return []Media{}, nil
}

func (m *MockAnimeService) User(ctx context.Context, name string) ([]TrackerUser, error) {
	return []TrackerUser{}, nil
}

func (m *MockAnimeService) Character(ctx context.Context, name string) ([]MediaCharacter, error) {
	return []MediaCharacter{}, nil
}

func TestExchange(t *testing.T) {
	tests := []struct {
		name    string
		userID  corde.Snowflake
		charID  int64
		setup   func(*MockStore)
		want    collectionstore.Character
		wantErr bool
	}{
		{
			name:   "successful exchange",
			userID: 123,
			charID: 456,
			setup: func(ms *MockStore) {
				ms.TxFunc = func(ctx context.Context) (storage.Store, error) {
					return ms, nil
				}
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						DeleteFunc: func(ctx context.Context, arg collectionstore.DeleteParams) (collectionstore.Character, error) {
							return collectionstore.Character{
								ID:     456,
								UserID: 123,
								Name:   "Test Char",
								Image:  "test.jpg",
								Type:   "ROLL",
								Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
							}, nil
						},
					}
				}
				ms.UserStoreFunc = func() userstore.Querier {
					return &MockUserStore{
						IncTokensFunc: func(ctx context.Context, userID uint64) error {
							return nil
						},
					}
				}
				ms.CommitFunc = func(ctx context.Context) error {
					return nil
				}
			},
			want: collectionstore.Character{
				ID:     456,
				UserID: 123,
				Name:   "Test Char",
				Image:  "test.jpg",
				Type:   "ROLL",
				Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
			},
			wantErr: false,
		},
		{
			name:   "tx error",
			userID: 123,
			charID: 456,
			setup: func(ms *MockStore) {
				ms.TxFunc = func(ctx context.Context) (storage.Store, error) {
					return nil, assert.AnError
				}
			},
			want:    collectionstore.Character{},
			wantErr: true,
		},
		{
			name:   "delete error",
			userID: 123,
			charID: 456,
			setup: func(ms *MockStore) {
				ms.TxFunc = func(ctx context.Context) (storage.Store, error) {
					return ms, nil
				}
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						DeleteFunc: func(ctx context.Context, arg collectionstore.DeleteParams) (collectionstore.Character, error) {
							return collectionstore.Character{}, assert.AnError
						},
					}
				}
				ms.RollbackFunc = func(ctx context.Context) error {
					return nil
				}
			},
			want:    collectionstore.Character{},
			wantErr: true,
		},
		{
			name:   "inc tokens error",
			userID: 123,
			charID: 456,
			setup: func(ms *MockStore) {
				ms.TxFunc = func(ctx context.Context) (storage.Store, error) {
					return ms, nil
				}
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						DeleteFunc: func(ctx context.Context, arg collectionstore.DeleteParams) (collectionstore.Character, error) {
							return collectionstore.Character{
								ID:     456,
								UserID: 123,
								Name:   "Test Char",
								Image:  "test.jpg",
								Type:   "ROLL",
								Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
							}, nil
						},
					}
				}
				ms.UserStoreFunc = func() userstore.Querier {
					return &MockUserStore{
						IncTokensFunc: func(ctx context.Context, userID uint64) error {
							return assert.AnError
						},
					}
				}
				ms.RollbackFunc = func(ctx context.Context) error {
					return nil
				}
			},
			want:    collectionstore.Character{},
			wantErr: true,
		},
		{
			name:   "commit error",
			userID: 123,
			charID: 456,
			setup: func(ms *MockStore) {
				ms.TxFunc = func(ctx context.Context) (storage.Store, error) {
					return ms, nil
				}
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						DeleteFunc: func(ctx context.Context, arg collectionstore.DeleteParams) (collectionstore.Character, error) {
							return collectionstore.Character{
								ID:     456,
								UserID: 123,
								Name:   "Test Char",
								Image:  "test.jpg",
								Type:   "ROLL",
								Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
							}, nil
						},
					}
				}
				ms.UserStoreFunc = func() userstore.Querier {
					return &MockUserStore{
						IncTokensFunc: func(ctx context.Context, userID uint64) error {
							return nil
						},
					}
				}
				ms.CommitFunc = func(ctx context.Context) error {
					return assert.AnError
				}
			},
			want:    collectionstore.Character{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &MockStore{}
			tt.setup(store)

			got, err := Exchange(context.Background(), store, tt.userID, tt.charID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exchange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.ID != tt.want.ID || got.UserID != tt.want.UserID || got.Name != tt.want.Name || got.Image != tt.want.Image || got.Type != tt.want.Type {
					t.Errorf("Exchange() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestGive(t *testing.T) {
	tests := []struct {
		name    string
		from    corde.Snowflake
		to      corde.Snowflake
		charID  int64
		setup   func(*MockStore)
		want    Character
		wantErr bool
	}{
		{
			name:   "successful give",
			from:   123,
			to:     456,
			charID: 789,
			setup: func(ms *MockStore) {
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						GetFunc: func(ctx context.Context, arg collectionstore.GetParams) (collectionstore.Character, error) {
							if arg.UserID == 123 {
								return collectionstore.Character{
									ID:     789,
									UserID: 123,
									Name:   "Test Char",
									Image:  "test.jpg",
									Type:   "ROLL",
									Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
								}, nil
							}
							return collectionstore.Character{}, assert.AnError
						},
						GiveFunc: func(ctx context.Context, arg collectionstore.GiveParams) (collectionstore.Character, error) {
							return collectionstore.Character{
								ID:     789,
								UserID: 456,
								Name:   "Test Char",
								Image:  "test.jpg",
								Type:   "ROLL",
								Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
							}, nil
						},
					}
				}
			},
			want: Character{
				ID:     789,
				UserID: 123,
				Name:   "Test Char",
				Image:  "test.jpg",
				Type:   "ROLL",
				Date:   time.Now(),
			},
			wantErr: false,
		},
		{
			name:   "from user does not own char",
			from:   123,
			to:     456,
			charID: 789,
			setup: func(ms *MockStore) {
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						GetFunc: func(ctx context.Context, arg collectionstore.GetParams) (collectionstore.Character, error) {
							return collectionstore.Character{}, assert.AnError
						},
					}
				}
			},
			want:    Character{},
			wantErr: true,
		},
		{
			name:   "to user already owns char",
			from:   123,
			to:     456,
			charID: 789,
			setup: func(ms *MockStore) {
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						GetFunc: func(ctx context.Context, arg collectionstore.GetParams) (collectionstore.Character, error) {
							if arg.UserID == 123 {
								return collectionstore.Character{
									ID:     789,
									UserID: 123,
									Name:   "Test Char",
									Image:  "test.jpg",
									Type:   "ROLL",
									Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
								}, nil
							}
							return collectionstore.Character{
								ID:     789,
								UserID: 456,
								Name:   "Test Char",
								Image:  "test.jpg",
								Type:   "ROLL",
								Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
							}, nil
						},
					}
				}
			},
			want:    Character{},
			wantErr: true,
		},
		{
			name:   "give error",
			from:   123,
			to:     456,
			charID: 789,
			setup: func(ms *MockStore) {
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						GetFunc: func(ctx context.Context, arg collectionstore.GetParams) (collectionstore.Character, error) {
							if arg.UserID == 123 {
								return collectionstore.Character{
									ID:     789,
									UserID: 123,
									Name:   "Test Char",
									Image:  "test.jpg",
									Type:   "ROLL",
									Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
								}, nil
							}
							return collectionstore.Character{}, assert.AnError
						},
						GiveFunc: func(ctx context.Context, arg collectionstore.GiveParams) (collectionstore.Character, error) {
							return collectionstore.Character{}, assert.AnError
						},
					}
				}
			},
			want:    Character{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &MockStore{}
			tt.setup(store)

			got, err := Give(context.Background(), store, tt.from, tt.to, tt.charID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Give() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.ID != tt.want.ID || got.UserID != tt.want.UserID || got.Name != tt.want.Name || got.Image != tt.want.Image || got.Type != tt.want.Type {
					t.Errorf("Give() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestCharacterHolders(t *testing.T) {
	tests := []struct {
		name     string
		guildID  corde.Snowflake
		charID   int64
		setup    func(*MockStore)
		wantName string
		wantIDs  []corde.Snowflake
		wantErr  bool
	}{
		{
			name:    "successful holders",
			guildID: 111,
			charID:  222,
			setup: func(ms *MockStore) {
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						GetByIDFunc: func(ctx context.Context, id int64) (collectionstore.Character, error) {
							return collectionstore.Character{Name: "Test Char"}, nil
						},
					}
				}
				ms.GuildStoreFunc = func() guildstore.Querier {
					return &MockGuildStore{
						GetGuildMembersFunc: func(ctx context.Context, guildID uint64) ([]int64, error) {
							return []int64{333, 444}, nil
						},
						UsersOwningCharInGuildFunc: func(ctx context.Context, arg guildstore.UsersOwningCharInGuildParams) ([]uint64, error) {
							return []uint64{333}, nil
						},
					}
				}
			},
			wantName: "Test Char",
			wantIDs:  []corde.Snowflake{333},
			wantErr:  false,
		},
		{
			name:    "guildID zero",
			guildID: 0,
			charID:  222,
			setup:   func(ms *MockStore) {},
			wantErr: true,
		},
		{
			name:    "no one has char",
			guildID: 111,
			charID:  222,
			setup: func(ms *MockStore) {
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						GetByIDFunc: func(ctx context.Context, id int64) (collectionstore.Character, error) {
							return collectionstore.Character{}, assert.AnError
						},
					}
				}
			},
			wantErr: true,
		},
		{
			name:    "guild members not indexed",
			guildID: 111,
			charID:  222,
			setup: func(ms *MockStore) {
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						GetByIDFunc: func(ctx context.Context, id int64) (collectionstore.Character, error) {
							return collectionstore.Character{Name: "Test Char"}, nil
						},
					}
				}
				ms.GuildStoreFunc = func() guildstore.Querier {
					return &MockGuildStore{
						GetGuildMembersFunc: func(ctx context.Context, guildID uint64) ([]int64, error) {
							return []int64{}, nil
						},
					}
				}
			},
			wantErr: true,
		},
		{
			name:    "get guild members error",
			guildID: 111,
			charID:  222,
			setup: func(ms *MockStore) {
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						GetByIDFunc: func(ctx context.Context, id int64) (collectionstore.Character, error) {
							return collectionstore.Character{Name: "Test Char"}, nil
						},
					}
				}
				ms.GuildStoreFunc = func() guildstore.Querier {
					return &MockGuildStore{
						GetGuildMembersFunc: func(ctx context.Context, guildID uint64) ([]int64, error) {
							return nil, assert.AnError
						},
					}
				}
			},
			wantErr: true,
		},
		{
			name:    "get holders error",
			guildID: 111,
			charID:  222,
			setup: func(ms *MockStore) {
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						GetByIDFunc: func(ctx context.Context, id int64) (collectionstore.Character, error) {
							return collectionstore.Character{Name: "Test Char"}, nil
						},
					}
				}
				ms.GuildStoreFunc = func() guildstore.Querier {
					return &MockGuildStore{
						GetGuildMembersFunc: func(ctx context.Context, guildID uint64) ([]int64, error) {
							return []int64{333, 444}, nil
						},
						UsersOwningCharInGuildFunc: func(ctx context.Context, arg guildstore.UsersOwningCharInGuildParams) ([]uint64, error) {
							return nil, assert.AnError
						},
					}
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &MockStore{}
			tt.setup(store)

			gotName, gotIDs, err := CharacterHolders(context.Background(), store, tt.guildID, tt.charID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CharacterHolders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotName != tt.wantName {
					t.Errorf("CharacterHolders() gotName = %v, want %v", gotName, tt.wantName)
				}
				if len(gotIDs) != len(tt.wantIDs) {
					t.Errorf("CharacterHolders() gotIDs len = %v, want %v", len(gotIDs), len(tt.wantIDs))
				}
				for i, id := range gotIDs {
					if i < len(tt.wantIDs) && id != tt.wantIDs[i] {
						t.Errorf("CharacterHolders() gotIDs[%d] = %v, want %v", i, id, tt.wantIDs[i])
					}
				}
			}
		})
	}
}

func TestUserProfile(t *testing.T) {
	tests := []struct {
		name    string
		userID  corde.Snowflake
		setup   func(*MockStore)
		want    Profile
		wantErr bool
	}{
		{
			name:   "successful profile with favorite",
			userID: 123,
			setup: func(ms *MockStore) {
				ms.UserStoreFunc = func() userstore.Querier {
					return &MockUserStore{
						GetFunc: func(ctx context.Context, userID uint64) (userstore.User, error) {
							return userstore.User{
								UserID:     123,
								Date:       pgtype.Timestamp{Time: time.Now(), Valid: true},
								Quote:      "Test quote",
								Favorite:   pgtype.Int8{Int64: 456, Valid: true},
								Tokens:     10,
								AnilistUrl: "https://anilist.co/user/test",
							}, nil
						},
					}
				}
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						GetFunc: func(ctx context.Context, arg collectionstore.GetParams) (collectionstore.Character, error) {
							return collectionstore.Character{
								ID:     456,
								UserID: 123,
								Name:   "Fav Char",
								Image:  "fav.jpg",
								Type:   "ROLL",
								Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
							}, nil
						},
						CountFunc: func(ctx context.Context, userID uint64) (int64, error) {
							return 5, nil
						},
					}
				}
			},
			want: Profile{
				User: User{
					UserID:     123,
					Date:       time.Now(),
					Quote:      "Test quote",
					Favorite:   456,
					Tokens:     10,
					AnilistURL: "https://anilist.co/user/test",
				},
				Favorite: Character{
					ID:     456,
					UserID: 123,
					Name:   "Fav Char",
					Image:  "fav.jpg",
					Type:   "ROLL",
					Date:   time.Now(),
				},
				CharacterCount: 5,
			},
			wantErr: false,
		},
		{
			name:   "successful profile without favorite",
			userID: 123,
			setup: func(ms *MockStore) {
				ms.UserStoreFunc = func() userstore.Querier {
					return &MockUserStore{
						GetFunc: func(ctx context.Context, userID uint64) (userstore.User, error) {
							return userstore.User{
								UserID: 123,
								Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
								Quote:  "Test quote",
								Tokens: 10,
							}, nil
						},
					}
				}
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						CountFunc: func(ctx context.Context, userID uint64) (int64, error) {
							return 3, nil
						},
					}
				}
			},
			want: Profile{
				User: User{
					UserID: 123,
					Date:   time.Now(),
					Quote:  "Test quote",
					Tokens: 10,
				},
				CharacterCount: 3,
			},
			wantErr: false,
		},
		{
			name:   "get user error",
			userID: 123,
			setup: func(ms *MockStore) {
				ms.UserStoreFunc = func() userstore.Querier {
					return &MockUserStore{
						GetFunc: func(ctx context.Context, userID uint64) (userstore.User, error) {
							return userstore.User{}, assert.AnError
						},
					}
				}
			},
			want:    Profile{},
			wantErr: true,
		},
		{
			name:   "count error",
			userID: 123,
			setup: func(ms *MockStore) {
				ms.UserStoreFunc = func() userstore.Querier {
					return &MockUserStore{
						GetFunc: func(ctx context.Context, userID uint64) (userstore.User, error) {
							return userstore.User{
								UserID: 123,
								Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
								Quote:  "Test quote",
								Tokens: 10,
							}, nil
						},
					}
				}
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						CountFunc: func(ctx context.Context, userID uint64) (int64, error) {
							return 0, assert.AnError
						},
					}
				}
			},
			want:    Profile{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &MockStore{}
			tt.setup(store)

			got, err := UserProfile(context.Background(), store, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Compare fields, ignoring time differences
				if got.UserID != tt.want.UserID || got.Quote != tt.want.Quote || got.User.Favorite != tt.want.User.Favorite || got.Tokens != tt.want.Tokens || got.AnilistURL != tt.want.AnilistURL || got.CharacterCount != tt.want.CharacterCount {
					t.Errorf("UserProfile() user fields mismatch")
				}
				if tt.want.Favorite.ID != 0 {
					if got.Favorite.ID != tt.want.Favorite.ID || got.Favorite.UserID != tt.want.Favorite.UserID || got.Favorite.Name != tt.want.Favorite.Name || got.Favorite.Image != tt.want.Favorite.Image || got.Favorite.Type != tt.want.Favorite.Type {
						t.Errorf("UserProfile() favorite mismatch")
					}
				}
			}
		})
	}
}

func TestSetFavorite(t *testing.T) {
	tests := []struct {
		name    string
		userID  corde.Snowflake
		charID  int64
		setup   func(*MockStore)
		wantErr bool
	}{
		{
			name:   "successful set favorite",
			userID: 123,
			charID: 456,
			setup: func(ms *MockStore) {
				ms.UserStoreFunc = func() userstore.Querier {
					return &MockUserStore{
						UpdateFavoriteFunc: func(ctx context.Context, arg userstore.UpdateFavoriteParams) error {
							return nil
						},
					}
				}
			},
			wantErr: false,
		},
		{
			name:   "update favorite error",
			userID: 123,
			charID: 456,
			setup: func(ms *MockStore) {
				ms.UserStoreFunc = func() userstore.Querier {
					return &MockUserStore{
						UpdateFavoriteFunc: func(ctx context.Context, arg userstore.UpdateFavoriteParams) error {
							return assert.AnError
						},
					}
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &MockStore{}
			tt.setup(store)

			err := SetFavorite(context.Background(), store, tt.userID, tt.charID)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetFavorite() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetAnilistURL(t *testing.T) {
	tests := []struct {
		name    string
		userID  corde.Snowflake
		url     string
		setup   func(*MockStore)
		wantErr bool
	}{
		{
			name:   "successful set anilist URL",
			userID: 123,
			url:    "https://anilist.co/user/testuser",
			setup: func(ms *MockStore) {
				ms.UserStoreFunc = func() userstore.Querier {
					return &MockUserStore{
						UpdateAnilistURLFunc: func(ctx context.Context, arg userstore.UpdateAnilistURLParams) error {
							return nil
						},
					}
				}
			},
			wantErr: false,
		},
		{
			name:    "invalid URL",
			userID:  123,
			url:     "not-a-url",
			setup:   func(ms *MockStore) {},
			wantErr: true,
		},
		{
			name:    "wrong host",
			userID:  123,
			url:     "https://example.com/user/test",
			setup:   func(ms *MockStore) {},
			wantErr: true,
		},
		{
			name:    "wrong path",
			userID:  123,
			url:     "https://anilist.co/profile/test",
			setup:   func(ms *MockStore) {},
			wantErr: true,
		},
		{
			name:   "update error",
			userID: 123,
			url:    "https://anilist.co/user/testuser",
			setup: func(ms *MockStore) {
				ms.UserStoreFunc = func() userstore.Querier {
					return &MockUserStore{
						UpdateAnilistURLFunc: func(ctx context.Context, arg userstore.UpdateAnilistURLParams) error {
							return assert.AnError
						},
					}
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &MockStore{}
			tt.setup(store)

			err := SetAnilistURL(context.Background(), store, tt.userID, tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetAnilistURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetQuote(t *testing.T) {
	tests := []struct {
		name    string
		userID  corde.Snowflake
		quote   string
		setup   func(*MockStore)
		wantErr bool
	}{
		{
			name:   "successful set quote",
			userID: 123,
			quote:  "This is a test quote",
			setup: func(ms *MockStore) {
				ms.UserStoreFunc = func() userstore.Querier {
					return &MockUserStore{
						UpdateQuoteFunc: func(ctx context.Context, arg userstore.UpdateQuoteParams) error {
							return nil
						},
					}
				}
			},
			wantErr: false,
		},
		{
			name:    "quote too long",
			userID:  123,
			quote:   string(make([]byte, 1025)), // 1025 chars
			setup:   func(ms *MockStore) {},
			wantErr: true,
		},
		{
			name:   "update error",
			userID: 123,
			quote:  "Test quote",
			setup: func(ms *MockStore) {
				ms.UserStoreFunc = func() userstore.Querier {
					return &MockUserStore{
						UpdateQuoteFunc: func(ctx context.Context, arg userstore.UpdateQuoteParams) error {
							return assert.AnError
						},
					}
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &MockStore{}
			tt.setup(store)

			err := SetQuote(context.Background(), store, tt.userID, tt.quote)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetQuote() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSearchCharacters(t *testing.T) {
	tests := []struct {
		name    string
		userID  corde.Snowflake
		term    string
		setup   func(*MockStore)
		want    []collectionstore.Character
		wantErr bool
	}{
		{
			name:   "successful search",
			userID: 123,
			term:   "test",
			setup: func(ms *MockStore) {
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						SearchCharactersFunc: func(ctx context.Context, arg collectionstore.SearchCharactersParams) ([]collectionstore.Character, error) {
							return []collectionstore.Character{
								{
									ID:     456,
									UserID: 123,
									Name:   "Test Char",
									Image:  "test.jpg",
									Type:   "ROLL",
									Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
								},
							}, nil
						},
					}
				}
			},
			want: []collectionstore.Character{
				{
					ID:     456,
					UserID: 123,
					Name:   "Test Char",
					Image:  "test.jpg",
					Type:   "ROLL",
					Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
				},
			},
			wantErr: false,
		},
		{
			name:   "search error",
			userID: 123,
			term:   "test",
			setup: func(ms *MockStore) {
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						SearchCharactersFunc: func(ctx context.Context, arg collectionstore.SearchCharactersParams) ([]collectionstore.Character, error) {
							return nil, assert.AnError
						},
					}
				}
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &MockStore{}
			tt.setup(store)

			got, err := SearchCharacters(context.Background(), store, tt.userID, tt.term)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchCharacters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("SearchCharacters() len = %v, want %v", len(got), len(tt.want))
				}
				for i, g := range got {
					if i < len(tt.want) {
						w := tt.want[i]
						if g.ID != w.ID || g.UserID != w.UserID || g.Name != w.Name || g.Image != w.Image || g.Type != w.Type {
							t.Errorf("SearchCharacters()[%d] = %v, want %v", i, g, w)
						}
					}
				}
			}
		})
	}
}

func TestRoll(t *testing.T) {
	tests := []struct {
		name    string
		userID  corde.Snowflake
		config  Config
		setup   func(*MockStore, *MockAnimeService)
		want    MediaCharacter
		wantErr bool
	}{
		{
			name:   "successful roll with cooldown",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 5},
			setup: func(ms *MockStore, mas *MockAnimeService) {
				ms.TxFunc = func(ctx context.Context) (storage.Store, error) {
					return ms, nil
				}
				ms.UserStoreFunc = func() userstore.Querier {
					return &MockUserStore{
						GetFunc: func(ctx context.Context, userID uint64) (userstore.User, error) {
							return userstore.User{
								UserID: 123,
								Date:   pgtype.Timestamp{Time: time.Now().Add(-2 * time.Hour), Valid: true},
								Tokens: 3,
							}, nil
						},
						UpdateDateFunc: func(ctx context.Context, arg userstore.UpdateDateParams) error {
							return nil
						},
					}
				}
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						ListIDsFunc: func(ctx context.Context, userID uint64) ([]int64, error) {
							return []int64{456}, nil
						},
						InsertFunc: func(ctx context.Context, arg collectionstore.InsertParams) error {
							return nil
						},
					}
				}
				mas.RandomCharFunc = func(ctx context.Context, notIn ...int64) (MediaCharacter, error) {
					return MediaCharacter{
						ID:          789,
						Name:        "New Char",
						ImageURL:    "new.jpg",
						URL:         "url",
						Description: "desc",
						MediaTitle:  "Anime",
					}, nil
				}
				ms.CommitFunc = func(ctx context.Context) error {
					return nil
				}
			},
			want: MediaCharacter{
				ID:          789,
				Name:        "New Char",
				ImageURL:    "new.jpg",
				URL:         "url",
				Description: "desc",
				MediaTitle:  "Anime",
			},
			wantErr: false,
		},
		{
			name:   "successful roll with tokens",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 5},
			setup: func(ms *MockStore, mas *MockAnimeService) {
				ms.TxFunc = func(ctx context.Context) (storage.Store, error) {
					return ms, nil
				}
				ms.UserStoreFunc = func() userstore.Querier {
					return &MockUserStore{
						GetFunc: func(ctx context.Context, userID uint64) (userstore.User, error) {
							return userstore.User{
								UserID: 123,
								Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
								Tokens: 10,
							}, nil
						},
						ConsumeTokensFunc: func(ctx context.Context, arg userstore.ConsumeTokensParams) (userstore.User, error) {
							return userstore.User{}, nil
						},
					}
				}
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						ListIDsFunc: func(ctx context.Context, userID uint64) ([]int64, error) {
							return []int64{}, nil
						},
						InsertFunc: func(ctx context.Context, arg collectionstore.InsertParams) error {
							return nil
						},
					}
				}
				mas.RandomCharFunc = func(ctx context.Context, notIn ...int64) (MediaCharacter, error) {
					return MediaCharacter{
						ID:          789,
						Name:        "New Char",
						ImageURL:    "new.jpg",
						URL:         "url",
						Description: "desc",
						MediaTitle:  "Anime",
					}, nil
				}
				ms.CommitFunc = func(ctx context.Context) error {
					return nil
				}
			},
			want: MediaCharacter{
				ID:          789,
				Name:        "New Char",
				ImageURL:    "new.jpg",
				URL:         "url",
				Description: "desc",
				MediaTitle:  "Anime",
			},
			wantErr: false,
		},
		{
			name:   "insufficient tokens",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 5},
			setup: func(ms *MockStore, mas *MockAnimeService) {
				ms.TxFunc = func(ctx context.Context) (storage.Store, error) {
					return ms, nil
				}
				ms.UserStoreFunc = func() userstore.Querier {
					return &MockUserStore{
						GetFunc: func(ctx context.Context, userID uint64) (userstore.User, error) {
							return userstore.User{
								UserID: 123,
								Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
								Tokens: 3,
							}, nil
						},
					}
				}
				ms.RollbackFunc = func(ctx context.Context) error {
					return nil
				}
			},
			want:    MediaCharacter{},
			wantErr: true,
		},
		{
			name:   "tx error",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 5},
			setup: func(ms *MockStore, mas *MockAnimeService) {
				ms.TxFunc = func(ctx context.Context) (storage.Store, error) {
					return nil, assert.AnError
				}
			},
			want:    MediaCharacter{},
			wantErr: true,
		},
		{
			name:   "get user error",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 5},
			setup: func(ms *MockStore, mas *MockAnimeService) {
				ms.TxFunc = func(ctx context.Context) (storage.Store, error) {
					return ms, nil
				}
				ms.UserStoreFunc = func() userstore.Querier {
					return &MockUserStore{
						GetFunc: func(ctx context.Context, userID uint64) (userstore.User, error) {
							return userstore.User{}, assert.AnError
						},
					}
				}
				ms.RollbackFunc = func(ctx context.Context) error {
					return nil
				}
			},
			want:    MediaCharacter{},
			wantErr: true,
		},
		{
			name:   "random char error",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 5},
			setup: func(ms *MockStore, mas *MockAnimeService) {
				ms.TxFunc = func(ctx context.Context) (storage.Store, error) {
					return ms, nil
				}
				ms.UserStoreFunc = func() userstore.Querier {
					return &MockUserStore{
						GetFunc: func(ctx context.Context, userID uint64) (userstore.User, error) {
							return userstore.User{
								UserID: 123,
								Date:   pgtype.Timestamp{Time: time.Now().Add(-2 * time.Hour), Valid: true},
								Tokens: 3,
							}, nil
						},
					}
				}
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						ListIDsFunc: func(ctx context.Context, userID uint64) ([]int64, error) {
							return []int64{}, nil
						},
					}
				}
				mas.RandomCharFunc = func(ctx context.Context, notIn ...int64) (MediaCharacter, error) {
					return MediaCharacter{}, assert.AnError
				}
				ms.RollbackFunc = func(ctx context.Context) error {
					return nil
				}
			},
			want:    MediaCharacter{},
			wantErr: true,
		},
		{
			name:   "commit error",
			userID: 123,
			config: Config{RollCooldown: time.Hour, TokensNeeded: 5},
			setup: func(ms *MockStore, mas *MockAnimeService) {
				ms.TxFunc = func(ctx context.Context) (storage.Store, error) {
					return ms, nil
				}
				ms.UserStoreFunc = func() userstore.Querier {
					return &MockUserStore{
						GetFunc: func(ctx context.Context, userID uint64) (userstore.User, error) {
							return userstore.User{
								UserID: 123,
								Date:   pgtype.Timestamp{Time: time.Now().Add(-2 * time.Hour), Valid: true},
								Tokens: 3,
							}, nil
						},
						UpdateDateFunc: func(ctx context.Context, arg userstore.UpdateDateParams) error {
							return nil
						},
					}
				}
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						ListIDsFunc: func(ctx context.Context, userID uint64) ([]int64, error) {
							return []int64{}, nil
						},
						InsertFunc: func(ctx context.Context, arg collectionstore.InsertParams) error {
							return nil
						},
					}
				}
				mas.RandomCharFunc = func(ctx context.Context, notIn ...int64) (MediaCharacter, error) {
					return MediaCharacter{
						ID:          789,
						Name:        "New Char",
						ImageURL:    "new.jpg",
						URL:         "url",
						Description: "desc",
						MediaTitle:  "Anime",
					}, nil
				}
				ms.CommitFunc = func(ctx context.Context) error {
					return assert.AnError
				}
			},
			want:    MediaCharacter{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &MockStore{}
			animeService := &MockAnimeService{}
			tt.setup(store, animeService)

			got, err := Roll(context.Background(), store, animeService, tt.config, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Roll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.ID != tt.want.ID || got.Name != tt.want.Name || got.ImageURL != tt.want.ImageURL || got.URL != tt.want.URL || got.Description != tt.want.Description || got.MediaTitle != tt.want.MediaTitle {
					t.Errorf("Roll() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestCheckOwnership(t *testing.T) {
	tests := []struct {
		name     string
		userID   corde.Snowflake
		charID   int64
		setup    func(*MockStore)
		want     bool
		wantChar collectionstore.Character
		wantErr  bool
	}{
		{
			name:   "user owns character",
			userID: 123,
			charID: 456,
			setup: func(ms *MockStore) {
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						GetFunc: func(ctx context.Context, arg collectionstore.GetParams) (collectionstore.Character, error) {
							return collectionstore.Character{
								ID:     456,
								UserID: 123,
								Name:   "Test Char",
								Image:  "test.jpg",
								Type:   "ROLL",
								Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
							}, nil
						},
					}
				}
			},
			want: true,
			wantChar: collectionstore.Character{
				ID:     456,
				UserID: 123,
				Name:   "Test Char",
				Image:  "test.jpg",
				Type:   "ROLL",
				Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
			},
			wantErr: false,
		},
		{
			name:   "user does not own character",
			userID: 123,
			charID: 456,
			setup: func(ms *MockStore) {
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						GetFunc: func(ctx context.Context, arg collectionstore.GetParams) (collectionstore.Character, error) {
							return collectionstore.Character{}, assert.AnError
						},
					}
				}
			},
			want:     false,
			wantChar: collectionstore.Character{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &MockStore{}
			tt.setup(store)

			got, gotChar, err := CheckOwnership(context.Background(), store, tt.userID, tt.charID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckOwnership() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got != tt.want {
					t.Errorf("CheckOwnership() got = %v, want %v", got, tt.want)
				}
				if gotChar.ID != tt.wantChar.ID || gotChar.UserID != tt.wantChar.UserID || gotChar.Name != tt.wantChar.Name || gotChar.Image != tt.wantChar.Image || gotChar.Type != tt.wantChar.Type {
					t.Errorf("CheckOwnership() gotChar = %v, want %v", gotChar, tt.wantChar)
				}
			}
		})
	}
}

func TestSearchGlobalCharacters(t *testing.T) {
	tests := []struct {
		name    string
		term    string
		setup   func(*MockStore)
		want    []collectionstore.SearchGlobalCharactersRow
		wantErr bool
	}{
		{
			name: "successful global search",
			term: "naruto",
			setup: func(ms *MockStore) {
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						SearchGlobalCharactersFunc: func(ctx context.Context, arg collectionstore.SearchGlobalCharactersParams) ([]collectionstore.SearchGlobalCharactersRow, error) {
							return []collectionstore.SearchGlobalCharactersRow{
								{
									Name: "Naruto Uzumaki",
									ID:   123,
								},
							}, nil
						},
					}
				}
			},
			want: []collectionstore.SearchGlobalCharactersRow{
				{
					Name: "Naruto Uzumaki",
					ID:   123,
				},
			},
			wantErr: false,
		},
		{
			name: "global search error",
			term: "naruto",
			setup: func(ms *MockStore) {
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						SearchGlobalCharactersFunc: func(ctx context.Context, arg collectionstore.SearchGlobalCharactersParams) ([]collectionstore.SearchGlobalCharactersRow, error) {
							return nil, assert.AnError
						},
					}
				}
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &MockStore{}
			tt.setup(store)

			got, err := SearchGlobalCharacters(context.Background(), store, tt.term)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchGlobalCharacters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("SearchGlobalCharacters() len = %v, want %v", len(got), len(tt.want))
				}
				for i, g := range got {
					if i < len(tt.want) {
						w := tt.want[i]
						if g.Name != w.Name || g.ID != w.ID {
							t.Errorf("SearchGlobalCharacters()[%d] = %v, want %v", i, g, w)
						}
					}
				}
			}
		})
	}
}

func TestCharacters(t *testing.T) {
	tests := []struct {
		name    string
		userID  corde.Snowflake
		setup   func(*MockStore)
		want    []Character
		wantErr bool
	}{
		{
			name:   "successful list",
			userID: 123,
			setup: func(ms *MockStore) {
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						ListFunc: func(ctx context.Context, userID uint64) ([]collectionstore.Character, error) {
							return []collectionstore.Character{
								{
									ID:     456,
									UserID: 123,
									Name:   "Char1",
									Image:  "img1.jpg",
									Type:   "ROLL",
									Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
								},
								{
									ID:     789,
									UserID: 123,
									Name:   "Char2",
									Image:  "img2.jpg",
									Type:   "GIFT",
									Date:   pgtype.Timestamp{Time: time.Now(), Valid: true},
								},
							}, nil
						},
					}
				}
			},
			want: []Character{
				{
					ID:     456,
					UserID: 123,
					Name:   "Char1",
					Image:  "img1.jpg",
					Type:   "ROLL",
					Date:   time.Now(),
				},
				{
					ID:     789,
					UserID: 123,
					Name:   "Char2",
					Image:  "img2.jpg",
					Type:   "GIFT",
					Date:   time.Now(),
				},
			},
			wantErr: false,
		},
		{
			name:   "list error",
			userID: 123,
			setup: func(ms *MockStore) {
				ms.CollectionStoreFunc = func() collectionstore.Querier {
					return &MockCollectionStore{
						ListFunc: func(ctx context.Context, userID uint64) ([]collectionstore.Character, error) {
							return nil, assert.AnError
						},
					}
				}
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &MockStore{}
			tt.setup(store)

			got, err := Characters(context.Background(), store, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Characters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("Characters() len = %v, want %v", len(got), len(tt.want))
				}
				for i, g := range got {
					if i < len(tt.want) {
						w := tt.want[i]
						if g.ID != w.ID || g.UserID != w.UserID || g.Name != w.Name || g.Image != w.Image || g.Type != w.Type {
							t.Errorf("Characters()[%d] = %v, want %v", i, g, w)
						}
					}
				}
			}
		})
	}
}
