package commands

import (
	"fmt"
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
				Description: "Emoji to watch for when you react to a message",
				Required:    true,
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

	rawEmoji := strings.TrimSpace(options[0].StringValue())
	if rawEmoji == "" {
		return fmt.Errorf("emoji option is required")
	}

	normalized := normalizeEmoji(rawEmoji)

	user := i.Member.User
	if user == nil {
		user = i.User
	}
	if user == nil {
		return fmt.Errorf("unable to resolve user from interaction")
	}

	c.store.Set(user.ID, normalized)

	response := fmt.Sprintf("Set %s as your bookmark emoji. React with it to save messages to your DM!", rawEmoji)
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
