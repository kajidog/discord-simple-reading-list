package store

import "sync"

// UserPreferences stores the emoji and presentation configuration for a user.
type UserPreferences struct {
	Emojis   []string
	Color    int
	HasColor bool
}

// EmojiStore provides thread-safe storage for user specific emoji preferences.
type EmojiStore struct {
	mu    sync.RWMutex
	prefs map[string]UserPreferences
}

// NewEmojiStore initializes an empty EmojiStore.
func NewEmojiStore() *EmojiStore {
	return &EmojiStore{
		prefs: make(map[string]UserPreferences),
	}
}

// Set associates emoji preferences with a given user ID.
func (s *EmojiStore) Set(userID string, prefs UserPreferences) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prefs[userID] = prefs
}

// Get retrieves the preferences associated with the user ID, if any.
func (s *EmojiStore) Get(userID string) (UserPreferences, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	prefs, ok := s.prefs[userID]
	return prefs, ok
}
