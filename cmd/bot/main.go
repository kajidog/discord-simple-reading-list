package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/example/discord-simple-reading-list/internal/bot"
	"github.com/example/discord-simple-reading-list/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	b, err := bot.New(cfg)
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	if err := b.Open(); err != nil {
		log.Fatalf("failed to open bot connection: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	if err := b.Close(); err != nil {
		log.Fatalf("failed to close bot: %v", err)
	}
}
