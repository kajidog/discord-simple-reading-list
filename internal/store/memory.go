package store

import "sync"

// EmojiStore provides thread-safe storage for user specific emoji preferences.
type EmojiStore struct {
	mu     sync.RWMutex
	emojis map[string]string
}

// NewEmojiStore initializes an empty EmojiStore.
func NewEmojiStore() *EmojiStore {
	return &EmojiStore{
		emojis: make(map[string]string),
	}
}

// Set associates an emoji with a given user ID.
func (s *EmojiStore) Set(userID, emoji string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.emojis[userID] = emoji
}

// Get retrieves the emoji associated with the user ID, if any.
func (s *EmojiStore) Get(userID string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	emoji, ok := s.emojis[userID]
	return emoji, ok
}
