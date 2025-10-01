package bot

import (
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/example/discord-simple-reading-list/internal/commands"
	"github.com/example/discord-simple-reading-list/internal/config"
	"github.com/example/discord-simple-reading-list/internal/handlers"
	"github.com/example/discord-simple-reading-list/internal/store"
)

// Bot encapsulates the Discord session and all registered handlers.
type Bot struct {
	session        *discordgo.Session
	config         *config.Config
	store          *store.EmojiStore
	registerCmd    *commands.SetBookmarkCommand
	listCmd        *commands.ListBookmarksCommand
	helpCmd        *commands.HelpCommand
	reactionHandle *handlers.ReactionHandler
	commandIDs     []string
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

	registerCommand := commands.NewSetBookmarkCommand(emojiStore)
	listCommand := commands.NewListBookmarksCommand(emojiStore)
	helpCommand := commands.NewHelpCommand()
	reactionHandler := handlers.NewReactionHandler(emojiStore)

	b := &Bot{
		session:        session,
		config:         cfg,
		store:          emojiStore,
		registerCmd:    registerCommand,
		listCmd:        listCommand,
		helpCmd:        helpCommand,
		reactionHandle: reactionHandler,
	}

	session.AddHandler(b.onInteraction)
	session.AddHandler(reactionHandler.Handle)

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
		switch i.ApplicationCommandData().Name {
		case commands.SetBookmarkCommandName:
			if err := b.registerCmd.Handle(s, i); err != nil {
				log.Printf("failed to handle register command: %v", err)
			}
		case commands.ListBookmarksCommandName:
			if err := b.listCmd.Handle(s, i); err != nil {
				log.Printf("failed to handle list command: %v", err)
			}
		case commands.HelpCommandName:
			if err := b.helpCmd.Handle(s, i); err != nil {
				log.Printf("failed to handle help command: %v", err)
			}
		}
	case discordgo.InteractionMessageComponent:
		handlers.ComponentHandler(s, i)
	}
}
