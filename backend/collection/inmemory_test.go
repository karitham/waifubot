package collection

import (
	"context"
	"testing"

	"github.com/Karitham/corde"

	"github.com/karitham/waifubot/storage/collectionstore"
)

func TestInMemoryStore_UserProfile(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()
	userID := corde.Snowflake(123)

	store.UserStore().Create(ctx, uint64(userID))

	_, err := UserProfile(ctx, store, userID)
	if err != nil {
		t.Fatalf("UserProfile() error = %v", err)
	}

	user, err := store.UserStore().Get(ctx, uint64(userID))
	if err != nil {
		t.Fatalf("UserStore().Get() error = %v", err)
	}
	if user.UserID != uint64(userID) {
		t.Errorf("UserID = %v, want %v", user.UserID, uint64(userID))
	}
}

func TestInMemoryStore_SetQuote(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()
	userID := corde.Snowflake(123)
	quote := "Hello World"

	store.UserStore().Create(ctx, uint64(userID))

	err := SetQuote(ctx, store, userID, quote)
	if err != nil {
		t.Fatalf("SetQuote() error = %v", err)
	}

	user, err := store.UserStore().Get(ctx, uint64(userID))
	if err != nil {
		t.Fatalf("UserStore().Get() error = %v", err)
	}
	if user.Quote != quote {
		t.Errorf("Quote = %v, want %v", user.Quote, quote)
	}
}

func TestInMemoryStore_SetFavorite(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()
	userID := corde.Snowflake(123)
	charID := int64(456)

	store.UserStore().Create(ctx, uint64(userID))

	err := SetFavorite(ctx, store, userID, charID)
	if err != nil {
		t.Fatalf("SetFavorite() error = %v", err)
	}

	user, err := store.UserStore().Get(ctx, uint64(userID))
	if err != nil {
		t.Fatalf("UserStore().Get() error = %v", err)
	}
	if !user.Favorite.Valid || user.Favorite.Int64 != charID {
		t.Errorf("Favorite = %v, want %v", user.Favorite, charID)
	}
}

func TestInMemoryStore_AddCharacter(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()
	userID := uint64(123)
	charID := int64(456)

	char, err := store.CollectionStore().UpsertCharacter(ctx, collectionstore.UpsertCharacterParams{
		ID:    charID,
		Name:  "Test Character",
		Image: "test.jpg",
	})
	if err != nil {
		t.Fatalf("UpsertCharacter() error = %v", err)
	}

	_, err = store.CollectionStore().Insert(ctx, collectionstore.InsertParams{
		UserID:      userID,
		CharacterID: char.ID,
		Source:      "ROLL",
	})
	if err != nil {
		t.Fatalf("Insert() error = %v", err)
	}

	count, err := store.CollectionStore().Count(ctx, userID)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 1 {
		t.Errorf("Count = %v, want 1", count)
	}
}

func TestInMemoryStore_ListCharacters(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()
	userID := uint64(123)

	for i := int64(1); i <= 3; i++ {
		store.CollectionStore().UpsertCharacter(ctx, collectionstore.UpsertCharacterParams{
			ID:    i,
			Name:  "Char",
			Image: "img",
		})
		store.CollectionStore().Insert(ctx, collectionstore.InsertParams{
			UserID:      userID,
			CharacterID: i,
			Source:      "ROLL",
		})
	}

	chars, err := Characters(ctx, store, corde.Snowflake(userID))
	if err != nil {
		t.Fatalf("Characters() error = %v", err)
	}
	if len(chars) != 3 {
		t.Errorf("len(chars) = %v, want 3", len(chars))
	}
}

func TestInMemoryStore_CheckOwnership(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()
	userID := uint64(123)
	charID := int64(456)

	store.CollectionStore().UpsertCharacter(ctx, collectionstore.UpsertCharacterParams{
		ID:    charID,
		Name:  "Test Char",
		Image: "test.jpg",
	})

	store.CollectionStore().Insert(ctx, collectionstore.InsertParams{
		UserID:      userID,
		CharacterID: charID,
		Source:      "ROLL",
	})

	owns, char, err := CheckOwnership(ctx, store, corde.Snowflake(userID), charID)
	if err != nil {
		t.Fatalf("CheckOwnership() error = %v", err)
	}
	if !owns {
		t.Error("expected to own character")
	}
	if char.ID != charID {
		t.Errorf("char.ID = %v, want %v", char.ID, charID)
	}
}

func TestInMemoryStore_CheckOwnership_NotOwned(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()
	userID := uint64(123)
	charID := int64(456)

	store.CollectionStore().UpsertCharacter(ctx, collectionstore.UpsertCharacterParams{
		ID:    charID,
		Name:  "Test Char",
		Image: "test.jpg",
	})
	store.UserStore().Create(ctx, userID)

	_, _, err := CheckOwnership(ctx, store, corde.Snowflake(userID), charID)
	if err == nil {
		t.Error("expected error when character not owned")
	}
}

func TestInMemoryStore_SearchCharacters(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()
	userID := uint64(123)

	store.CollectionStore().UpsertCharacter(ctx, collectionstore.UpsertCharacterParams{
		ID:    1,
		Name:  "Sailor Moon",
		Image: "moon.jpg",
	})
	store.CollectionStore().UpsertCharacter(ctx, collectionstore.UpsertCharacterParams{
		ID:    2,
		Name:  "Naruto Uzumaki",
		Image: "naruto.jpg",
	})
	store.CollectionStore().UpsertCharacter(ctx, collectionstore.UpsertCharacterParams{
		ID:    3,
		Name:  "Goku",
		Image: "goku.jpg",
	})

	store.CollectionStore().Insert(ctx, collectionstore.InsertParams{UserID: userID, CharacterID: 1, Source: "ROLL"})
	store.CollectionStore().Insert(ctx, collectionstore.InsertParams{UserID: userID, CharacterID: 2, Source: "ROLL"})

	chars, err := SearchCharacters(ctx, store, corde.Snowflake(userID), "Sailor")
	if err != nil {
		t.Fatalf("SearchCharacters() error = %v", err)
	}
	if len(chars) != 1 {
		t.Errorf("len(chars) = %v, want 1", len(chars))
	}
}

func TestInMemoryStore_SearchGlobalCharacters(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	store.CollectionStore().UpsertCharacter(ctx, collectionstore.UpsertCharacterParams{
		ID:    1,
		Name:  "Sailor Moon",
		Image: "moon.jpg",
	})
	store.CollectionStore().UpsertCharacter(ctx, collectionstore.UpsertCharacterParams{
		ID:    2,
		Name:  "Naruto",
		Image: "naruto.jpg",
	})

	chars, err := SearchGlobalCharacters(ctx, store, "Moon")
	if err != nil {
		t.Fatalf("SearchGlobalCharacters() error = %v", err)
	}
	if len(chars) != 1 {
		t.Errorf("len(chars) = %v, want 1", len(chars))
	}
	if chars[0].Name != "Sailor Moon" {
		t.Errorf("Name = %v, want Sailor Moon", chars[0].Name)
	}
}
