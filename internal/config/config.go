package config

import (
	"fmt"
	"os"
)

// Config holds runtime configuration values loaded from environment variables.
type Config struct {
	BotToken  string
	AppID     string
	GuildID   string
	StorePath string
}

// Load reads configuration from environment variables and validates that the required
// values are provided.
func Load() (*Config, error) {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN is required")
	}

	appID := os.Getenv("DISCORD_APP_ID")
	if appID == "" {
		return nil, fmt.Errorf("DISCORD_APP_ID is required")
	}

	guildID := os.Getenv("DISCORD_GUILD_ID")

	storePath := os.Getenv("BOOKMARK_STORE_PATH")
	if storePath == "" {
		storePath = "bookmarks.json"
	}

	return &Config{
		BotToken:  token,
		AppID:     appID,
		GuildID:   guildID,
		StorePath: storePath,
	}, nil
}
