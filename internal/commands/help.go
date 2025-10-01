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

	helpText := "ブックマークボットの使い方:\n" +
		"• `/set-bookmark emoji:😊 mode:lightweight color:#FFD700` — 絵文字に保存モードと色を割り当てます。\n" +
		"• `/list-bookmarks` — 現在登録している絵文字とモードを確認できます。\n" +
		"メッセージに指定した絵文字でリアクションすると、設定したモードでDMに送信されます。"

	return respondEphemeral(s, i, helpText)
}
