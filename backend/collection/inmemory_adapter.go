package collection

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
	return &collectionAdapter{charStore: s.store.CharacterStore()}
}

func (s *InMemoryStore) UserStore() userstore.Querier {
	return &userAdapter{userStore: s.store.UserStore()}
}

func (s *InMemoryStore) GuildStore() guildstore.Querier {
	return &guildAdapter{guildStore: s.store.GuildStore()}
}

func (s *InMemoryStore) WishlistStore() wishliststore.Querier {
	return &wishlistAdapter{wishlistStore: s.store.WishlistStore()}
}

func (s *InMemoryStore) CommandStore() commandstore.Querier {
	return &commandAdapter{cmdStore: s.store.CommandStore()}
}

func (s *InMemoryStore) DropStore() dropstore.Querier {
	return nil
}

func (s *InMemoryStore) InteractionStore() interactionstore.Querier {
	return nil
}

type collectionAdapter struct {
	charStore interfaces.CharacterRepository
}

func (a *collectionAdapter) Count(ctx context.Context, userID uint64) (int64, error) {
	return a.charStore.Count(ctx, userID)
}

func (a *collectionAdapter) Delete(ctx context.Context, arg collectionstore.DeleteParams) (collectionstore.Collection, error) {
	coll, err := a.charStore.Delete(ctx, arg.UserID, arg.CharacterID)
	if err != nil {
		return collectionstore.Collection{}, err
	}
	return collectionstore.Collection{
		UserID:      coll.UserID,
		CharacterID: coll.CharacterID,
		Source:      coll.Source,
		AcquiredAt:  pgtype.Timestamp{Time: coll.AcquiredAt, Valid: true},
	}, nil
}

func (a *collectionAdapter) Get(ctx context.Context, arg collectionstore.GetParams) (collectionstore.GetRow, error) {
	char, err := a.charStore.Get(ctx, arg.UserID, arg.ID)
	if err != nil {
		return collectionstore.GetRow{}, err
	}
	return collectionstore.GetRow{
		ID:         char.ID,
		Name:       char.Name,
		Image:      char.Image,
		MediaTitle: char.MediaTitle,
	}, nil
}

func (a *collectionAdapter) GetByID(ctx context.Context, id int64) (collectionstore.Character, error) {
	char, err := a.charStore.GetByID(ctx, id)
	if err != nil {
		return collectionstore.Character{}, err
	}
	return collectionstore.Character{
		ID:         char.ID,
		Name:       char.Name,
		Image:      char.Image,
		MediaTitle: char.MediaTitle,
	}, nil
}

func (a *collectionAdapter) Give(ctx context.Context, arg collectionstore.GiveParams) (collectionstore.Collection, error) {
	coll, err := a.charStore.Give(ctx, arg.UserID, arg.CharacterID, "TRADE")
	if err != nil {
		return collectionstore.Collection{}, err
	}
	return collectionstore.Collection{
		UserID:      coll.UserID,
		CharacterID: coll.CharacterID,
		Source:      coll.Source,
		AcquiredAt:  pgtype.Timestamp{Time: coll.AcquiredAt, Valid: true},
	}, nil
}

func (a *collectionAdapter) Insert(ctx context.Context, arg collectionstore.InsertParams) (collectionstore.Collection, error) {
	var source string
	if arg.Source != "" {
		source = arg.Source
	} else {
		source = "ROLL"
	}
	coll, err := a.charStore.Insert(ctx, arg.UserID, arg.CharacterID, source)
	if err != nil {
		return collectionstore.Collection{}, err
	}
	return collectionstore.Collection{
		UserID:      coll.UserID,
		CharacterID: coll.CharacterID,
		Source:      coll.Source,
		AcquiredAt:  pgtype.Timestamp{Time: coll.AcquiredAt, Valid: true},
	}, nil
}

func (a *collectionAdapter) List(ctx context.Context, userID uint64) ([]collectionstore.ListRow, error) {
	colls, err := a.charStore.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]collectionstore.ListRow, len(colls))
	for i, c := range colls {
		char, _ := a.charStore.GetByID(ctx, c.CharacterID)
		result[i] = collectionstore.ListRow{
			ID:         char.ID,
			Name:       char.Name,
			Image:      char.Image,
			MediaTitle: char.MediaTitle,
			Source:     c.Source,
			Date:       pgtype.Timestamp{Time: c.AcquiredAt, Valid: true},
		}
	}
	return result, nil
}

func (a *collectionAdapter) ListIDs(ctx context.Context, userID uint64) ([]int64, error) {
	return a.charStore.ListIDs(ctx, userID)
}

func (a *collectionAdapter) ListPaginated(ctx context.Context, userID uint64) ([]collectionstore.ListPaginatedRow, error) {
	return nil, nil
}

func (a *collectionAdapter) SearchCharacters(ctx context.Context, arg collectionstore.SearchCharactersParams) ([]collectionstore.SearchCharactersRow, error) {
	chars, err := a.charStore.SearchCharacters(ctx, arg.UserID, arg.Term, int(arg.Lim))
	if err != nil {
		return nil, err
	}
	result := make([]collectionstore.SearchCharactersRow, len(chars))
	for i, c := range chars {
		result[i] = collectionstore.SearchCharactersRow{
			ID:    c.ID,
			Name:  c.Name,
			Image: c.Image,
		}
	}
	return result, nil
}

func (a *collectionAdapter) SearchGlobalCharacters(ctx context.Context, arg collectionstore.SearchGlobalCharactersParams) ([]collectionstore.Character, error) {
	chars, err := a.charStore.SearchGlobalCharacters(ctx, arg.Term, int(arg.Lim))
	if err != nil {
		return nil, err
	}
	result := make([]collectionstore.Character, len(chars))
	for i, c := range chars {
		result[i] = collectionstore.Character{
			ID:         c.ID,
			Name:       c.Name,
			Image:      c.Image,
			MediaTitle: c.MediaTitle,
		}
	}
	return result, nil
}

func (a *collectionAdapter) UpdateImageName(ctx context.Context, arg collectionstore.UpdateImageNameParams) (collectionstore.Character, error) {
	char, err := a.charStore.UpdateImageName(ctx, arg.ID, arg.Name, arg.Image)
	if err != nil {
		return collectionstore.Character{}, err
	}
	return collectionstore.Character{
		ID:         char.ID,
		Name:       char.Name,
		Image:      char.Image,
		MediaTitle: char.MediaTitle,
	}, nil
}

func (a *collectionAdapter) UpsertCharacter(ctx context.Context, arg collectionstore.UpsertCharacterParams) (collectionstore.Character, error) {
	char, err := a.charStore.UpsertCharacter(ctx, interfaces.Character{
		ID:    arg.ID,
		Name:  arg.Name,
		Image: arg.Image,
	})
	if err != nil {
		return collectionstore.Character{}, err
	}
	return collectionstore.Character{
		ID:         char.ID,
		Name:       char.Name,
		Image:      char.Image,
		MediaTitle: char.MediaTitle,
	}, nil
}

func (a *collectionAdapter) UsersOwningCharFiltered(ctx context.Context, arg collectionstore.UsersOwningCharFilteredParams) ([]uint64, error) {
	userIDs := make([]uint64, len(arg.UserIds))
	for i, id := range arg.UserIds {
		userIDs[i] = uint64(id)
	}
	return a.charStore.UsersOwningCharFiltered(ctx, arg.CharacterID, userIDs)
}

type userAdapter struct {
	userStore interfaces.UserRepository
}

func (a *userAdapter) CountFiltered(ctx context.Context, arg userstore.CountFilteredParams) (int64, error) {
	return a.userStore.CountFiltered(ctx, arg.UsernamePrefix)
}

func (a *userAdapter) Create(ctx context.Context, userID uint64) error {
	return a.userStore.Create(ctx, userID)
}

func (a *userAdapter) Get(ctx context.Context, userID uint64) (userstore.User, error) {
	u, err := a.userStore.Get(ctx, userID)
	if err != nil {
		return userstore.User{}, err
	}
	return toUserStoreUser(u), nil
}

func (a *userAdapter) GetByAnilist(ctx context.Context, lower string) (userstore.User, error) {
	u, err := a.userStore.GetByAnilist(ctx, lower)
	if err != nil {
		return userstore.User{}, err
	}
	return toUserStoreUser(u), nil
}

func (a *userAdapter) GetByDiscordUsername(ctx context.Context, discordUsername string) (userstore.User, error) {
	u, err := a.userStore.GetByDiscordUsername(ctx, discordUsername)
	if err != nil {
		return userstore.User{}, err
	}
	return toUserStoreUser(u), nil
}

func (a *userAdapter) List(ctx context.Context, arg userstore.ListParams) ([]userstore.User, error) {
	users, err := a.userStore.List(ctx, int(arg.PageSize), int(arg.PageOffset))
	if err != nil {
		return nil, err
	}
	result := make([]userstore.User, len(users))
	for i, u := range users {
		result[i] = toUserStoreUser(u)
	}
	return result, nil
}

func (a *userAdapter) UpdateAnilistURL(ctx context.Context, arg userstore.UpdateAnilistURLParams) error {
	return a.userStore.UpdateAnilistURL(ctx, arg.UserID, arg.AnilistUrl)
}

func (a *userAdapter) UpdateDate(ctx context.Context, arg userstore.UpdateDateParams) error {
	return nil
}

func (a *userAdapter) UpdateDiscordInfo(ctx context.Context, arg userstore.UpdateDiscordInfoParams) error {
	return a.userStore.UpdateDiscordInfo(ctx, arg.UserID, arg.DiscordUsername, arg.DiscordAvatar)
}

func (a *userAdapter) UpdateFavorite(ctx context.Context, arg userstore.UpdateFavoriteParams) error {
	var fav int64
	if arg.Favorite.Valid {
		fav = arg.Favorite.Int64
	}
	return a.userStore.UpdateFavorite(ctx, arg.UserID, fav)
}

func (a *userAdapter) UpdateQuote(ctx context.Context, arg userstore.UpdateQuoteParams) error {
	return a.userStore.UpdateQuote(ctx, arg.UserID, arg.Quote)
}

func (a *userAdapter) UpdateTokens(ctx context.Context, arg userstore.UpdateTokensParams) (userstore.User, error) {
	u, err := a.userStore.UpdateTokens(ctx, arg.UserID, int(arg.Tokens))
	if err != nil {
		return userstore.User{}, err
	}
	return toUserStoreUser(u), nil
}

func toUserStoreUser(u interfaces.User) userstore.User {
	return userstore.User{
		ID:              int32(u.ID),
		UserID:          u.UserID,
		Quote:           u.Quote,
		Date:            pgtype.Timestamp{Time: u.Date, Valid: true},
		Favorite:        pgtype.Int8{Int64: u.Favorite, Valid: u.Favorite != 0},
		Tokens:          int32(u.Tokens),
		AnilistUrl:      u.AnilistURL,
		DiscordUsername: u.DiscordUsername,
		DiscordAvatar:   u.DiscordAvatar,
		LastUpdated:     pgtype.Timestamp{Time: u.LastUpdated, Valid: true},
	}
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
	status, _, err := a.guildStore.IsGuildIndexed(ctx, guildID)
	var idxStatus guildstore.IndexingStatus
	if status {
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

type wishlistAdapter struct {
	wishlistStore interfaces.WishlistRepository
}

func (a *wishlistAdapter) AddCharacterToWishlist(ctx context.Context, arg wishliststore.AddCharacterToWishlistParams) error {
	return a.wishlistStore.AddCharacter(ctx, arg.UserID, arg.CharacterID)
}

func (a *wishlistAdapter) AddMultipleCharactersToWishlist(ctx context.Context, arg wishliststore.AddMultipleCharactersToWishlistParams) error {
	return a.wishlistStore.AddMultipleCharacters(ctx, arg.UserID, arg.Column2)
}

func (a *wishlistAdapter) CompareWithUser(ctx context.Context, arg wishliststore.CompareWithUserParams) ([]wishliststore.CompareWithUserRow, error) {
	comp, err := a.wishlistStore.CompareWithUser(ctx, arg.UserID, arg.UserID_2)
	if err != nil {
		return nil, err
	}
	var result []wishliststore.CompareWithUserRow
	for _, id := range comp.UserHas {
		result = append(result, wishliststore.CompareWithUserRow{
			Type: "has",
			ID:   id,
		})
	}
	for _, id := range comp.UserWants {
		result = append(result, wishliststore.CompareWithUserRow{
			Type: "wants",
			ID:   id,
		})
	}
	return result, nil
}

func (a *wishlistAdapter) GetUserCharacterWishlist(ctx context.Context, userID uint64) ([]wishliststore.GetUserCharacterWishlistRow, error) {
	ids, err := a.wishlistStore.GetUserWishlist(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]wishliststore.GetUserCharacterWishlistRow, len(ids))
	for i, id := range ids {
		result[i] = wishliststore.GetUserCharacterWishlistRow{
			ID: id,
		}
	}
	return result, nil
}

func (a *wishlistAdapter) GetWantedCharacters(ctx context.Context, arg wishliststore.GetWantedCharactersParams) ([]wishliststore.GetWantedCharactersRow, error) {
	return nil, nil
}

func (a *wishlistAdapter) GetWishlistHolders(ctx context.Context, arg wishliststore.GetWishlistHoldersParams) ([]wishliststore.GetWishlistHoldersRow, error) {
	charIDs := make([]int64, len(arg.Column1))
	for i, id := range arg.Column1 {
		charIDs[i] = id
	}
	holders, err := a.wishlistStore.GetWishlistHolders(ctx, charIDs)
	if err != nil {
		return nil, err
	}
	var result []wishliststore.GetWishlistHoldersRow
	for charID, userIDs := range holders {
		for _, userID := range userIDs {
			result = append(result, wishliststore.GetWishlistHoldersRow{
				CharacterID: charID,
				UserID:      int64(userID),
			})
		}
	}
	return result, nil
}

func (a *wishlistAdapter) RemoveAllFromWishlist(ctx context.Context, userID uint64) error {
	return a.wishlistStore.RemoveAll(ctx, userID)
}

func (a *wishlistAdapter) RemoveCharacterFromWishlist(ctx context.Context, arg wishliststore.RemoveCharacterFromWishlistParams) error {
	return a.wishlistStore.RemoveCharacter(ctx, arg.UserID, arg.CharacterID)
}

func (a *wishlistAdapter) RemoveMultipleCharactersFromWishlist(ctx context.Context, arg wishliststore.RemoveMultipleCharactersFromWishlistParams) error {
	return a.wishlistStore.RemoveMultipleCharacters(ctx, arg.UserID, arg.Column2)
}

type commandAdapter struct {
	cmdStore interfaces.CommandRepository
}

func (a *commandAdapter) GetCommandHash(ctx context.Context) (string, error) {
	return a.cmdStore.GetCommandHash(ctx)
}

func (a *commandAdapter) SetCommandHash(ctx context.Context, hash string) error {
	return a.cmdStore.SetCommandHash(ctx, hash)
}

func (a *commandAdapter) UpdateCommandHash(ctx context.Context, hash string) error {
	return a.cmdStore.SetCommandHash(ctx, hash)
}

type InMemoryAnimeService struct{}

func NewInMemoryAnimeService() *InMemoryAnimeService {
	return &InMemoryAnimeService{}
}

func (s *InMemoryAnimeService) RandomChar(ctx context.Context, notIn ...int64) (MediaCharacter, error) {
	return MediaCharacter{}, nil
}

func (s *InMemoryAnimeService) Anime(ctx context.Context, name string) ([]Media, error) {
	return nil, nil
}

func (s *InMemoryAnimeService) Manga(ctx context.Context, name string) ([]Media, error) {
	return nil, nil
}

func (s *InMemoryAnimeService) User(ctx context.Context, name string) ([]TrackerUser, error) {
	return nil, nil
}

func (s *InMemoryAnimeService) Character(ctx context.Context, name string) ([]MediaCharacter, error) {
	return nil, nil
}
