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

	helpText := "🛠️ Bookmark bot quick guide:\n" +
		"• `/set-bookmark emoji:😊 mode:lightweight color:#FFD700` — Assign a save mode and embed color to an emoji.\n" +
		"• `/set-bookmark emoji:⏰ mode:lightweight reminder:8:00` — Add a reminder time or duration like `30m`.\n" +
		"  Keep the reminder after tapping Done with `keep-reminder-on-complete:true`.\n" +
		"• `/list-bookmarks` — Review every emoji shortcut you've saved.\n" +
		"• `/remove-bookmark emoji:😊` — Delete an emoji shortcut you no longer need.\n" +
		"React with a saved emoji to receive a DM in the layout you chose."

	return respondEphemeral(s, i, helpText)
}
