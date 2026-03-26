//go:build integration

package collection_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/catalogpg"
	"github.com/karitham/waifubot/storage/collectionpg"
	"github.com/karitham/waifubot/storage/droppg"
	"github.com/karitham/waifubot/storage/guildpg"
	"github.com/karitham/waifubot/storage/userpg"
	"github.com/karitham/waifubot/storage/userstore"
)

var testDBURL string

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase("waifubot_test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		panic("failed to start postgres container: " + err.Error())
	}
	defer func() {
		if err := testcontainers.TerminateContainer(pgContainer); err != nil {
			panic("failed to terminate container: " + err.Error())
		}
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic("failed to get connection string: " + err.Error())
	}
	testDBURL = connStr

	if err := storage.Migrate(testDBURL); err != nil {
		panic("migration failed: " + err.Error())
	}

	m.Run()
}

func buildStore(s storage.Store) collection.Store {
	return collection.NewPostgresStore(
		userpg.New(s.UserStore()),
		collectionpg.New(s.CollectionStore(), s.WishlistStore()),
		droppg.New(s.DropStore()),
		guildpg.New(s.GuildStore()),
		catalogpg.New(s.CollectionStore(), s.GuildStore()),
		s.DB(),
		nil,
	)
}

func setupStore(t *testing.T) collection.Store {
	t.Helper()

	dbStore, err := storage.NewStore(t.Context(), testDBURL)
	require.NoError(t, err)

	txStore, err := dbStore.Tx(t.Context())
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = txStore.Rollback(t.Context())
	})

	return buildStore(txStore)
}

func setupStoreWithSeed(t *testing.T, userIDs ...uint64) collection.Store {
	t.Helper()
	store := setupStore(t)
	ctx := t.Context()
	for _, uid := range userIDs {
		require.NoError(t, store.CreateUser(ctx, uid))
	}
	return store
}

func TestIntegration_UserCRUD(t *testing.T) {
	store := setupStore(t)
	ctx := t.Context()
	const uid uint64 = 900001

	require.NoError(t, store.CreateUser(ctx, uid))

	user, err := store.GetUser(ctx, uid)
	require.NoError(t, err)
	assert.Equal(t, uid, user.UserID)
	assert.Equal(t, int32(0), user.Tokens)

	require.NoError(t, store.UpdateQuote(ctx, uid, "test quote"))
	user, _ = store.GetUser(ctx, uid)
	assert.Equal(t, "test quote", user.Quote)

	require.NoError(t, store.UpdateFavorite(ctx, uid, 42))
	user, _ = store.GetUser(ctx, uid)
	assert.Equal(t, int64(42), user.Favorite)

	require.NoError(t, store.UpdateAnilistURL(ctx, uid, "https://anilist.co/user/test"))
	user, _ = store.GetUser(ctx, uid)
	assert.Equal(t, "https://anilist.co/user/test", user.AnilistURL)

	require.NoError(t, store.UpdateLastRoll(ctx, uid, time.Now()))

	user, err = store.AddTokens(ctx, uid, 10)
	require.NoError(t, err)
	assert.Equal(t, int32(10), user.Tokens)

	user, err = store.SpendTokens(ctx, uid, 3)
	require.NoError(t, err)
	assert.Equal(t, int32(7), user.Tokens)

	_, err = store.GetUser(ctx, 999999)
	require.ErrorIs(t, err, collection.ErrNotFound)
}

func TestIntegration_CharacterAndCollection(t *testing.T) {
	const uid uint64 = 900002
	store := setupStoreWithSeed(t, uid)
	ctx := t.Context()

	require.NoError(t, store.UpsertCharacter(ctx, collection.Character{ID: 1001, Name: "TestChar", Image: "test.jpg"}))
	char, err := store.GetCharacterByID(ctx, 1001)
	require.NoError(t, err)
	assert.Equal(t, "TestChar", char.Name)

	err = store.AddToCollection(ctx, uid, collection.Character{ID: 1001}, "ROLL", time.Now())
	require.NoError(t, err, "first AddToCollection should succeed")

	chars, err := store.GetCollection(ctx, uid)
	require.NoError(t, err)
	require.Len(t, chars, 1)
	assert.Equal(t, int64(1001), chars[0].ID)
	assert.Equal(t, "ROLL", chars[0].Source)

	ids, _ := store.GetCollectionIDs(ctx, uid)
	assert.Equal(t, []int64{1001}, ids)

	oc, _ := store.GetOwnedCharacter(ctx, uid, 1001)
	assert.Equal(t, int64(1001), oc.ID)
	_, err = store.GetOwnedCharacter(ctx, uid, 9999)
	require.ErrorIs(t, err, collection.ErrNotFound)

	count, _ := store.CountCollection(ctx, uid)
	assert.Equal(t, int64(1), count)

	require.NoError(t, store.RemoveFromCollection(ctx, uid, 1001))
	chars, _ = store.GetCollection(ctx, uid)
	assert.Empty(t, chars)
}

func TestIntegration_CharacterAlreadyOwned(t *testing.T) {
	const uid uint64 = 900099
	store := setupStoreWithSeed(t, uid)
	ctx := t.Context()

	require.NoError(t, store.UpsertCharacter(ctx, collection.Character{ID: 1099, Name: "DupeChar", Image: "dupe.jpg"}))
	require.NoError(t, store.AddToCollection(ctx, uid, collection.Character{ID: 1099}, "ROLL", time.Now()))

	err := store.AddToCollection(ctx, uid, collection.Character{ID: 1099}, "ROLL", time.Now())
	require.ErrorIs(t, err, collection.ErrAlreadyOwned)
}

func TestIntegration_GiveCharacter(t *testing.T) {
	const u1, u2 uint64 = 900003, 900004
	store := setupStoreWithSeed(t, u1, u2)
	ctx := t.Context()

	require.NoError(t, store.UpsertCharacter(ctx, collection.Character{ID: 2001, Name: "GiveChar", Image: "give.jpg"}))
	require.NoError(t, store.AddToCollection(ctx, u1, collection.Character{ID: 2001}, "ROLL", time.Now()))

	oc, err := store.GiveCharacter(ctx, u1, u2, 2001)
	require.NoError(t, err)
	assert.Equal(t, "TRADE", oc.Source)
	assert.Equal(t, u2, oc.UserID)
}

func TestIntegration_SearchCharacters(t *testing.T) {
	const uid uint64 = 900005
	store := setupStoreWithSeed(t, uid)
	ctx := t.Context()

	require.NoError(t, store.UpsertCharacter(ctx, collection.Character{ID: 3001, Name: "Searchable Char", Image: "s.jpg"}))
	require.NoError(t, store.AddToCollection(ctx, uid, collection.Character{ID: 3001}, "ROLL", time.Now()))

	results, err := store.SearchCharacters(ctx, uid, "Search")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, int64(3001), results[0].ID)

	results, err = store.SearchGlobalCharacters(ctx, "Searchable")
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(results), 1)
	found := false
	for _, r := range results {
		if r.ID == 3001 {
			assert.Equal(t, "Searchable Char", r.Name)
			found = true
		}
	}
	assert.True(t, found)
}

func TestIntegration_Transaction(t *testing.T) {
	const uid uint64 = 900006

	dbStore, err := storage.NewStore(t.Context(), testDBURL)
	require.NoError(t, err)
	ctx := t.Context()

	require.NoError(t, dbStore.UserStore().Create(ctx, uid))

	txStore, err := dbStore.Tx(ctx)
	require.NoError(t, err)
	require.NoError(t, txStore.UserStore().UpdateQuote(ctx, userstore.UpdateQuoteParams{
		Quote:  "committed",
		UserID: uid,
	}))
	require.NoError(t, txStore.Commit(ctx))

	user, err := dbStore.UserStore().Get(ctx, uid)
	require.NoError(t, err)
	assert.Equal(t, "committed", user.Quote)

	txStore, err = dbStore.Tx(ctx)
	require.NoError(t, err)
	require.NoError(t, txStore.UserStore().UpdateQuote(ctx, userstore.UpdateQuoteParams{
		Quote:  "rolled back",
		UserID: uid,
	}))
	require.NoError(t, txStore.Rollback(ctx))

	user, err = dbStore.UserStore().Get(ctx, uid)
	require.NoError(t, err)
	assert.Equal(t, "committed", user.Quote)
}

func TestIntegration_GuildOperations(t *testing.T) {
	store := setupStore(t)
	ctx := t.Context()
	const gid uint64 = 900007

	require.NoError(t, store.StartIndexingJob(ctx, gid))
	status, _ := store.IsGuildIndexed(ctx, gid)
	assert.Equal(t, collection.IndexingInProgress, status.Status)

	require.NoError(t, store.UpsertGuildMembers(ctx, gid, []uint64{100, 200, 300}, time.Now()))
	require.NoError(t, store.CompleteIndexingJob(ctx, gid))

	status, _ = store.IsGuildIndexed(ctx, gid)
	assert.Equal(t, collection.IndexingCompleted, status.Status)

	require.NoError(t, store.DeleteGuildMembersNotIn(ctx, gid, []uint64{100, 200}))
	require.NoError(t, store.UpsertGuildMembers(ctx, gid, []uint64{100, 200}, time.Now()))
}

func TestIntegration_TokenConsistency(t *testing.T) {
	const uid uint64 = 900008
	store := setupStore(t)
	ctx := t.Context()

	require.NoError(t, store.CreateUser(ctx, uid))
	user, err := store.AddTokens(ctx, uid, 100)
	require.NoError(t, err)
	assert.Equal(t, int32(100), user.Tokens)

	for i := range 10 {
		user, err = store.SpendTokens(ctx, uid, 1)
		require.NoError(t, err)
		assert.Equal(t, int32(99-i), user.Tokens)
	}

	user, _ = store.GetUser(ctx, uid)
	assert.Equal(t, int32(90), user.Tokens)
}

func TestIntegration_WishlistBridge(t *testing.T) {
	const uid uint64 = 900009
	store := setupStoreWithSeed(t, uid)
	ctx := t.Context()

	require.NoError(t, store.UpsertCharacter(ctx, collection.Character{ID: 4001, Name: "WishChar"}))
	require.NoError(t, store.RemoveFromWishlist(ctx, uid, 4001))
}

func TestIntegration_DropOperations(t *testing.T) {
	store := setupStore(t)
	ctx := t.Context()

	_, err := store.GetDropForUpdate(ctx, 888001)
	require.ErrorIs(t, err, collection.ErrNotFound)
}

func TestIntegration_GetUserByAnilist(t *testing.T) {
	const uid uint64 = 900010
	store := setupStore(t)
	ctx := t.Context()

	require.NoError(t, store.CreateUser(ctx, uid))
	const url = "https://anilist.co/user/test900010"
	require.NoError(t, store.UpdateAnilistURL(ctx, uid, url))

	user, err := store.GetUserByAnilist(ctx, url)
	require.NoError(t, err)
	assert.Equal(t, uid, user.UserID)
}

func TestIntegration_GetUserByDiscordUsername(t *testing.T) {
	const uid uint64 = 900011
	store := setupStore(t)
	ctx := t.Context()

	require.NoError(t, store.CreateUser(ctx, uid))
	const username = "testuser900011"
	require.NoError(t, store.UpdateDiscordInfo(ctx, uid, username, "avatar_hash", time.Now()))

	user, err := store.GetUserByDiscordUsername(ctx, username)
	require.NoError(t, err)
	assert.Equal(t, uid, user.UserID)
}

func TestIntegration_GetCharacterHoldersInGuild(t *testing.T) {
	const u1, u2, gid uint64 = 900012, 900013, 900014
	store := setupStoreWithSeed(t, u1, u2)
	ctx := t.Context()

	require.NoError(t, store.StartIndexingJob(ctx, gid))
	require.NoError(t, store.UpsertGuildMembers(ctx, gid, []uint64{u1, u2}, time.Now()))
	require.NoError(t, store.CompleteIndexingJob(ctx, gid))

	require.NoError(t, store.UpsertCharacter(ctx, collection.Character{ID: 5001, Name: "GuildChar"}))
	require.NoError(t, store.AddToCollection(ctx, u1, collection.Character{ID: 5001}, "ROLL", time.Now()))

	holders, err := store.GetCharacterHoldersInGuild(ctx, gid, 5001)
	require.NoError(t, err)
	assert.Equal(t, []uint64{u1}, holders)

	holders, _ = store.GetCharacterHoldersInGuild(ctx, gid, 9999)
	assert.Empty(t, holders)
}
