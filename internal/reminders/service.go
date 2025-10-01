package reminders

import (
	"fmt"
	"log"
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
	removeOnComplete bool
	payload          Payload
}

// Service keeps track of scheduled reminders and delivers them at the appropriate time.
type Service struct {
	session   *discordgo.Session
	mu        sync.Mutex
	scheduled map[string]*scheduledReminder
}

// NewService constructs a reminder service bound to the provided Discord session.
func NewService(session *discordgo.Session) *Service {
	return &Service{
		session:   session,
		scheduled: make(map[string]*scheduledReminder),
	}
}

// Schedule registers a reminder for the given bookmark message ID.
func (s *Service) Schedule(messageID string, when time.Time, payload Payload, removeOnComplete bool) {
	if when.IsZero() {
		return
	}

	delay := time.Until(when)
	if delay <= 0 {
		delay = time.Second
	}

	s.mu.Lock()
	if existing, ok := s.scheduled[messageID]; ok {
		existing.timer.Stop()
	}

	reminder := &scheduledReminder{
		removeOnComplete: removeOnComplete,
		payload:          payload,
	}
	reminder.timer = time.AfterFunc(delay, func() {
		s.deliver(messageID)
	})

	s.scheduled[messageID] = reminder
	s.mu.Unlock()
}

// Cancel removes any pending reminder for the provided bookmark message ID.
func (s *Service) Cancel(messageID string) {
	s.mu.Lock()
	reminder, ok := s.scheduled[messageID]
	if ok {
		reminder.timer.Stop()
		delete(s.scheduled, messageID)
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
	for id, reminder := range s.scheduled {
		reminder.timer.Stop()
		delete(s.scheduled, id)
	}
	s.mu.Unlock()
}

func (s *Service) deliver(messageID string) {
	s.mu.Lock()
	reminder, ok := s.scheduled[messageID]
	if ok {
		delete(s.scheduled, messageID)
	}
	s.mu.Unlock()

	if !ok {
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "⏰ リマインド",
		Description: fmt.Sprintf("#%s のメッセージを確認しましょう。", reminder.payload.ChannelName),
		Color:       0xFEE75C,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	if reminder.payload.ContentSnippet != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "メモ",
			Value: reminder.payload.ContentSnippet,
		})
	}

	if reminder.payload.JumpURL != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "元メッセージ",
			Value: fmt.Sprintf("[リンクはこちら](%s)", reminder.payload.JumpURL),
		})
	}

	if reminder.payload.BookmarkURL != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "ブックマーク",
			Value: fmt.Sprintf("[DMを開く](%s)", reminder.payload.BookmarkURL),
		})
	}

	_, err := s.session.ChannelMessageSendComplex(reminder.payload.ChannelID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{embed},
	})
	if err != nil {
		log.Printf("failed to deliver reminder: %v", err)
	}
}
