package reminders

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Payload contains contextual information used when sending the reminder message.
type Payload struct {
	ChannelID      string
	JumpURL        string
	BookmarkURL    string
	ChannelName    string
	ContentSnippet string
}

type scheduledReminder struct {
	timer            *time.Timer
	when             time.Time
	removeOnComplete bool
	payload          Payload
}

// Service keeps track of scheduled reminders and delivers them at the appropriate time.
type Service struct {
	session   *discordgo.Session
	mu        sync.Mutex
	scheduled map[string]*scheduledReminder
	filePath  string
}

type persistedReminder struct {
	When             string  `json:"when"`
	RemoveOnComplete bool    `json:"removeOnComplete"`
	Payload          Payload `json:"payload"`
}

// NewService constructs a reminder service bound to the provided Discord session.
func NewService(session *discordgo.Session, filePath string) (*Service, error) {
	service := &Service{
		session:   session,
		scheduled: make(map[string]*scheduledReminder),
		filePath:  filePath,
	}

	if err := service.restore(); err != nil {
		return nil, err
	}

	return service, nil
}

// Schedule registers a reminder for the given bookmark message ID.
func (s *Service) Schedule(messageID string, when time.Time, payload Payload, removeOnComplete bool) {
	if when.IsZero() {
		return
	}

	s.mu.Lock()
	s.scheduleLocked(messageID, when, payload, removeOnComplete)
	if err := s.persistLocked(); err != nil {
		log.Printf("failed to persist reminders: %v", err)
	}
	s.mu.Unlock()
}

// Cancel removes any pending reminder for the provided bookmark message ID.
func (s *Service) Cancel(messageID string) {
	s.mu.Lock()
	reminder, ok := s.scheduled[messageID]
	if ok {
		if reminder.timer != nil {
			reminder.timer.Stop()
		}
		delete(s.scheduled, messageID)
		if err := s.persistLocked(); err != nil {
			log.Printf("failed to persist reminders: %v", err)
		}
	}
	s.mu.Unlock()
}

// Complete handles the completion action. Depending on the configuration the reminder is optionally cancelled.
func (s *Service) Complete(messageID string) {
	s.mu.Lock()
	reminder, ok := s.scheduled[messageID]
	s.mu.Unlock()
	if !ok {
		return
	}

	if reminder.removeOnComplete {
		s.Cancel(messageID)
	}
}

// Close stops all scheduled reminders. It should be called during shutdown to prevent dangling timers.
func (s *Service) Close() {
	s.mu.Lock()
	for _, reminder := range s.scheduled {
		if reminder.timer != nil {
			reminder.timer.Stop()
		}
		reminder.timer = nil
	}
	s.mu.Unlock()
}

func (s *Service) deliver(messageID string) {
	s.mu.Lock()
	reminder, ok := s.scheduled[messageID]
	if ok {
		delete(s.scheduled, messageID)
		if err := s.persistLocked(); err != nil {
			log.Printf("failed to persist reminders: %v", err)
		}
	}
	s.mu.Unlock()

	if !ok {
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "⏰ Reminder",
		Description: fmt.Sprintf("Take another look at #%s.", reminder.payload.ChannelName),
		Color:       0xFEE75C,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	if reminder.payload.ContentSnippet != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "📝 Note",
			Value: reminder.payload.ContentSnippet,
		})
	}

	if reminder.payload.JumpURL != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "🔗 Source Message",
			Value: fmt.Sprintf("[Open message](%s)", reminder.payload.JumpURL),
		})
	}

	if reminder.payload.BookmarkURL != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "📬 Saved Bookmark",
			Value: fmt.Sprintf("[Open DM](%s)", reminder.payload.BookmarkURL),
		})
	}

	_, err := s.session.ChannelMessageSendComplex(reminder.payload.ChannelID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{embed},
	})
	if err != nil {
		log.Printf("failed to deliver reminder: %v", err)
	}
}

func (s *Service) scheduleLocked(messageID string, when time.Time, payload Payload, removeOnComplete bool) {
	delay := time.Until(when)
	if delay <= 0 {
		delay = time.Second
	}

	if existing, ok := s.scheduled[messageID]; ok {
		if existing.timer != nil {
			existing.timer.Stop()
		}
	}

	reminder := &scheduledReminder{
		removeOnComplete: removeOnComplete,
		payload:          payload,
	}
	reminder.when = time.Now().Add(delay)
	reminder.timer = time.AfterFunc(delay, func() {
		s.deliver(messageID)
	})

	s.scheduled[messageID] = reminder
}

func (s *Service) restore() error {
	if s.filePath == "" {
		return nil
	}

	file, err := os.Open(s.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var persisted map[string]persistedReminder
	if err := decoder.Decode(&persisted); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	for messageID, stored := range persisted {
		when, err := time.Parse(time.RFC3339Nano, stored.When)
		if err != nil {
			log.Printf("failed to parse reminder time for %s: %v", messageID, err)
			continue
		}
		if !when.After(now) {
			when = now.Add(time.Second)
		}
		s.scheduleLocked(messageID, when, stored.Payload, stored.RemoveOnComplete)
	}

	return nil
}

func (s *Service) persistLocked() error {
	if s.filePath == "" {
		return nil
	}

	toPersist := make(map[string]persistedReminder, len(s.scheduled))
	for id, reminder := range s.scheduled {
		if reminder == nil {
			continue
		}

		when := reminder.when
		if when.IsZero() {
			when = time.Now().Add(time.Second)
		}

		toPersist[id] = persistedReminder{
			When:             when.Format(time.RFC3339Nano),
			RemoveOnComplete: reminder.removeOnComplete,
			Payload:          reminder.payload,
		}
	}

	dir := filepath.Dir(s.filePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	tempFile, err := os.CreateTemp(dir, "reminders-*.json")
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
