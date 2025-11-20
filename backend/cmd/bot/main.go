package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"

	"connpass-requirement/internal/config"
	"connpass-requirement/internal/services"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	discordService, err := services.NewDiscordService(cfg.DiscordBotToken)
	if err != nil {
		log.Fatalf("failed to init discord service: %v", err)
	}

	if err := discordService.Open(); err != nil {
		log.Fatalf("failed to open discord session: %v", err)
	}
	defer discordService.Close()

	discordService.Session().AddHandler(messageCreateHandler)

	<-ctx.Done()
}

func messageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}
	if m.Content == "!ping" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "pong")
	}
}
