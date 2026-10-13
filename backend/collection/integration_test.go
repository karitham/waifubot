//go:build integration

package collection_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/karitham/waifubot/catalog"
	"github.com/karitham/waifubot/collection"
	"github.com/karitham/waifubot/drops"
	"github.com/karitham/waifubot/guild"
	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/pg"
)

var testDBURL string

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, "postgres:18-alpine",
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

func buildTx(s storage.TxStore) collection.Tx {
	return pg.NewTx(s)
}

func setupStore(t *testing.T) (collection.Tx, catalog.Store) {
	t.Helper()

	dbStore, err := storage.NewStore(t.Context(), testDBURL)
	require.NoError(t, err)

	txStore, err := dbStore.Tx(t.Context())
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = txStore.Rollback(t.Context())
	})

	tx := pg.NewTx(txStore)
	return tx, tx
}

func setupStoreWithSeed(t *testing.T, userIDs ...collection.UserID) (collection.Tx, catalog.Store) {
	t.Helper()
	store, cat := setupStore(t)
	ctx := t.Context()
	for _, uid := range userIDs {
		require.NoError(t, store.CreateUser(ctx, uid))
	}
	return store, cat
}

func TestIntegration_UserCRUD(t *testing.T) {
	store, cat := setupStore(t)
	ctx := t.Context()
	const uid collection.UserID = 900001

	require.NoError(t, store.CreateUser(ctx, uid))

	user, err := store.GetUser(ctx, uid)
	require.NoError(t, err)
	assert.Equal(t, uid, user.UserID)
	assert.Equal(t, int32(0), user.Tokens)

	require.NoError(t, store.UpdateQuote(ctx, uid, "test quote"))
	user, _ = store.GetUser(ctx, uid)
	assert.Equal(t, "test quote", user.Quote)

	require.NoError(t, cat.UpsertCharacter(ctx, catalog.Character{ID: 42, Name: "FavChar"}))
	require.NoError(t, store.AddToCollection(ctx, uid, catalog.CharacterID(42), collection.SourceRoll, time.Now()))
	require.NoError(t, store.SetFavorite(ctx, uid, catalog.CharacterID(42)))
	user, _ = store.GetUser(ctx, uid)
	assert.Equal(t, catalog.CharacterID(42), user.FavoriteCharacterID)

	require.NoError(t, store.UpdateAnilistURL(ctx, uid, "https://anilist.co/user/test"))
	user, _ = store.GetUser(ctx, uid)
	assert.Equal(t, "https://anilist.co/user/test", user.AnilistURL)

	require.NoError(t, store.RecordRoll(ctx, uid, time.Now(), time.Hour))

	user, err = store.AddTokens(ctx, uid, tokenAmount(t, 10))
	require.NoError(t, err)
	assert.Equal(t, int32(10), user.Tokens)

	user, err = store.SpendTokens(ctx, uid, tokenAmount(t, 3))
	require.NoError(t, err)
	assert.Equal(t, int32(7), user.Tokens)

	_, err = store.GetUser(ctx, 999999)
	require.ErrorIs(t, err, collection.ErrNotFound)
}

func TestIntegration_CharacterAndCollection(t *testing.T) {
	const uid collection.UserID = 900002
	store, cat := setupStoreWithSeed(t, uid)
	ctx := t.Context()

	require.NoError(t, cat.UpsertCharacter(ctx, catalog.Character{ID: 1001, Name: "TestChar", Image: "test.jpg"}))
	char, err := store.GetCharacterByID(ctx, 1001)
	require.NoError(t, err)
	assert.Equal(t, "TestChar", char.Name)

	err = store.AddToCollection(ctx, uid, catalog.CharacterID(1001), collection.SourceRoll, time.Now())
	require.NoError(t, err, "first AddToCollection should succeed")

	chars, err := store.GetCollection(ctx, uid)
	require.NoError(t, err)
	require.Len(t, chars, 1)
	assert.Equal(t, catalog.CharacterID(1001), chars[0].ID)
	assert.Equal(t, collection.SourceRoll, chars[0].Source)

	ids, _ := store.GetCollectionIDs(ctx, uid)
	assert.Equal(t, []catalog.CharacterID{1001}, ids)

	oc, _ := store.GetOwnedCharacter(ctx, uid, 1001)
	assert.Equal(t, catalog.CharacterID(1001), oc.ID)
	_, err = store.GetOwnedCharacter(ctx, uid, 9999)
	require.ErrorIs(t, err, collection.ErrNotFound)

	count, _ := store.CountCollection(ctx, uid)
	assert.Equal(t, int64(1), count)

	require.NoError(t, store.RemoveFromCollection(ctx, uid, 1001))
	chars, _ = store.GetCollection(ctx, uid)
	assert.Empty(t, chars)
}

func TestIntegration_CharacterAlreadyOwned(t *testing.T) {
	const uid collection.UserID = 900099
	store, cat := setupStoreWithSeed(t, uid)
	ctx := t.Context()

	require.NoError(t, cat.UpsertCharacter(ctx, catalog.Character{ID: 1099, Name: "DupeChar", Image: "dupe.jpg"}))
	require.NoError(t, store.AddToCollection(ctx, uid, catalog.CharacterID(1099), collection.SourceRoll, time.Now()))

	err := store.AddToCollection(ctx, uid, catalog.CharacterID(1099), collection.SourceRoll, time.Now())
	require.ErrorIs(t, err, collection.ErrAlreadyOwned)
}

func TestIntegration_GiveCharacter(t *testing.T) {
	const u1, u2 collection.UserID = 900003, 900004
	store, cat := setupStoreWithSeed(t, u1, u2)
	ctx := t.Context()

	require.NoError(t, cat.UpsertCharacter(ctx, catalog.Character{ID: 2001, Name: "GiveChar", Image: "give.jpg"}))
	require.NoError(t, store.AddToCollection(ctx, u1, catalog.CharacterID(2001), collection.SourceRoll, time.Now()))

	oc, err := store.GiveCharacter(ctx, u1, u2, 2001)
	require.NoError(t, err)
	assert.Equal(t, collection.SourceTrade, oc.Source)
	assert.Equal(t, u2, oc.UserID)
}

func TestIntegration_SearchCharacters(t *testing.T) {
	const uid collection.UserID = 900005
	store, cat := setupStoreWithSeed(t, uid)
	ctx := t.Context()

	require.NoError(t, cat.UpsertCharacter(ctx, catalog.Character{ID: 3001, Name: "Searchable Char", Image: "s.jpg"}))
	require.NoError(t, store.AddToCollection(ctx, uid, catalog.CharacterID(3001), collection.SourceRoll, time.Now()))

	results, err := store.SearchCharacters(ctx, uid, "Search")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, catalog.CharacterID(3001), results[0].ID)

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
	const uid collection.UserID = 900006

	dbStore, err := storage.NewStore(t.Context(), testDBURL)
	require.NoError(t, err)
	ctx := t.Context()

	require.NoError(t, dbStore.Querier().CreateUser(ctx, uint64(uid)))

	txStore, err := dbStore.Tx(ctx)
	require.NoError(t, err)
	require.NoError(t, txStore.Querier().UpdateQuote(ctx, storage.UpdateQuoteParams{
		Quote:  "committed",
		UserID: uint64(uid),
	}))
	require.NoError(t, txStore.Commit(ctx))

	user, err := dbStore.Querier().GetUser(ctx, uint64(uid))
	require.NoError(t, err)
	assert.Equal(t, "committed", user.Quote)

	txStore, err = dbStore.Tx(ctx)
	require.NoError(t, err)
	require.NoError(t, txStore.Querier().UpdateQuote(ctx, storage.UpdateQuoteParams{
		Quote:  "rolled back",
		UserID: uint64(uid),
	}))
	require.NoError(t, txStore.Rollback(ctx))

	user, err = dbStore.Querier().GetUser(ctx, uint64(uid))
	require.NoError(t, err)
	assert.Equal(t, "committed", user.Quote)
}

func TestIntegration_GuildOperations(t *testing.T) {
	store, _ := setupStore(t)
	ctx := t.Context()
	const gid guild.ID = 900007

	require.NoError(t, store.StartIndexingJob(ctx, gid))
	status, _ := store.IsGuildIndexed(ctx, gid)
	assert.Equal(t, guild.IndexingInProgress, status.Status)

	require.NoError(t, store.UpsertGuildMembers(ctx, gid, []guild.MemberID{100, 200, 300}, time.Now()))
	require.NoError(t, store.CompleteIndexingJob(ctx, gid))

	status, _ = store.IsGuildIndexed(ctx, gid)
	assert.Equal(t, guild.IndexingCompleted, status.Status)

	require.NoError(t, store.DeleteGuildMembersNotIn(ctx, gid, []guild.MemberID{100, 200}))
	require.NoError(t, store.UpsertGuildMembers(ctx, gid, []guild.MemberID{100, 200}, time.Now()))
}

func TestIntegration_TokenConsistency(t *testing.T) {
	const uid collection.UserID = 900008
	store, _ := setupStore(t)
	ctx := t.Context()

	require.NoError(t, store.CreateUser(ctx, uid))
	user, err := store.AddTokens(ctx, uid, tokenAmount(t, 100))
	require.NoError(t, err)
	assert.Equal(t, int32(100), user.Tokens)

	for i := range 10 {
		user, err = store.SpendTokens(ctx, uid, tokenAmount(t, 1))
		require.NoError(t, err)
		assert.Equal(t, int32(99-i), user.Tokens)
	}

	user, _ = store.GetUser(ctx, uid)
	assert.Equal(t, int32(90), user.Tokens)
}

func TestIntegration_WishlistBridge(t *testing.T) {
	const uid collection.UserID = 900009
	store, cat := setupStoreWithSeed(t, uid)
	ctx := t.Context()

	require.NoError(t, cat.UpsertCharacter(ctx, catalog.Character{ID: 4001, Name: "WishChar"}))
	require.NoError(t, store.RemoveFromWishlist(ctx, uid, 4001))
}

func TestIntegration_ReadsForFreshUser(t *testing.T) {
	const uid collection.UserID = 900027
	store, _ := setupStore(t)
	ctx := t.Context()

	profile, err := collection.UserProfile(ctx, store, uid)
	require.NoError(t, err)
	assert.Equal(t, uid, profile.User.UserID)
	assert.Equal(t, int32(0), profile.User.Tokens)
	assert.Equal(t, 0, profile.CharacterCount)
	assert.Nil(t, profile.FavoriteCharacter)

	tokens, err := collection.Balance(ctx, store, uid)
	require.NoError(t, err)
	assert.Equal(t, int32(0), tokens)

	_, err = store.GetUser(ctx, uid)
	require.ErrorIs(t, err, collection.ErrNotFound, "reads must not create the user")
}

func TestIntegration_AddToWishlistFreshUser(t *testing.T) {
	const uid collection.UserID = 900026
	ctx := t.Context()

	dbStore, err := storage.NewStore(ctx, testDBURL)
	require.NoError(t, err)
	store := pg.NewPool(dbStore)
	cat := pg.NewPool(dbStore)

	require.NoError(t, cat.UpsertCharacter(ctx, catalog.Character{ID: 4002, Name: "FreshWish"}))

	added, err := collection.AddToWishlist(ctx, store, uid, []catalog.CharacterID{4002})
	require.NoError(t, err)
	assert.Equal(t, int64(1), added)

	_, err = store.GetUser(ctx, uid)
	require.NoError(t, err, "adding to the wishlist must materialize the user")

	entries, err := pg.NewPool(dbStore).GetUserCharacterWishlist(ctx, uid)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, catalog.CharacterID(4002), entries[0].ID)
}

func TestIntegration_WishlistSkipsOwnedCharacters(t *testing.T) {
	const uid collection.UserID = 900032
	ctx := t.Context()

	dbStore, err := storage.NewStore(ctx, testDBURL)
	require.NoError(t, err)
	store := pg.NewPool(dbStore)
	cat := pg.NewPool(dbStore)

	require.NoError(t, cat.UpsertCharacter(ctx, catalog.Character{ID: 4003, Name: "OwnedWish"}))
	require.NoError(t, cat.UpsertCharacter(ctx, catalog.Character{ID: 4004, Name: "WantedWish"}))

	tx, err := store.Begin(ctx)
	require.NoError(t, err)
	require.NoError(t, tx.CreateUser(ctx, uid))
	require.NoError(t, tx.AddToCollection(ctx, uid, 4003, collection.SourceRoll, time.Now()))
	require.NoError(t, tx.Commit(ctx))

	added, err := collection.AddToWishlist(ctx, store, uid, []catalog.CharacterID{4003, 4004})
	require.NoError(t, err)
	assert.Equal(t, int64(1), added, "only the unowned character may be added")

	entries, err := pg.NewPool(dbStore).GetUserCharacterWishlist(ctx, uid)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, catalog.CharacterID(4004), entries[0].ID)
}

func TestIntegration_DropOperations(t *testing.T) {
	store, _ := setupStore(t)
	ctx := t.Context()

	_, err := store.GetDropForUpdate(ctx, 888001)
	require.ErrorIs(t, err, drops.ErrNoDropInChannel)
}

func TestIntegration_Trigger(t *testing.T) {
	ctx := t.Context()

	dbStore, err := storage.NewStore(ctx, testDBURL)
	require.NoError(t, err)
	cat := pg.NewPool(dbStore)

	require.NoError(t, cat.UpsertCharacter(ctx, catalog.Character{ID: 7001, Name: "DropChar", Image: "img", Popularity: 100}))

	drop, triggered, err := drops.Trigger(ctx, pg.NewDropsDB(dbStore), 888002, 1)
	require.NoError(t, err)
	require.True(t, triggered)

	txStore, err := dbStore.Tx(ctx)
	require.NoError(t, err)
	tx := pg.NewTx(txStore)
	got, err := tx.GetDropForUpdate(ctx, 888002)
	require.NoError(t, err)
	require.NoError(t, tx.Rollback(ctx))
	assert.Equal(t, drop.Character.ID, got.Character.ID)
}

// TestIntegration_TriggerSerializes runs concurrent triggers against one
// channel: the increments serialize on the counter row, so every full
// threshold window is consumed exactly once.
func TestIntegration_TriggerSerializes(t *testing.T) {
	ctx := t.Context()

	dbStore, err := storage.NewStore(ctx, testDBURL)
	require.NoError(t, err)
	cat := pg.NewPool(dbStore)

	require.NoError(t, cat.UpsertCharacter(ctx, catalog.Character{ID: 7002, Name: "ConcurrentChar", Image: "img", Popularity: 100}))

	const (
		channel   drops.ChannelID = 888003
		threshold int64           = 5
		callers                   = 10
	)

	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		triggered int
		errs      = make(chan error, callers)
	)

	for range callers {
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, ok, err := drops.Trigger(ctx, pg.NewDropsDB(dbStore), channel, threshold)
			if err != nil {
				errs <- err
				return
			}
			if ok {
				mu.Lock()
				triggered++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}
	assert.Equal(t, callers/int(threshold), triggered, "each threshold window must be consumed exactly once")
}

func TestIntegration_GetUserByAnilist(t *testing.T) {
	const uid collection.UserID = 900010
	store, _ := setupStore(t)
	ctx := t.Context()

	require.NoError(t, store.CreateUser(ctx, uid))
	const url = "https://anilist.co/user/test900010"
	require.NoError(t, store.UpdateAnilistURL(ctx, uid, url))

	user, err := store.GetUserByAnilist(ctx, url)
	require.NoError(t, err)
	assert.Equal(t, uid, user.UserID)
}

func TestIntegration_GetUserByDiscordUsername(t *testing.T) {
	const uid collection.UserID = 900011
	store, _ := setupStore(t)
	ctx := t.Context()

	require.NoError(t, store.CreateUser(ctx, uid))
	const username = "testuser900011"
	require.NoError(t, store.UpdateDiscordInfo(ctx, uid, username, "avatar_hash", time.Now()))

	user, err := store.GetUserByDiscordUsername(ctx, username)
	require.NoError(t, err)
	assert.Equal(t, uid, user.UserID)
}

func TestIntegration_GetCharacterHoldersInGuild(t *testing.T) {
	const u1, u2 collection.UserID = 900012, 900013
	const gid guild.ID = 900014
	store, cat := setupStore(t)
	ctx := t.Context()

	require.NoError(t, store.CreateUser(ctx, u1))
	require.NoError(t, store.CreateUser(ctx, u2))

	require.NoError(t, store.StartIndexingJob(ctx, gid))
	require.NoError(t, store.UpsertGuildMembers(ctx, gid, []guild.MemberID{guild.MemberID(u1), guild.MemberID(u2)}, time.Now()))
	require.NoError(t, store.CompleteIndexingJob(ctx, gid))

	require.NoError(t, cat.UpsertCharacter(ctx, catalog.Character{ID: 5001, Name: "GuildChar"}))
	require.NoError(t, store.AddToCollection(ctx, u1, catalog.CharacterID(5001), collection.SourceRoll, time.Now()))

	holders, err := store.GetCharacterHoldersInGuild(ctx, gid, 5001)
	require.NoError(t, err)
	assert.Equal(t, []guild.MemberID{guild.MemberID(u1)}, holders)

	holders, _ = store.GetCharacterHoldersInGuild(ctx, gid, 9999)
	assert.Empty(t, holders)
}

func TestIntegration_RandomCharNotOwned_ExcludesInactive(t *testing.T) {
	store, cat := setupStore(t)
	ctx := t.Context()

	// Insert active characters.
	for i := int64(1); i <= 10; i++ {
		require.NoError(t, cat.UpsertCharacter(ctx, catalog.Character{
			ID: catalog.CharacterID(i), Name: fmt.Sprintf("Char%d", i), Image: "img", Popularity: 100,
		}))
	}

	// Mark some as inactive.
	require.NoError(t, cat.MarkCharactersInactive(ctx, []catalog.CharacterID{3, 7, 9}))

	// Roll 50 times and ensure we never get 3, 7, or 9.
	for range 50 {
		char, err := store.RandomCharNotOwned(ctx, 999, collection.RollWeightExponent) // 999 = user with no collection
		require.NoError(t, err)
		assert.NotContains(t, []catalog.CharacterID{3, 7, 9}, char.ID,
			"inactive character %d was rolled", char.ID)
	}
}

func TestIntegration_KnownCharacterIDs(t *testing.T) {
	store, cat := setupStore(t)
	ctx := t.Context()

	require.NoError(t, cat.UpsertCharacter(ctx, catalog.Character{ID: 6001, Name: "Known", Image: "img"}))
	require.NoError(t, cat.MarkCharactersInactive(ctx, []catalog.CharacterID{6001}))

	known, err := store.KnownCharacterIDs(ctx, []catalog.CharacterID{6001, 6002})
	require.NoError(t, err)
	assert.Equal(t, []catalog.CharacterID{6001}, known, "presence decides, not the active flag")
}

func TestIntegration_RollCooldownIsAtomic(t *testing.T) {
	ctx := t.Context()
	const uid collection.UserID = 900015
	const cooldown = time.Hour

	dbStore, err := storage.NewStore(ctx, testDBURL)
	require.NoError(t, err)

	tx1, err := dbStore.Tx(ctx)
	require.NoError(t, err)
	tx2, err := dbStore.Tx(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = tx1.Rollback(ctx)
		_ = tx2.Rollback(ctx)
	})

	store1, store2 := buildTx(tx1), buildTx(tx2)

	require.NoError(t, dbStore.Querier().CreateUser(ctx, uint64(uid)))

	// Both transactions read the same pre-roll state, so both pass the
	// cooldown check.
	_, err = store1.GetUser(ctx, uid)
	require.NoError(t, err)
	_, err = store2.GetUser(ctx, uid)
	require.NoError(t, err)

	now := time.Now()
	require.NoError(t, store1.RecordRoll(ctx, uid, now, cooldown))
	require.NoError(t, tx1.Commit(ctx))

	// The second roll loses: the conditional update sees the committed
	// timestamp even though its read happened before the commit.
	err = store2.RecordRoll(ctx, uid, now, cooldown)

	var cooldownErr collection.ErrRollCooldown
	require.ErrorAs(t, err, &cooldownErr)
	assert.WithinDuration(t, now.Add(cooldown), cooldownErr.Until, time.Minute)
}

func TestIntegration_TransactionsAreScoped(t *testing.T) {
	ctx := t.Context()
	const uid collection.UserID = 900016

	dbStore, err := storage.NewStore(ctx, testDBURL)
	require.NoError(t, err)

	// Wire through the same composition the binaries use, so a regression
	// in the tx factory fails this test.
	store := pg.NewPool(dbStore)

	tx, err := store.Begin(ctx)
	require.NoError(t, err)
	require.NoError(t, tx.CreateUser(ctx, uid))
	require.NoError(t, tx.Rollback(ctx))

	_, err = dbStore.Querier().GetUser(ctx, uint64(uid))
	require.ErrorIs(t, err, pgx.ErrNoRows, "write escaped the transaction")
}

func TestIntegration_TransferToFreshUser(t *testing.T) {
	const giver, recipient, other collection.UserID = 900021, 900022, 900023
	ctx := t.Context()

	dbStore, err := storage.NewStore(ctx, testDBURL)
	require.NoError(t, err)
	store := pg.NewPool(dbStore)

	tx, err := store.Begin(ctx)
	require.NoError(t, err)
	require.NoError(t, tx.CreateUser(ctx, giver))
	_, err = tx.AddTokens(ctx, giver, tokenAmount(t, 5))
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))

	require.NoError(t, collection.TransferTokens(ctx, store, giver, recipient, tokenAmount(t, 3)))

	fresh, err := store.GetUser(ctx, recipient)
	require.NoError(t, err, "receiving tokens must materialize the recipient")
	assert.Equal(t, int32(3), fresh.Tokens)

	sender, err := store.GetUser(ctx, giver)
	require.NoError(t, err)
	assert.Equal(t, int32(2), sender.Tokens)

	err = collection.TransferTokens(ctx, store, giver, other, tokenAmount(t, 999))
	require.ErrorIs(t, err, collection.ErrInsufficientTokens)

	_, err = store.GetUser(ctx, other)
	require.ErrorIs(t, err, collection.ErrNotFound, "a failed transfer must not materialize its target")
}

func TestIntegration_GiveCharacterToFreshUser(t *testing.T) {
	const giver, recipient collection.UserID = 900024, 900025
	ctx := t.Context()

	dbStore, err := storage.NewStore(ctx, testDBURL)
	require.NoError(t, err)
	store := pg.NewPool(dbStore)
	cat := pg.NewPool(dbStore)

	require.NoError(t, cat.UpsertCharacter(ctx, catalog.Character{ID: 2002, Name: "FreshChar", Image: "fresh.jpg"}))

	tx, err := store.Begin(ctx)
	require.NoError(t, err)
	require.NoError(t, tx.CreateUser(ctx, giver))
	require.NoError(t, tx.AddToCollection(ctx, giver, catalog.CharacterID(2002), collection.SourceRoll, time.Now()))
	require.NoError(t, tx.Commit(ctx))

	given, err := collection.Give(ctx, store, giver, recipient, 2002)
	require.NoError(t, err)
	assert.Equal(t, recipient, given.UserID)

	fresh, err := store.GetUser(ctx, recipient)
	require.NoError(t, err, "receiving a character must materialize the recipient")
	assert.Equal(t, recipient, fresh.UserID)

	owned, err := store.GetOwnedCharacter(ctx, recipient, 2002)
	require.NoError(t, err)
	assert.Equal(t, collection.SourceTrade, owned.Source)
}

func TestIntegration_ConcurrentOppositeTransfers(t *testing.T) {
	const a, b collection.UserID = 900030, 900031
	ctx := t.Context()

	dbStore, err := storage.NewStore(ctx, testDBURL)
	require.NoError(t, err)
	store := pg.NewPool(dbStore)

	// Both rows exist before the race: concurrent transfers between existing
	// users must acquire their row locks in the same order, or Postgres
	// aborts one with a deadlock error.
	tx, err := store.Begin(ctx)
	require.NoError(t, err)
	for _, uid := range []collection.UserID{a, b} {
		require.NoError(t, tx.CreateUser(ctx, uid))
		_, err := tx.AddTokens(ctx, uid, tokenAmount(t, 1000))
		require.NoError(t, err)
	}
	require.NoError(t, tx.Commit(ctx))

	amount := tokenAmount(t, 1)

	before := make(map[collection.UserID]int32)
	for _, uid := range []collection.UserID{a, b} {
		user, err := store.GetUser(ctx, uid)
		require.NoError(t, err)
		before[uid] = user.Tokens
	}

	errs := make(chan error, 8)
	var wg sync.WaitGroup
	for i := range 8 {
		from, to := a, b
		if i%2 == 1 {
			from, to = b, a
		}
		wg.Add(1)
		go func(from, to collection.UserID) {
			defer wg.Done()
			for round := range 250 {
				if err := collection.TransferTokens(ctx, store, from, to, amount); err != nil {
					errs <- fmt.Errorf("transfer %d->%d (round %d): %w", from, to, round, err)
					return
				}
			}
		}(from, to)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}

	// The transfers net out, so each balance must be unchanged.
	for _, uid := range []collection.UserID{a, b} {
		user, err := store.GetUser(ctx, uid)
		require.NoError(t, err)
		assert.Equal(t, before[uid], user.Tokens, "tokens must be conserved")
	}
}
