package handlers

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/example/discord-simple-reading-list/internal/store"
)

// ReactionHandler sends a direct message when a user reacts with their registered emoji.
type ReactionHandler struct {
	store *store.EmojiStore
}

// NewReactionHandler constructs a ReactionHandler.
func NewReactionHandler(store *store.EmojiStore) *ReactionHandler {
	return &ReactionHandler{store: store}
}

// Handle reacts to MessageReactionAdd events.
func (h *ReactionHandler) Handle(s *discordgo.Session, event *discordgo.MessageReactionAdd) {
	if event.UserID == "" {
		return
	}

	if botUser := s.State.User; botUser != nil && event.UserID == botUser.ID {
		return
	}

	emoji, ok := h.store.Get(event.UserID)
	if !ok {
		return
	}

	reactionID := event.Emoji.APIName()
	if reactionID == "" {
		reactionID = event.Emoji.Name
	}

	if reactionID != emoji {
		return
	}

	msg, err := s.ChannelMessage(event.ChannelID, event.MessageID)
	if err != nil {
		log.Printf("failed to fetch message: %v", err)
		return
	}

	channelName := fetchChannelName(s, event.ChannelID)

	dmChannel, err := s.UserChannelCreate(event.UserID)
	if err != nil {
		log.Printf("failed to create DM channel: %v", err)
		return
	}

	content := buildForwardedMessageContent(msg, channelName)
	_, err = s.ChannelMessageSendComplex(dmChannel.ID, &discordgo.MessageSend{
		Content: content,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Close",
						Style:    discordgo.DangerButton,
						CustomID: CloseButtonID,
					},
				},
			},
		},
	})
	if err != nil {
		log.Printf("failed to send DM: %v", err)
	}
}

func fetchChannelName(s *discordgo.Session, channelID string) string {
	if channel, err := s.State.Channel(channelID); err == nil && channel != nil {
		return channel.Name
	}

	channel, err := s.Channel(channelID)
	if err != nil {
		log.Printf("failed to fetch channel name: %v", err)
		return channelID
	}

	if channel.Name != "" {
		return channel.Name
	}

	return channelID
}

func buildForwardedMessageContent(msg *discordgo.Message, channelName string) string {
	var builder strings.Builder
	builder.WriteString("**Saved message**\n")
	builder.WriteString(fmt.Sprintf("Author: %s\n", msg.Author.Username))
	builder.WriteString(fmt.Sprintf("Channel: #%s\n\n", channelName))

	if msg.Content != "" {
		builder.WriteString(msg.Content)
		builder.WriteString("\n")
	}

	if len(msg.Attachments) > 0 {
		builder.WriteString("\nAttachments:\n")
		for _, attachment := range msg.Attachments {
			builder.WriteString(attachment.URL)
			builder.WriteString("\n")
		}
	}

	return strings.TrimSpace(builder.String())
}
