package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/example/discord-simple-reading-list/internal/store"
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
		return respondEphemeral(s, i, "まだブックマークは登録されていません。`/set-bookmark` コマンドで保存を始めましょう！")
	}

	emojis := make([]string, 0, len(prefs.Emojis))
	for emoji := range prefs.Emojis {
		emojis = append(emojis, emoji)
	}
	sort.Strings(emojis)

	var builder strings.Builder
	builder.WriteString("現在登録されているブックマーク設定:\n")
	for _, emoji := range emojis {
		pref := prefs.Emojis[emoji]
		display := formatEmojiForDisplay(emoji)
		mode := string(pref.Mode)
		colorDescription := "デフォルト"
		if pref.HasColor {
			colorDescription = fmt.Sprintf("#%06X", pref.Color)
		}
		builder.WriteString(fmt.Sprintf("• %s — %s モード (色: %s)\n", display, mode, colorDescription))
	}

	builder.WriteString("\n設定を変更する場合は `/set-bookmark` を使用してください。")

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
