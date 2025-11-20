package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"connpass-requirement/internal/config"
	"connpass-requirement/internal/database"
	"connpass-requirement/internal/handlers"
	"connpass-requirement/internal/repository"
	"connpass-requirement/internal/services"
	"connpass-requirement/migrations"
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

	if err := migrations.Run(ctx, db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	ruleRepo := repository.NewRuleRepository(db)
	logRepo := repository.NewLogRepository(db)
	eventRepo := repository.NewEventRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)

	oauthService := services.NewOAuthService(cfg)
	loggerService := services.NewLoggerService(logRepo)
	connpassService := services.NewConnpassService(cfg)

	var discordService *services.DiscordService
	if cfg.DiscordBotToken != "" {
		if svc, err := services.NewDiscordService(cfg.DiscordBotToken); err == nil {
			discordService = svc
			if err := svc.Open(); err != nil {
				log.Printf("failed to open discord session: %v", err)
			}
		} else {
			log.Printf("failed to init discord service: %v", err)
		}
	} else {
		log.Printf("DISCORD_BOT_TOKEN is not set. Channel listing and test notification APIs are disabled")
	}

	var notifierService *services.NotifierService
	var schedulerService *services.SchedulerService
	if discordService != nil {
		notifierService = services.NewNotifierService(notificationRepo, eventRepo, discordService, loggerService, cfg.NotificationDefaultLimit)
		schedulerService = services.NewSchedulerService(ruleRepo, notificationRepo, eventRepo, logRepo, connpassService, notifierService, loggerService)
	}

	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = customHTTPErrorHandler
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	e.GET("/healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	api := e.Group("/api")
	authHandler := handlers.NewAuthHandler(cfg, oauthService, userRepo, loggerService)
	handlers.RegisterAuthRoutes(api, authHandler)

	authenticated := api.Group("")
	authenticated.Use(handlers.JWTMiddleware(cfg))

	handlers.RegisterAuthRoutesWithMiddleware(authenticated, authHandler)
	handlers.RegisterGuildRoutes(authenticated, handlers.NewGuildHandler(userRepo, discordService))
	handlers.RegisterRuleRoutes(authenticated, handlers.NewRuleHandler(ruleRepo, userRepo, loggerService, discordService))
	handlers.RegisterStatusRoutes(authenticated, handlers.NewStatusHandler(logRepo))
	handlers.RegisterLogRoutes(authenticated, handlers.NewLogHandler(logRepo))
	if schedulerService != nil {
		handlers.RegisterSchedulerRoutes(authenticated, handlers.NewSchedulerHandler(schedulerService))
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: e,
	}

	go func() {
		if err := e.StartServer(server); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}

func customHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	if he, ok := err.(*echo.HTTPError); ok {
		_ = c.JSON(he.Code, map[string]any{
			"message": he.Message,
		})
		return
	}

	_ = c.JSON(http.StatusInternalServerError, map[string]any{
		"message": err.Error(),
	})
}
