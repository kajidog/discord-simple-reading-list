package store

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/example/discord-bookmark-manager/internal/reminders"
)

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

// DestinationType identifies where the bookmark should be delivered.
type DestinationType string

const (
	// DestinationDM sends the bookmark as a direct message.
	DestinationDM DestinationType = "dm"
	// DestinationChannel sends the bookmark to a guild text channel.
	DestinationChannel DestinationType = "channel"
)

// EmojiPreference stores configuration for a specific emoji bookmark.
type EmojiPreference struct {
	Mode        BookmarkMode          `json:"mode"`
	Color       int                   `json:"color"`
	HasColor    bool                  `json:"hasColor"`
	Reminder    *reminders.Preference `json:"reminder,omitempty"`
	Destination DestinationType       `json:"destination,omitempty"`
	ChannelID   string                `json:"channelId,omitempty"`
}

func normalizeEmojiPreference(pref EmojiPreference) EmojiPreference {
	if pref.Destination == "" {
		pref.Destination = DestinationDM
	}

	if pref.Destination != DestinationChannel {
		pref.ChannelID = ""
	}

	return pref
}

// UserPreferences stores the emoji and presentation configuration for a user.
type UserPreferences struct {
	Emojis map[string]EmojiPreference `json:"emojis"`
}

// EmojiStore provides thread-safe storage for user specific emoji preferences.
type EmojiStore struct {
	mu       sync.RWMutex
	prefs    map[string]UserPreferences
	filePath string
}

// NewEmojiStore initializes an EmojiStore and loads any persisted data from filePath.
//
// If filePath is empty, the store behaves as an in-memory only store.
func NewEmojiStore(filePath string) (*EmojiStore, error) {
	store := &EmojiStore{
		prefs:    make(map[string]UserPreferences),
		filePath: filePath,
	}

	if filePath == "" {
		return store, nil
	}

	if err := store.load(); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return store, nil
		}
		return nil, err
	}

	return store, nil
}

// SetEmoji associates emoji preferences with a given user ID and emoji.
func (s *EmojiStore) SetEmoji(userID, emoji string, pref EmojiPreference) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	userPrefs, ok := s.prefs[userID]

	var current map[string]EmojiPreference
	if !ok || userPrefs.Emojis == nil {
		current = make(map[string]EmojiPreference)
	} else {
		current = userPrefs.Emojis
	}

	next := make(map[string]EmojiPreference, len(current)+1)
	for key, value := range current {
		next[key] = normalizeEmojiPreference(value)
	}
	next[emoji] = normalizeEmojiPreference(pref)

	previous := userPrefs
	s.prefs[userID] = UserPreferences{Emojis: next}

	if err := s.saveLocked(); err != nil {
		if ok {
			s.prefs[userID] = previous
		} else {
			delete(s.prefs, userID)
		}
		return err
	}

	return nil
}

// DeleteEmoji removes an emoji preference for the given user ID. It returns true when a
// preference was removed.
func (s *EmojiStore) DeleteEmoji(userID, emoji string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	userPrefs, ok := s.prefs[userID]
	if !ok || len(userPrefs.Emojis) == 0 {
		return false, nil
	}

	if _, exists := userPrefs.Emojis[emoji]; !exists {
		return false, nil
	}

	previous := userPrefs

	next := make(map[string]EmojiPreference, len(userPrefs.Emojis)-1)
	for key, value := range userPrefs.Emojis {
		if key == emoji {
			continue
		}
		next[key] = value
	}

	if len(next) == 0 {
		delete(s.prefs, userID)
	} else {
		s.prefs[userID] = UserPreferences{Emojis: next}
	}

	if err := s.saveLocked(); err != nil {
		s.prefs[userID] = previous
		return false, err
	}

	return true, nil
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
	} else {
		normalized := make(map[string]EmojiPreference, len(prefs.Emojis))
		for emoji, pref := range prefs.Emojis {
			normalized[emoji] = normalizeEmojiPreference(pref)
		}
		prefs.Emojis = normalized
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
	if !ok {
		return EmojiPreference{}, false
	}

	return normalizeEmojiPreference(pref), true
}

func (s *EmojiStore) load() error {
	file, err := os.Open(s.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var persisted map[string]UserPreferences
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&persisted); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	for userID, prefs := range persisted {
		if prefs.Emojis == nil {
			prefs.Emojis = make(map[string]EmojiPreference)
		}

		normalized := make(map[string]EmojiPreference, len(prefs.Emojis))
		for emoji, pref := range prefs.Emojis {
			normalized[emoji] = normalizeEmojiPreference(pref)
		}
		prefs.Emojis = normalized
		s.prefs[userID] = prefs
	}

	return nil
}

func (s *EmojiStore) saveLocked() error {
	if s.filePath == "" {
		return nil
	}

	toPersist := make(map[string]UserPreferences, len(s.prefs))
	for userID, prefs := range s.prefs {
		if len(prefs.Emojis) == 0 {
			continue
		}

		if prefs.Emojis == nil {
			continue
		}

		toPersist[userID] = prefs
	}

	dir := filepath.Dir(s.filePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	tempFile, err := os.CreateTemp(dir, "prefs-*.json")
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(tempFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(toPersist); err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return err
	}

	if err := tempFile.Close(); err != nil {
		os.Remove(tempFile.Name())
		return err
	}

	if err := os.Rename(tempFile.Name(), s.filePath); err != nil {
		os.Remove(tempFile.Name())
		return err
	}

	return nil
}
