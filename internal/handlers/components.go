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
	case CompleteButtonID:
		// Mark as complete: dim the message and disable buttons
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredMessageUpdate}); err != nil {
			log.Printf("failed to acknowledge complete interaction: %v", err)
			return
		}

		if i.Message != nil && len(i.Message.Embeds) > 0 {
			// Clone embeds and reduce opacity by making color dimmer
			updatedEmbeds := make([]*discordgo.MessageEmbed, len(i.Message.Embeds))
			for idx, embed := range i.Message.Embeds {
				if embed == nil {
					continue
				}
				cloned := cloneEmbedForComplete(embed)
				// Add ✅ prefix to title to indicate completion
				if cloned.Title != "" {
					cloned.Title = "✅ " + cloned.Title
				}
				// Dim the color (make it grayer)
				if cloned.Color != 0 {
					cloned.Color = 0x808080 // Gray color
				}
				updatedEmbeds[idx] = cloned
			}

			// Remove all buttons
			_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
				Channel:    i.ChannelID,
				ID:         i.Message.ID,
				Embeds:     updatedEmbeds,
				Components: []discordgo.MessageComponent{},
			})
			if err != nil {
				log.Printf("failed to update completed bookmark: %v", err)
			}
		}

		if h.reminders != nil {
			h.reminders.Complete(i.Message.ID)
		}

	case DeleteButtonID:
		// Delete the message completely
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredMessageUpdate}); err != nil {
			log.Printf("failed to acknowledge delete interaction: %v", err)
			return
		}

		if err := s.ChannelMessageDelete(i.ChannelID, i.Message.ID); err != nil {
			log.Printf("failed to delete bookmarked message: %v", err)
		}

		if h.reminders != nil {
			h.reminders.Cancel(i.Message.ID)
		}
	}
}

func cloneEmbedForComplete(embed *discordgo.MessageEmbed) *discordgo.MessageEmbed {
	if embed == nil {
		return nil
	}

	cloned := *embed

	if embed.Fields != nil {
		cloned.Fields = make([]*discordgo.MessageEmbedField, len(embed.Fields))
		for i, field := range embed.Fields {
			if field == nil {
				continue
			}
			copied := *field
			cloned.Fields[i] = &copied
		}
	}

	if embed.Author != nil {
		copied := *embed.Author
		cloned.Author = &copied
	}

	if embed.Footer != nil {
		copied := *embed.Footer
		cloned.Footer = &copied
	}

	if embed.Image != nil {
		copied := *embed.Image
		cloned.Image = &copied
	}

	if embed.Thumbnail != nil {
		copied := *embed.Thumbnail
		cloned.Thumbnail = &copied
	}

	if embed.Provider != nil {
		copied := *embed.Provider
		cloned.Provider = &copied
	}

	if embed.Video != nil {
		copied := *embed.Video
		cloned.Video = &copied
	}

	return &cloned
}
