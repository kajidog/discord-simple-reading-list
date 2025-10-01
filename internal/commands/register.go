package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/example/discord-bookmark-manager/internal/reminders"
	"github.com/example/discord-bookmark-manager/internal/store"
)

// SetBookmarkCommandName identifies the slash command for selecting the bookmark reaction emoji and mode.
const SetBookmarkCommandName = "set-bookmark"

// SetBookmarkCommand handles the `/set-bookmark` slash command lifecycle.
type SetBookmarkCommand struct {
	store *store.EmojiStore
}

// NewSetBookmarkCommand constructs a new SetBookmarkCommand.
func NewSetBookmarkCommand(store *store.EmojiStore) *SetBookmarkCommand {
	return &SetBookmarkCommand{store: store}
}

// Definition returns the discordgo.ApplicationCommand definition for registration.
func (c *SetBookmarkCommand) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        SetBookmarkCommandName,
		Description: "Choose how each emoji saves messages to your DM",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "emoji",
				Description: "Emoji to watch for when you react to a message",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "mode",
				Description: "Save mode: lightweight, balanced, or complete",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "👀 Lightweight", Value: string(store.ModeLightweight)},
					{Name: "🔖 Balanced", Value: string(store.ModeBalanced)},
					{Name: "📌 Complete", Value: string(store.ModeComplete)},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "color",
				Description: "Optional hex color (e.g. #ffcc00) for the saved message embed",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reminder",
				Description: "Optional reminder such as 08:00 or 45m",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "keep-reminder-on-complete",
				Description: "Keep reminder when pressing the complete button",
				Required:    false,
			},
		},
	}
}

// Handle executes the command when invoked by a user.
func (c *SetBookmarkCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if i.Type != discordgo.InteractionApplicationCommand {
		return nil
	}

	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		return fmt.Errorf("emoji option is required")
	}

	var rawEmoji string
	var rawColor string
	var rawMode string
	var rawReminder string
	var reminderProvided bool
	var keepReminder bool
	var keepProvided bool

	for _, option := range options {
		switch option.Name {
		case "emoji":
			rawEmoji = strings.TrimSpace(option.StringValue())
		case "mode":
			rawMode = strings.TrimSpace(option.StringValue())
		case "color":
			rawColor = strings.TrimSpace(option.StringValue())
		case "reminder":
			rawReminder = strings.TrimSpace(option.StringValue())
			reminderProvided = true
		case "keep-reminder-on-complete":
			keepReminder = option.BoolValue()
			keepProvided = true
		}
	}

	if rawEmoji == "" {
		return fmt.Errorf("emoji option is required")
	}

	if rawMode == "" {
		return fmt.Errorf("mode option is required")
	}

	emojiTokens := splitEmojiInput(rawEmoji)
	if len(emojiTokens) == 0 {
		return fmt.Errorf("please provide an emoji")
	}

	if len(emojiTokens) != 1 {
		return fmt.Errorf("please configure one emoji at a time")
	}

	normalized := normalizeEmoji(emojiTokens[0])
	if normalized == "" {
		return fmt.Errorf("unable to understand the provided emoji")
	}

	color, hasColor, err := parseColor(rawColor)
	if err != nil {
		return err
	}

	mode := store.BookmarkMode(strings.ToLower(rawMode))
	switch mode {
	case store.ModeLightweight, store.ModeBalanced, store.ModeComplete:
	default:
		return fmt.Errorf("invalid mode. choose lightweight, balanced, or complete")
	}

	user := i.Member.User
	if user == nil {
		user = i.User
	}
	if user == nil {
		return fmt.Errorf("unable to resolve user from interaction")
	}

	existingPref, hasExisting := c.store.GetEmoji(user.ID, normalized)

	var reminderPref *reminders.Preference
	if hasExisting && existingPref.Reminder != nil {
		copied := *existingPref.Reminder
		reminderPref = &copied
	}

	if reminderProvided {
		parsedReminder, err := reminders.Parse(rawReminder)
		if err != nil {
			return err
		}

		if parsedReminder == nil {
			if keepProvided {
				return fmt.Errorf("リマインドを削除する場合は keep-reminder-on-complete オプションは使用できません")
			}
			reminderPref = nil
		} else {
			removeOnComplete := true
			if keepProvided {
				removeOnComplete = !keepReminder
			} else if reminderPref != nil {
				removeOnComplete = reminderPref.RemoveOnComplete
			}
			parsedReminder.RemoveOnComplete = removeOnComplete
			reminderPref = parsedReminder
		}
	} else if keepProvided {
		if reminderPref == nil {
			return fmt.Errorf("リマインドが設定されていないため keep-reminder-on-complete を変更できません。reminder オプションを設定してください。")
		}
		reminderPref.RemoveOnComplete = !keepReminder
	}

	prefToSave := store.EmojiPreference{Mode: mode, Color: color, HasColor: hasColor}
	if reminderPref != nil {
		copied := *reminderPref
		prefToSave.Reminder = &copied
	}

	if err := c.store.SetEmoji(user.ID, normalized, prefToSave); err != nil {
		return fmt.Errorf("failed to save emoji preference: %w", err)
	}

	response := fmt.Sprintf("Saved %s in %s mode. React with it to save messages to your DM!", emojiTokens[0], string(mode))
	if hasColor {
		response += fmt.Sprintf(" Embed color set to #%s.", strings.ToUpper(fmt.Sprintf("%06x", color)))
	}
	if reminderProvided {
		if reminderPref == nil {
			response += " Reminder cleared."
		} else {
			response += fmt.Sprintf(" Reminder: %s.", reminders.Describe(reminderPref))
		}
	} else if reminderPref != nil {
		response += fmt.Sprintf(" Reminder: %s.", reminders.Describe(reminderPref))
	}
	if reminderPref != nil {
		if reminderPref.RemoveOnComplete {
			response += " 完了ボタンでリマインドも削除されます。"
		} else {
			response += " 完了ボタンを押してもリマインドは残ります。"
		}
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
