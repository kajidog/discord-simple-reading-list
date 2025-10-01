package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/example/discord-bookmark-manager/internal/reminders"
	"github.com/example/discord-bookmark-manager/internal/store"
)

// ListBookmarksCommandName identifies the slash command that shows saved bookmark preferences.
const ListBookmarksCommandName = "list-bookmarks"

// ListBookmarksCommand handles the `/list-bookmarks` slash command lifecycle.
type ListBookmarksCommand struct {
	store *store.EmojiStore
}

// NewListBookmarksCommand constructs a new ListBookmarksCommand.
func NewListBookmarksCommand(store *store.EmojiStore) *ListBookmarksCommand {
	return &ListBookmarksCommand{store: store}
}

// Definition returns the discordgo.ApplicationCommand definition for registration.
func (c *ListBookmarksCommand) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        ListBookmarksCommandName,
		Description: "Show the emojis and modes currently configured for your bookmarks",
	}
}

// Handle executes the command when invoked by a user.
func (c *ListBookmarksCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if i.Type != discordgo.InteractionApplicationCommand {
		return nil
	}

	user := i.Member.User
	if user == nil {
		user = i.User
	}
	if user == nil {
		return fmt.Errorf("unable to resolve user from interaction")
	}

	prefs, ok := c.store.Get(user.ID)
	if !ok || len(prefs.Emojis) == 0 {
		return respondEphemeral(s, i, "📭 No bookmark emojis saved yet. Use `/set-bookmark` to create one!")
	}

	emojis := make([]string, 0, len(prefs.Emojis))
	for emoji := range prefs.Emojis {
		emojis = append(emojis, emoji)
	}
	sort.Strings(emojis)

	var builder strings.Builder
	builder.WriteString("⭐ Saved bookmark shortcuts:\n")
	for _, emoji := range emojis {
		pref := prefs.Emojis[emoji]
		display := formatEmojiForDisplay(emoji)
		mode := string(pref.Mode)
		colorDescription := "default"
		if pref.HasColor {
			colorDescription = fmt.Sprintf("#%06X", pref.Color)
		}
		builder.WriteString(fmt.Sprintf("• %s — %s mode (color: %s)\n", display, mode, colorDescription))
		destinationLine := "  ↳ 📬 Destination: DMs"
		if pref.Destination == store.DestinationChannel && pref.ChannelID != "" {
			destinationLine = fmt.Sprintf("  ↳ 📬 Destination: <#%s>", pref.ChannelID)
		}
		builder.WriteString(destinationLine + "\n")
		reminderLine := fmt.Sprintf("  ↳ ⏰ Reminder: %s", reminders.Describe(pref.Reminder))
		if pref.Reminder != nil {
			if pref.Reminder.RemoveOnComplete {
				reminderLine += " / ✅ clears on Done"
			} else {
				reminderLine += " / 🔁 stays after Done"
			}
		}
		builder.WriteString(reminderLine + "\n")
	}

	builder.WriteString("\nUse `/set-bookmark` to tweak settings or `/remove-bookmark` to delete one.")

	return respondEphemeral(s, i, builder.String())
}

func formatEmojiForDisplay(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return value
	}

	parts := strings.Split(trimmed, ":")
	switch len(parts) {
	case 2:
		return fmt.Sprintf("<:%s:%s>", parts[0], parts[1])
	case 3:
		if parts[0] == "a" {
			return fmt.Sprintf("<a:%s:%s>", parts[1], parts[2])
		}
	}

	return trimmed
}
