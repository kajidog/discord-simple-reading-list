package handlers

import (
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/example/discord-bookmark-manager/internal/reminders"
)

// CompleteButtonID identifies the button that marks a lightweight bookmark as complete.
const CompleteButtonID = "bookmark_complete"

// DeleteButtonID identifies the button that deletes a saved bookmark message.
const DeleteButtonID = "bookmark_delete"

// ComponentHandler processes interactions originating from message components.
type ComponentHandler struct {
	reminders *reminders.Service
}

// NewComponentHandler constructs a component handler instance.
func NewComponentHandler(reminders *reminders.Service) *ComponentHandler {
	return &ComponentHandler{reminders: reminders}
}

// Handle reacts to button presses on bookmarked messages.
func (h *ComponentHandler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

		if h.reminders != nil {
			if customID == DeleteButtonID {
				h.reminders.Cancel(i.Message.ID)
			} else {
				h.reminders.Complete(i.Message.ID)
			}
		}
	}
}
