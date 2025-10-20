package bot

import (
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/example/discord-bookmark-manager/internal/commands"
	"github.com/example/discord-bookmark-manager/internal/config"
	"github.com/example/discord-bookmark-manager/internal/handlers"
	"github.com/example/discord-bookmark-manager/internal/reminders"
	"github.com/example/discord-bookmark-manager/internal/store"
)

// Bot encapsulates the Discord session and all registered handlers.
type Bot struct {
	session         *discordgo.Session
	config          *config.Config
	store           *store.EmojiStore
	registerCmd     *commands.SetBookmarkCommand
	removeCmd       *commands.RemoveBookmarkCommand
	listCmd         *commands.ListBookmarksCommand
	helpCmd         *commands.HelpCommand
	reactionHandle  *handlers.ReactionHandler
	componentHandle *handlers.ComponentHandler
	reminders       *reminders.Service
	commandIDs      []string
}

// New constructs a new Bot instance.
func New(cfg *config.Config) (*Bot, error) {
	session, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		return nil, err
	}

	emojiStore, err := store.NewEmojiStore(cfg.StorePath)
	if err != nil {
		return nil, err
	}

	reminderService, err := reminders.NewService(session, cfg.ReminderStorePath)
	if err != nil {
		return nil, err
	}

	registerCommand := commands.NewSetBookmarkCommand(emojiStore)
	removeCommand := commands.NewRemoveBookmarkCommand(emojiStore)
	listCommand := commands.NewListBookmarksCommand(emojiStore)
	helpCommand := commands.NewHelpCommand()
	reactionHandler := handlers.NewReactionHandler(emojiStore, reminderService)
	componentHandler := handlers.NewComponentHandler(reminderService)

	b := &Bot{
		session:         session,
		config:          cfg,
		store:           emojiStore,
		registerCmd:     registerCommand,
		removeCmd:       removeCommand,
		listCmd:         listCommand,
		helpCmd:         helpCommand,
		reactionHandle:  reactionHandler,
		componentHandle: componentHandler,
		reminders:       reminderService,
	}

	session.AddHandler(b.onInteraction)
	session.AddHandler(reactionHandler.Handle)
	session.AddHandler(componentHandler.Handle)

	session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildMessageReactions | discordgo.IntentsDirectMessages

	return b, nil
}

// Open registers slash commands and opens the Discord websocket connection.
func (b *Bot) Open() error {
	if err := b.registerCommands(); err != nil {
		return err
	}

	if err := b.session.Open(); err != nil {
		return err
	}

	log.Println("bot is running. Press CTRL-C to exit")
	return nil
}

// Close cleans up resources and unregisters commands.
func (b *Bot) Close() error {
	if b.reminders != nil {
		// Ensure no reminders fire after shutdown.
		b.reminders.Close()
	}
	if len(b.commandIDs) > 0 {
		for _, id := range b.commandIDs {
			if err := b.session.ApplicationCommandDelete(b.config.AppID, b.config.GuildID, id); err != nil {
				log.Printf("failed to delete command %s: %v", id, err)
			}
		}
	}

	if err := b.session.Close(); err != nil {
		return err
	}

	return nil
}

func (b *Bot) registerCommands() error {
	definitions := []*discordgo.ApplicationCommand{
		b.registerCmd.Definition(),
		b.removeCmd.Definition(),
		b.listCmd.Definition(),
		b.helpCmd.Definition(),
	}

	for _, cmd := range definitions {
		created, err := b.session.ApplicationCommandCreate(b.config.AppID, b.config.GuildID, cmd)
		if err != nil {
			return err
		}

		b.commandIDs = append(b.commandIDs, created.ID)
	}

	return nil
}

func (b *Bot) onInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		var err error
		switch i.ApplicationCommandData().Name {
		case commands.SetBookmarkCommandName:
			err = b.registerCmd.Handle(s, i)
		case commands.RemoveBookmarkCommandName:
			err = b.removeCmd.Handle(s, i)
		case commands.ListBookmarksCommandName:
			err = b.listCmd.Handle(s, i)
		case commands.HelpCommandName:
			err = b.helpCmd.Handle(s, i)
		}

		if err != nil {
			log.Printf("command error: %v", err)
			// Try to send error message to user
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ Error: " + err.Error(),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
	case discordgo.InteractionMessageComponent:
		if b.componentHandle != nil {
			b.componentHandle.Handle(s, i)
		}
	}
}
