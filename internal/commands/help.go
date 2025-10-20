package commands

import "github.com/bwmarrin/discordgo"

// HelpCommandName identifies the slash command that returns usage instructions.
const HelpCommandName = "bookmark-help"

// HelpCommand handles the `/bookmark-help` slash command lifecycle.
type HelpCommand struct{}

// NewHelpCommand constructs a new HelpCommand.
func NewHelpCommand() *HelpCommand {
	return &HelpCommand{}
}

// Definition returns the discordgo.ApplicationCommand definition for registration.
func (c *HelpCommand) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        HelpCommandName,
		Description: "Show how to configure and use the bookmark bot",
	}
}

// Handle executes the command when invoked by a user.
func (c *HelpCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if i.Type != discordgo.InteractionApplicationCommand {
		return nil
	}

	helpText := "🛠️ Bookmark bot quick guide:\n\n" +
		"**Basic usage:**\n" +
		"• `/set-bookmark` — Set up an emoji with a bookmark mode\n" +
		"  - Choose emoji, mode (Lightweight/Balanced/Complete), and optional color\n" +
		"  - Example: Select mode \"👀 Lightweight\" and enter color `#FFD700`\n\n" +
		"**With reminders:**\n" +
		"• Add `reminder` option with time like `8:00` or duration like `30m`\n" +
		"• Use `keep-reminder-on-complete` if you want reminders to persist after marking Done\n\n" +
		"**Send to channel:**\n" +
		"• Set `destination` to \"# Channel\" and select a `destination-channel`\n\n" +
		"**Other commands:**\n" +
		"• `/list-bookmarks` — View all your configured emojis\n" +
		"• `/remove-bookmark` — Delete an emoji configuration\n\n" +
		"React with a saved emoji to bookmark messages. Reminders always arrive in your DMs."

	return respondEphemeral(s, i, helpText)
}
