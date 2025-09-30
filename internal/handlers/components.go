package handlers

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

// CloseButtonID identifies the component button that closes a forwarded message.
const CloseButtonID = "close_dm"

// ComponentHandler processes interactions originating from message components.
func ComponentHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	switch i.MessageComponentData().CustomID {
	case CloseButtonID:
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredMessageUpdate}); err != nil {
			log.Printf("failed to acknowledge close button interaction: %v", err)
			return
		}

		if err := s.ChannelMessageDelete(i.ChannelID, i.Message.ID); err != nil {
			log.Printf("failed to delete forwarded message: %v", err)
		}
	}
}
