package inmemory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/karitham/waifubot/storage/interfaces"
)

type CharacterStore struct {
	mu          sync.RWMutex
	characters  map[int64]interfaces.Character
	collections map[uint64]map[int64]interfaces.Collection
}

func NewCharacterStore() *CharacterStore {
	return &CharacterStore{
		characters:  make(map[int64]interfaces.Character),
		collections: make(map[uint64]map[int64]interfaces.Collection),
	}
}

func (s *CharacterStore) Count(ctx context.Context, userID uint64) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userCols, ok := s.collections[userID]
	if !ok {
		return 0, nil
	}
	return int64(len(userCols)), nil
}

func (s *CharacterStore) Get(ctx context.Context, userID uint64, charID int64) (interfaces.Character, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userCols, ok := s.collections[userID]
	if !ok {
		return interfaces.Character{}, ErrNotFound
	}

	if _, ok := userCols[charID]; !ok {
		return interfaces.Character{}, ErrNotFound
	}

	char, ok := s.characters[charID]
	if !ok {
		return interfaces.Character{}, ErrNotFound
	}

	return interfaces.Character{
		ID:         char.ID,
		Name:       char.Name,
		Image:      char.Image,
		MediaTitle: char.MediaTitle,
	}, nil
}

func (s *CharacterStore) GetByID(ctx context.Context, charID int64) (interfaces.Character, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	char, ok := s.characters[charID]
	if !ok {
		return interfaces.Character{}, ErrNotFound
	}
	return char, nil
}

func (s *CharacterStore) Give(ctx context.Context, userID uint64, charID int64, source string) (interfaces.Collection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.insertCollection(userID, charID, source)
}

func (s *CharacterStore) Insert(ctx context.Context, userID uint64, charID int64, source string) (interfaces.Collection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	userCols, ok := s.collections[userID]
	if ok {
		if _, exists := userCols[charID]; exists {
			return interfaces.Collection{}, ErrAlreadyExists
		}
	}

	return s.insertCollection(userID, charID, source)
}

func (s *CharacterStore) insertCollection(userID uint64, charID int64, source string) (interfaces.Collection, error) {
	if _, ok := s.characters[charID]; !ok {
		return interfaces.Collection{}, ErrNotFound
	}

	col := interfaces.Collection{
		UserID:      userID,
		CharacterID: charID,
		Source:      source,
		AcquiredAt:  time.Now(),
	}

	if s.collections[userID] == nil {
		s.collections[userID] = make(map[int64]interfaces.Collection)
	}
	s.collections[userID][charID] = col
	return col, nil
}

func (s *CharacterStore) Delete(ctx context.Context, userID uint64, charID int64) (interfaces.Collection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	userCols, ok := s.collections[userID]
	if !ok {
		return interfaces.Collection{}, ErrNotFound
	}

	col, ok := userCols[charID]
	if !ok {
		return interfaces.Collection{}, ErrNotFound
	}

	delete(userCols, charID)
	return col, nil
}

func (s *CharacterStore) List(ctx context.Context, userID uint64) ([]interfaces.Collection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userCols, ok := s.collections[userID]
	if !ok {
		return []interfaces.Collection{}, nil
	}

	cols := make([]interfaces.Collection, 0, len(userCols))
	for _, col := range userCols {
		cols = append(cols, col)
	}
	sort.Slice(cols, func(i, j int) bool {
		return cols[i].AcquiredAt.After(cols[j].AcquiredAt)
	})
	return cols, nil
}

func (s *CharacterStore) ListIDs(ctx context.Context, userID uint64) ([]int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userCols, ok := s.collections[userID]
	if !ok {
		return []int64{}, nil
	}

	ids := make([]int64, 0, len(userCols))
	for id := range userCols {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})
	return ids, nil
}

func (s *CharacterStore) ListPaginated(ctx context.Context, userID uint64, opts interfaces.ListOptions) ([]interfaces.Collection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userCols, ok := s.collections[userID]
	if !ok {
		return []interfaces.Collection{}, nil
	}

	cols := make([]interfaces.Collection, 0, len(userCols))
	for _, col := range userCols {
		cols = append(cols, col)
	}

	sort.Slice(cols, func(i, j int) bool {
		switch opts.OrderBy {
		case "name":
			c1, _ := s.characters[cols[i].CharacterID]
			c2, _ := s.characters[cols[j].CharacterID]
			if opts.Direction == "asc" {
				return c1.Name < c2.Name
			}
			return c1.Name > c2.Name
		case "anilist_id":
			if opts.Direction == "asc" {
				return cols[i].CharacterID < cols[j].CharacterID
			}
			return cols[i].CharacterID > cols[j].CharacterID
		default:
			if opts.Direction == "asc" {
				return cols[i].AcquiredAt.Before(cols[j].AcquiredAt)
			}
			return cols[i].AcquiredAt.After(cols[j].AcquiredAt)
		}
	})

	if opts.Offset >= len(cols) {
		return []interfaces.Collection{}, nil
	}
	if opts.Offset+opts.Limit > len(cols) {
		opts.Limit = len(cols) - opts.Offset
	}
	return cols[opts.Offset : opts.Offset+opts.Limit], nil
}

func (s *CharacterStore) SearchCharacters(ctx context.Context, userID uint64, search string, limit int) ([]interfaces.Character, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userCols, ok := s.collections[userID]
	if !ok {
		return []interfaces.Character{}, nil
	}

	var results []interfaces.Character
	for charID := range userCols {
		char, ok := s.characters[charID]
		if !ok {
			continue
		}
		if search == "" || containsIgnoreCase(char.Name, search) {
			results = append(results, char)
		}
		if len(results) >= limit {
			break
		}
	}
	return results, nil
}

func (s *CharacterStore) SearchGlobalCharacters(ctx context.Context, search string, limit int) ([]interfaces.Character, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []interfaces.Character
	for _, char := range s.characters {
		if search == "" || containsIgnoreCase(char.Name, search) {
			results = append(results, char)
		}
		if len(results) >= limit {
			break
		}
	}
	return results, nil
}

func (s *CharacterStore) UpsertCharacter(ctx context.Context, char interfaces.Character) (interfaces.Character, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.characters[char.ID] = char
	return char, nil
}

func (s *CharacterStore) UpdateImageName(ctx context.Context, charID int64, name, image string) (interfaces.Character, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	char, ok := s.characters[charID]
	if !ok {
		return interfaces.Character{}, ErrNotFound
	}
	char.Name = name
	char.Image = image
	s.characters[charID] = char
	return char, nil
}

func (s *CharacterStore) UsersOwningCharFiltered(ctx context.Context, charID int64, userIDs []uint64) ([]uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var owners []uint64
	for _, uid := range userIDs {
		userCols, ok := s.collections[uid]
		if !ok {
			continue
		}
		if _, exists := userCols[charID]; exists {
			owners = append(owners, uid)
		}
	}
	return owners, nil
}

func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
