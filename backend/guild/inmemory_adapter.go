package guild

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/collectionstore"
	"github.com/karitham/waifubot/storage/commandstore"
	"github.com/karitham/waifubot/storage/dropstore"
	"github.com/karitham/waifubot/storage/guildstore"
	"github.com/karitham/waifubot/storage/inmemory"
	"github.com/karitham/waifubot/storage/interactionstore"
	"github.com/karitham/waifubot/storage/interfaces"
	"github.com/karitham/waifubot/storage/userstore"
	"github.com/karitham/waifubot/storage/wishliststore"
)

type InMemoryStore struct {
	store *inmemory.Store
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		store: inmemory.NewStore(),
	}
}

func (s *InMemoryStore) DB() storage.TXer {
	return nil
}

func (s *InMemoryStore) Tx(ctx context.Context) (storage.Store, error) {
	return s, nil
}

func (s *InMemoryStore) Commit(ctx context.Context) error {
	return nil
}

func (s *InMemoryStore) Rollback(ctx context.Context) error {
	return nil
}

func (s *InMemoryStore) CollectionStore() collectionstore.Querier {
	return nil
}

func (s *InMemoryStore) UserStore() userstore.Querier {
	return nil
}

func (s *InMemoryStore) GuildStore() guildstore.Querier {
	return &guildAdapter{guildStore: s.store.GuildStore()}
}

func (s *InMemoryStore) WishlistStore() wishliststore.Querier {
	return nil
}

func (s *InMemoryStore) CommandStore() commandstore.Querier {
	return nil
}

func (s *InMemoryStore) DropStore() dropstore.Querier {
	return nil
}

func (s *InMemoryStore) InteractionStore() interactionstore.Querier {
	return nil
}

type guildAdapter struct {
	guildStore interfaces.GuildRepository
}

func (a *guildAdapter) CompleteIndexingJob(ctx context.Context, guildID uint64) error {
	return a.guildStore.CompleteIndexingJob(ctx, guildID)
}

func (a *guildAdapter) DeleteGuildMembers(ctx context.Context, guildID uint64) error {
	return a.guildStore.DeleteGuildMembers(ctx, guildID)
}

func (a *guildAdapter) DeleteGuildMembersNotIn(ctx context.Context, arg guildstore.DeleteGuildMembersNotInParams) error {
	userIDs := make([]uint64, len(arg.Column2))
	for i, id := range arg.Column2 {
		userIDs[i] = uint64(id)
	}
	return a.guildStore.DeleteGuildMembersNotIn(ctx, arg.GuildID, userIDs)
}

func (a *guildAdapter) GetGuildMembers(ctx context.Context, guildID uint64) ([]uint64, error) {
	return a.guildStore.GetGuildMembers(ctx, guildID)
}

func (a *guildAdapter) GetIndexingStatus(ctx context.Context, guildID uint64) (guildstore.GetIndexingStatusRow, error) {
	status, err := a.guildStore.GetIndexingStatus(ctx, guildID)
	return guildstore.GetIndexingStatusRow{
		Status:    guildstore.IndexingStatus(status.Status),
		UpdatedAt: pgtype.Timestamp{Time: status.UpdatedAt, Valid: true},
	}, err
}

func (a *guildAdapter) IsGuildIndexed(ctx context.Context, guildID uint64) (guildstore.IsGuildIndexedRow, error) {
	completed, _, err := a.guildStore.IsGuildIndexed(ctx, guildID)
	var idxStatus guildstore.IndexingStatus
	if completed {
		idxStatus = guildstore.IndexingStatusCompleted
	} else {
		idxStatus = guildstore.IndexingStatusPending
	}
	return guildstore.IsGuildIndexedRow{
		Status: idxStatus,
	}, err
}

func (a *guildAdapter) StartIndexingJob(ctx context.Context, guildID uint64) error {
	return a.guildStore.StartIndexingJob(ctx, guildID)
}

func (a *guildAdapter) UpsertGuildMembers(ctx context.Context, arg guildstore.UpsertGuildMembersParams) error {
	userIDs := make([]uint64, len(arg.Column2))
	for i, id := range arg.Column2 {
		userIDs[i] = uint64(id)
	}
	return a.guildStore.UpsertGuildMembers(ctx, arg.GuildID, userIDs)
}

func (a *guildAdapter) UsersOwningCharInGuild(ctx context.Context, arg guildstore.UsersOwningCharInGuildParams) ([]uint64, error) {
	return a.guildStore.UsersOwningCharInGuild(ctx, arg.GuildID, arg.CharacterID)
}
