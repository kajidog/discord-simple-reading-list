package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/example/discord-bookmark-manager/internal/store"
)

// RemoveBookmarkCommandName identifies the slash command used to delete saved bookmark emojis.
const RemoveBookmarkCommandName = "remove-bookmark"

// RemoveBookmarkCommand handles the `/remove-bookmark` slash command lifecycle.
type RemoveBookmarkCommand struct {
	store *store.EmojiStore
}

// NewRemoveBookmarkCommand constructs a new RemoveBookmarkCommand.
func NewRemoveBookmarkCommand(store *store.EmojiStore) *RemoveBookmarkCommand {
	return &RemoveBookmarkCommand{store: store}
}

// Definition returns the discordgo.ApplicationCommand definition for registration.
func (c *RemoveBookmarkCommand) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        RemoveBookmarkCommandName,
		Description: "Delete a saved emoji shortcut",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "emoji",
				Description: "Emoji to remove from your saved shortcuts",
				Required:    true,
			},
		},
	}
}

// Handle executes the command when invoked by a user.
func (c *RemoveBookmarkCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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

	emojiTokens := splitEmojiInput(rawEmoji)
	if len(emojiTokens) == 0 {
		return fmt.Errorf("please provide an emoji")
	}
	if len(emojiTokens) != 1 {
		return fmt.Errorf("please remove one emoji at a time")
	}

	normalized := normalizeEmoji(emojiTokens[0])
	if normalized == "" {
		return fmt.Errorf("please provide an emoji")
	}

	user := i.Member.User
	if user == nil {
		user = i.User
	}
	if user == nil {
		return fmt.Errorf("unable to resolve user from interaction")
	}

	removed, err := c.store.DeleteEmoji(user.ID, normalized)
	if err != nil {
		return fmt.Errorf("failed to remove emoji preference: %w", err)
	}

	var content string
	if !removed {
		content = "⚠️ That emoji isn't saved yet. Use `/set-bookmark` to add it first."
	} else {
		content = fmt.Sprintf("🧹 Removed %s from your shortcuts.", formatEmojiForDisplay(emojiTokens[0]))
	}

	return respondEphemeral(s, i, content)
}
