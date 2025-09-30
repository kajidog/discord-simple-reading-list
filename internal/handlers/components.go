package handlers

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

// CompleteButtonID identifies the button that marks a lightweight bookmark as complete.
const CompleteButtonID = "bookmark_complete"

// DeleteButtonID identifies the button that deletes a saved bookmark message.
const DeleteButtonID = "bookmark_delete"

// ComponentHandler processes interactions originating from message components.
func ComponentHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	customID := i.MessageComponentData().CustomID
	switch customID {
	case CompleteButtonID, DeleteButtonID:
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredMessageUpdate}); err != nil {
			log.Printf("failed to acknowledge bookmark component interaction: %v", err)
			return
		}

		if err := s.ChannelMessageDelete(i.ChannelID, i.Message.ID); err != nil {
			log.Printf("failed to delete bookmarked message: %v", err)
		}
	}
}
