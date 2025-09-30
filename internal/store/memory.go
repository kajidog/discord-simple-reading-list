package store

import "sync"

// BookmarkMode identifies how a bookmarked message should be formatted.
type BookmarkMode string

const (
	// ModeLightweight stores a condensed embed focused on quick review.
	ModeLightweight BookmarkMode = "lightweight"
	// ModeComplete stores all available information from the original message.
	ModeComplete BookmarkMode = "complete"
	// ModeBalanced stores a balanced view between lightweight and complete.
	ModeBalanced BookmarkMode = "balanced"
)

// EmojiPreference stores configuration for a specific emoji bookmark.
type EmojiPreference struct {
	Mode     BookmarkMode
	Color    int
	HasColor bool
}

// UserPreferences stores the emoji and presentation configuration for a user.
type UserPreferences struct {
	Emojis map[string]EmojiPreference
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

// SetEmoji associates emoji preferences with a given user ID and emoji.
func (s *EmojiStore) SetEmoji(userID, emoji string, pref EmojiPreference) {
	s.mu.Lock()
	defer s.mu.Unlock()
	userPrefs, ok := s.prefs[userID]
	if !ok || userPrefs.Emojis == nil {
		userPrefs = UserPreferences{Emojis: make(map[string]EmojiPreference)}
	}

	userPrefs.Emojis[emoji] = pref
	s.prefs[userID] = userPrefs
}

// Get retrieves the preferences associated with the user ID, if any.
func (s *EmojiStore) Get(userID string) (UserPreferences, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	prefs, ok := s.prefs[userID]
	if !ok {
		return UserPreferences{}, false
	}

	// Ensure the map is non-nil for consumers.
	if prefs.Emojis == nil {
		prefs.Emojis = make(map[string]EmojiPreference)
	}

	return prefs, true
}

// GetEmoji retrieves a specific emoji preference for the given user.
func (s *EmojiStore) GetEmoji(userID, emoji string) (EmojiPreference, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prefs, ok := s.prefs[userID]
	if !ok {
		return EmojiPreference{}, false
	}

	pref, ok := prefs.Emojis[emoji]
	return pref, ok
}
