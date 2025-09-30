package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/example/discord-simple-reading-list/internal/store"
)

// SetBookmarkEmojiCommandName identifies the slash command for selecting the bookmark reaction emoji.
const SetBookmarkEmojiCommandName = "set-bookmark-emoji"

// SetBookmarkEmojiCommand handles the `/set-bookmark-emoji` slash command lifecycle.
type SetBookmarkEmojiCommand struct {
	store *store.EmojiStore
}

// NewSetBookmarkEmojiCommand constructs a new SetBookmarkEmojiCommand.
func NewSetBookmarkEmojiCommand(store *store.EmojiStore) *SetBookmarkEmojiCommand {
	return &SetBookmarkEmojiCommand{store: store}
}

// Definition returns the discordgo.ApplicationCommand definition for registration.
func (c *SetBookmarkEmojiCommand) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        SetBookmarkEmojiCommandName,
		Description: "Choose the emoji used to save messages to your DM",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "emoji",
				Description: "Emoji (or emojis separated by spaces or commas) to watch for when you react to a message",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "color",
				Description: "Optional hex color (e.g. #ffcc00) for the saved message embed",
				Required:    false,
			},
		},
	}
}

// Handle executes the command when invoked by a user.
func (c *SetBookmarkEmojiCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if i.Type != discordgo.InteractionApplicationCommand {
		return nil
	}

	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		return fmt.Errorf("emoji option is required")
	}

	var rawEmoji string
	var rawColor string

	for _, option := range options {
		switch option.Name {
		case "emoji":
			rawEmoji = strings.TrimSpace(option.StringValue())
		case "color":
			rawColor = strings.TrimSpace(option.StringValue())
		}
	}

	if rawEmoji == "" {
		return fmt.Errorf("emoji option is required")
	}

	emojiTokens := splitEmojiInput(rawEmoji)
	if len(emojiTokens) == 0 {
		return fmt.Errorf("please provide at least one emoji")
	}

	normalized := normalizeEmojis(emojiTokens)
	if len(normalized) == 0 {
		return fmt.Errorf("unable to understand the provided emojis")
	}

	color, hasColor, err := parseColor(rawColor)
	if err != nil {
		return err
	}

	user := i.Member.User
	if user == nil {
		user = i.User
	}
	if user == nil {
		return fmt.Errorf("unable to resolve user from interaction")
	}

	c.store.Set(user.ID, store.UserPreferences{Emojis: normalized, Color: color, HasColor: hasColor})

	displayEmoji := strings.Join(emojiTokens, ", ")
	response := fmt.Sprintf("Saved %s as your bookmark emoji(s). React with them to save messages to your DM!", displayEmoji)
	if hasColor {
		response += fmt.Sprintf(" Embed color set to #%s.", strings.ToUpper(fmt.Sprintf("%06x", color)))
	}
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func normalizeEmoji(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	if strings.HasPrefix(trimmed, "<") && strings.HasSuffix(trimmed, ">") {
		inner := strings.Trim(trimmed[1:len(trimmed)-1], ":")
		parts := strings.Split(inner, ":")

		switch len(parts) {
		case 2:
			return strings.Join(parts, ":")
		case 3:
			if parts[0] == "a" {
				return strings.Join(parts, ":")
			}
			return strings.Join(parts[1:], ":")
		}
	}

	return trimmed
}

func splitEmojiInput(raw string) []string {
	replacer := strings.NewReplacer(",", " ", "\n", " ")
	cleaned := replacer.Replace(raw)
	fields := strings.Fields(cleaned)

	var result []string
	for _, field := range fields {
		trimmed := strings.TrimSpace(field)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

func normalizeEmojis(values []string) []string {
	seen := make(map[string]struct{})
	var normalized []string

	for _, value := range values {
		emoji := normalizeEmoji(value)
		if emoji == "" {
			continue
		}

		if _, ok := seen[emoji]; ok {
			continue
		}

		seen[emoji] = struct{}{}
		normalized = append(normalized, emoji)
	}

	return normalized
}

func parseColor(value string) (int, bool, error) {
	if value == "" {
		return 0, false, nil
	}

	cleaned := strings.ToLower(strings.TrimSpace(value))
	cleaned = strings.TrimPrefix(cleaned, "0x")
	cleaned = strings.TrimPrefix(cleaned, "#")

	if len(cleaned) != 6 {
		return 0, false, fmt.Errorf("color must be a 6 digit hex code")
	}

	parsed, err := strconv.ParseInt(cleaned, 16, 32)
	if err != nil {
		return 0, false, fmt.Errorf("invalid color value: %w", err)
	}

	return int(parsed), true, nil
}
