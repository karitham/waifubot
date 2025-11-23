package guild

//go:generate mockgen -source=../storage/guildstore/querier.go -destination=mocks/guildstore_mock.go -package=mocks -mock_names=Querier=MockGuildQuerier

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Karitham/corde"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/karitham/waifubot/storage"
	"github.com/karitham/waifubot/storage/guildstore"
)

// Store defines the interface for guild operations
type Store interface {
	GuildStore() guildstore.Querier
	Tx(ctx context.Context) (storage.Store, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// Service defines the interface for guild operations
type Service interface {
	IndexGuildIfNeeded(ctx context.Context, guildID corde.Snowflake) error
}

const maxAge = 7 * 24 * time.Hour

// Indexer handles guild indexing operations
type Indexer struct {
	store      Store
	botToken   string
	httpClient *http.Client
}

// NewIndexer creates a new guild indexer
func NewIndexer(store Store, botToken string) *Indexer {
	return &Indexer{
		store:      store,
		botToken:   botToken,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// IndexGuildIfNeeded checks if a guild needs indexing and starts it if necessary
func (i *Indexer) IndexGuildIfNeeded(ctx context.Context, guildID corde.Snowflake) error {
	row, err := i.store.GuildStore().IsGuildIndexed(ctx, uint64(guildID))
	if err != nil {
		return err
	}

	indexed := row.Status == guildstore.IndexingStatusCompleted && time.Since(row.UpdatedAt.Time) <= maxAge
	if indexed {
		return nil
	}

	var shouldStart bool
	txI, err := i.store.Tx(ctx)
	if err != nil {
		return err
	}
	tx := txI.(Store)

	row, err = tx.GuildStore().IsGuildIndexed(ctx, uint64(guildID))
	if err != nil {
		shouldStart = true
		err = tx.GuildStore().StartIndexingJob(ctx, uint64(guildID))
	} else {
		status := row.Status
		updatedAt := row.UpdatedAt
		if status == guildstore.IndexingStatusInProgress {
			if time.Since(updatedAt.Time) < 10*time.Minute {
				shouldStart = false
				err = nil
			} else {
				shouldStart = true
				err = tx.GuildStore().StartIndexingJob(ctx, uint64(guildID))
			}
		} else {
			shouldStart = true
			err = tx.GuildStore().StartIndexingJob(ctx, uint64(guildID))
		}
	}
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	if !shouldStart {
		return nil
	}

	_ = i.IndexGuild(ctx, guildID)
	return nil
}

// IndexGuild performs the actual guild indexing
func (i *Indexer) IndexGuild(ctx context.Context, guildID corde.Snowflake) error {
	memberIDs, err := i.fetchGuildMemberIDs(ctx, guildID)
	if err != nil {
		return err
	}

	// Convert memberIDs to []int64
	userIDsInt := make([]int64, len(memberIDs))
	for idx, id := range memberIDs {
		userIDsInt[idx] = int64(id)
	}

	// Delete members not in the new list
	err = i.store.GuildStore().DeleteGuildMembersNotIn(ctx, guildstore.DeleteGuildMembersNotInParams{
		GuildID: uint64(guildID),
		Column2: userIDsInt,
	})
	if err != nil {
		return err
	}

	// Upsert new members
	if len(memberIDs) > 0 {
		err = i.store.GuildStore().UpsertGuildMembers(ctx, guildstore.UpsertGuildMembersParams{
			GuildID:   uint64(guildID),
			Column2:   userIDsInt,
			IndexedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		})
		if err != nil {
			return err
		}
	}

	err = i.store.GuildStore().CompleteIndexingJob(ctx, uint64(guildID))
	if err != nil {
		return err
	}

	return nil
}

// fetchGuildMemberIDs fetches all member IDs from a Discord guild
func (i *Indexer) fetchGuildMemberIDs(ctx context.Context, guildID corde.Snowflake) ([]corde.Snowflake, error) {
	var allMemberIDs []corde.Snowflake

	after := corde.Snowflake(0)

	for {
		members, err := i.fetchGuildMembersPage(ctx, guildID, after)
		if err != nil {
			return nil, err
		}

		if len(members) == 0 {
			break
		}

		for _, member := range members {
			allMemberIDs = append(allMemberIDs, member.User.ID)
		}

		if len(members) < 1000 {
			break
		}

		after = members[len(members)-1].User.ID
	}

	return allMemberIDs, nil
}

// fetchGuildMembersPage fetches a page of guild members
func (i *Indexer) fetchGuildMembersPage(ctx context.Context, guildID, after corde.Snowflake) ([]corde.Member, error) {
	url := fmt.Sprintf("https://discord.com/api/v10/guilds/%d/members?limit=1000", guildID)
	if after != 0 {
		url = fmt.Sprintf("%s&after=%d", url, after)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bot "+i.botToken)

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch guild members: %w", err)
	}
	defer resp.Body.Close() //nolint: errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch guild members: status %d", resp.StatusCode)
	}

	var members []corde.Member
	if err := json.NewDecoder(resp.Body).Decode(&members); err != nil {
		return nil, fmt.Errorf("failed to decode members: %w", err)
	}

	return members, nil
}
