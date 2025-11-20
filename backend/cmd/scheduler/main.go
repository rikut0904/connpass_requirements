package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"connpass-requirement/internal/config"
	"connpass-requirement/internal/database"
	"connpass-requirement/internal/repository"
	"connpass-requirement/internal/services"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := database.Connect(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer db.Close()

	ruleRepo := repository.NewRuleRepository(db)
	eventRepo := repository.NewEventRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	logRepo := repository.NewLogRepository(db)

	logger := services.NewLoggerService(logRepo)
	connpass := services.NewConnpassService(cfg)

	discordService, err := services.NewDiscordService(cfg.DiscordBotToken)
	if err != nil {
		log.Fatalf("failed to init discord service: %v", err)
	}
	if err := discordService.Open(); err != nil {
		log.Fatalf("failed to open discord session: %v", err)
	}
	defer discordService.Close()

	notifier := services.NewNotifierService(notificationRepo, eventRepo, discordService, logger, cfg.NotificationDefaultLimit)
	scheduler := services.NewSchedulerService(ruleRepo, notificationRepo, eventRepo, logRepo, connpass, notifier, logger)

	if err := scheduler.Run(ctx); err != nil {
		log.Printf("scheduler run failed: %v", err)
	}
}
