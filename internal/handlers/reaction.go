package handlers

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/example/discord-simple-reading-list/internal/store"
)

const defaultEmbedColor = 0x5865F2

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

	prefs, ok := h.store.Get(event.UserID)
	if !ok || len(prefs.Emojis) == 0 {
		return
	}

	reactionID := event.Emoji.APIName()
	if reactionID == "" {
		reactionID = event.Emoji.Name
	}

	if !emojiMatches(prefs.Emojis, reactionID) {
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

	color := prefs.Color
	if !prefs.HasColor {
		color = defaultEmbedColor
	}

	jumpURL := buildJumpLink(event.GuildID, event.ChannelID, event.MessageID)
	embed := buildForwardedMessageEmbed(msg, channelName, jumpURL, color)
	_, err = s.ChannelMessageSendComplex(dmChannel.ID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{embed},
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

func buildForwardedMessageEmbed(msg *discordgo.Message, channelName, jumpURL string, color int) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title: "Saved message",
		Color: color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Author",
				Value:  msg.Author.String(),
				Inline: true,
			},
			{
				Name:   "Channel",
				Value:  fmt.Sprintf("#%s", channelName),
				Inline: true,
			},
		},
	}

	if msg.Content != "" {
		embed.Description = msg.Content
	}

	if jumpURL != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "Jump to message",
			Value: fmt.Sprintf("[Open original message](%s)", jumpURL),
		})
	}

	if len(msg.Attachments) > 0 {
		var attachments []string
		for _, attachment := range msg.Attachments {
			name := attachment.Filename
			if name == "" {
				name = attachment.URL
			}
			attachments = append(attachments, fmt.Sprintf("[%s](%s)", name, attachment.URL))
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "Attachments",
			Value: strings.Join(attachments, "\n"),
		})
	}

	return embed
}

func emojiMatches(emojis []string, value string) bool {
	for _, emoji := range emojis {
		if emoji == value {
			return true
		}
	}

	return false
}

func buildJumpLink(guildID, channelID, messageID string) string {
	if guildID == "" {
		return fmt.Sprintf("https://discord.com/channels/@me/%s/%s", channelID, messageID)
	}

	return fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guildID, channelID, messageID)
}
