package handlers

import (
	"fmt"
	"log"
	"strings"
	"time"

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

	pref, ok := prefs.Emojis[reactionID]
	if !ok {
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

	color := pref.Color
	if !pref.HasColor {
		color = defaultEmbedColor
	}

	jumpURL := buildJumpLink(event.GuildID, event.ChannelID, event.MessageID)
	var messageSend *discordgo.MessageSend

	switch pref.Mode {
	case store.ModeLightweight:
		messageSend = buildLightweightBookmark(msg, channelName, jumpURL, color, &event.Emoji)
	case store.ModeComplete:
		messageSend = buildCompleteBookmark(msg, channelName, jumpURL, color)
	case store.ModeBalanced:
		messageSend = buildBalancedBookmark(msg, channelName, jumpURL, color)
	default:
		messageSend = buildBalancedBookmark(msg, channelName, jumpURL, color)
	}

	if messageSend == nil {
		return
	}

	_, err = s.ChannelMessageSendComplex(dmChannel.ID, messageSend)
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

func buildJumpLink(guildID, channelID, messageID string) string {
	if guildID == "" {
		return fmt.Sprintf("https://discord.com/channels/@me/%s/%s", channelID, messageID)
	}

	return fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guildID, channelID, messageID)
}

func buildLightweightBookmark(msg *discordgo.Message, channelName, jumpURL string, color int, emoji *discordgo.Emoji) *discordgo.MessageSend {
	titleEmoji := "👀"
	if emoji != nil && emoji.Name != "" {
		titleEmoji = emoji.Name
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s 後で読む", titleEmoji),
		Color: color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "チャンネル",
				Value:  fmt.Sprintf("#%s", channelName),
				Inline: true,
			},
			{
				Name:   "保存日時",
				Value:  time.Now().Format("2006-01-02 15:04"),
				Inline: true,
			},
		},
	}

	if msg.Content != "" {
		truncated := []rune(msg.Content)
		if len(truncated) > 500 {
			truncated = truncated[:500]
		}
		embed.Description = strings.TrimSpace(string(truncated))
	}

	if jumpURL != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "元メッセージ",
			Value: fmt.Sprintf("[リンクはこちら](%s)", jumpURL),
		})
	}

	if imageURL := firstImageURL(msg); imageURL != "" {
		embed.Image = &discordgo.MessageEmbedImage{URL: imageURL}
	}

	return &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{embed},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "完了",
					Style:    discordgo.SuccessButton,
					CustomID: CompleteButtonID,
					Emoji:    discordgo.ComponentEmoji{Name: "✅"},
				},
				discordgo.Button{
					Label:    "削除",
					Style:    discordgo.DangerButton,
					CustomID: DeleteButtonID,
					Emoji:    discordgo.ComponentEmoji{Name: "🗑️"},
				},
			}},
		},
	}
}

func buildCompleteBookmark(msg *discordgo.Message, channelName, jumpURL string, color int) *discordgo.MessageSend {
	infoEmbed := buildInfoEmbed("📌 完全保存", msg, channelName, jumpURL, color, true)

	embeds := []*discordgo.MessageEmbed{infoEmbed}
	for _, e := range msg.Embeds {
		if e == nil {
			continue
		}
		embeds = append(embeds, cloneEmbed(e))
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{Components: []discordgo.MessageComponent{
			discordgo.Button{
				Style: discordgo.LinkButton,
				Label: "元メッセージ",
				URL:   jumpURL,
				Emoji: discordgo.ComponentEmoji{Name: "🔗"},
			},
			discordgo.Button{
				Label:    "削除",
				Style:    discordgo.DangerButton,
				CustomID: DeleteButtonID,
				Emoji:    discordgo.ComponentEmoji{Name: "🗑️"},
			},
		}},
	}

	return &discordgo.MessageSend{
		Embeds:     embeds,
		Components: components,
	}
}

func buildBalancedBookmark(msg *discordgo.Message, channelName, jumpURL string, color int) *discordgo.MessageSend {
	infoEmbed := buildInfoEmbed("🔖 ブックマーク", msg, channelName, jumpURL, color, false)

	embeds := []*discordgo.MessageEmbed{infoEmbed}
	if len(msg.Embeds) == 1 && msg.Embeds[0] != nil {
		embeds = append(embeds, cloneEmbed(msg.Embeds[0]))
	}

	return &discordgo.MessageSend{
		Embeds: embeds,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "削除",
					Style:    discordgo.DangerButton,
					CustomID: DeleteButtonID,
					Emoji:    discordgo.ComponentEmoji{Name: "🗑️"},
				},
			}},
		},
	}
}

func buildInfoEmbed(title string, msg *discordgo.Message, channelName, jumpURL string, color int, includeAllAttachments bool) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title: title,
		Color: color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "投稿者",
				Value:  msg.Author.String(),
				Inline: true,
			},
			{
				Name:   "チャンネル",
				Value:  fmt.Sprintf("#%s", channelName),
				Inline: true,
			},
			{
				Name:   "投稿日",
				Value:  msg.Timestamp.Format("2006-01-02 15:04"),
				Inline: true,
			},
		},
	}

	if msg.Content != "" {
		embed.Description = msg.Content
	}

	if jumpURL != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "元メッセージ",
			Value: fmt.Sprintf("[Jump](%s)", jumpURL),
		})
	}

	attachments := buildAttachmentField(msg.Attachments, includeAllAttachments)
	if attachments != nil {
		embed.Fields = append(embed.Fields, attachments)
	}

	return embed
}

func buildAttachmentField(attachments []*discordgo.MessageAttachment, includeAll bool) *discordgo.MessageEmbedField {
	if len(attachments) == 0 {
		return nil
	}

	limit := len(attachments)
	if !includeAll && limit > 3 {
		limit = 3
	}

	var entries []string
	for idx, attachment := range attachments {
		if idx >= limit {
			break
		}

		name := attachment.Filename
		if name == "" {
			name = attachment.URL
		}
		entries = append(entries, fmt.Sprintf("[%s](%s)", name, attachment.URL))
	}

	if !includeAll && len(attachments) > limit {
		remaining := len(attachments) - limit
		entries = append(entries, fmt.Sprintf("ほか%d件", remaining))
	}

	return &discordgo.MessageEmbedField{
		Name:  "添付ファイル",
		Value: strings.Join(entries, "\n"),
	}
}

func firstImageURL(msg *discordgo.Message) string {
	for _, attachment := range msg.Attachments {
		if isImageAttachment(attachment) {
			return attachment.URL
		}
	}

	for _, embed := range msg.Embeds {
		if embed == nil {
			continue
		}
		if embed.Image != nil && embed.Image.URL != "" {
			return embed.Image.URL
		}
		if embed.Thumbnail != nil && embed.Thumbnail.URL != "" {
			return embed.Thumbnail.URL
		}
	}

	return ""
}

func isImageAttachment(attachment *discordgo.MessageAttachment) bool {
	if attachment == nil {
		return false
	}

	if strings.HasPrefix(attachment.ContentType, "image/") {
		return true
	}

	lower := strings.ToLower(attachment.Filename)
	imageExts := []string{".png", ".jpg", ".jpeg", ".gif", ".webp"}
	for _, ext := range imageExts {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}

	return false
}

func cloneEmbed(embed *discordgo.MessageEmbed) *discordgo.MessageEmbed {
	if embed == nil {
		return nil
	}

	cloned := *embed

	if embed.Fields != nil {
		cloned.Fields = make([]*discordgo.MessageEmbedField, len(embed.Fields))
		for i, field := range embed.Fields {
			if field == nil {
				continue
			}
			copied := *field
			cloned.Fields[i] = &copied
		}
	}

	if embed.Author != nil {
		copied := *embed.Author
		cloned.Author = &copied
	}

	if embed.Footer != nil {
		copied := *embed.Footer
		cloned.Footer = &copied
	}

	if embed.Image != nil {
		copied := *embed.Image
		cloned.Image = &copied
	}

	if embed.Thumbnail != nil {
		copied := *embed.Thumbnail
		cloned.Thumbnail = &copied
	}

	if embed.Provider != nil {
		copied := *embed.Provider
		cloned.Provider = &copied
	}

	if embed.Video != nil {
		copied := *embed.Video
		cloned.Video = &copied
	}

	if embed.Color != 0 {
		cloned.Color = embed.Color
	}

	if embed.Timestamp != "" {
		cloned.Timestamp = embed.Timestamp
	}

	if embed.Fields != nil {
		// Already handled above but ensure nil entries remain nil.
		for i, field := range embed.Fields {
			if field == nil {
				cloned.Fields[i] = nil
			}
		}
	}

	return &cloned
}
