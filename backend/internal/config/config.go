package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config はサービス全体で利用する設定値を集約する。
type Config struct {
	Port                     int
	DatabaseURL              string
	JWTSecret                string
	DiscordClientID          string
	DiscordClientSecret      string
	DiscordRedirectURI       string
	DiscordBotToken          string
	DiscordPublicKey         string
	ConnpassBaseURL          string
	ConnpassAPIKey           string
	ConnpassRequestInterval  time.Duration
	NotificationDefaultLimit int
	SchedulerInterval        time.Duration
	SessionMode              string
	SessionDuration          time.Duration
}

// Load は環境変数から設定値を読み込み、バリデーションを行う。
func Load() (Config, error) {
	cfg := Config{}

	portStr := getEnv("PORT", "8080")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return cfg, fmt.Errorf("invalid PORT: %w", err)
	}
	cfg.Port = port

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	cfg.DiscordClientID = os.Getenv("DISCORD_CLIENT_ID")
	cfg.DiscordClientSecret = os.Getenv("DISCORD_CLIENT_SECRET")
	cfg.DiscordRedirectURI = os.Getenv("DISCORD_REDIRECT_URI")
	cfg.DiscordBotToken = os.Getenv("DISCORD_BOT_TOKEN")
	cfg.DiscordPublicKey = os.Getenv("DISCORD_PUBLIC_KEY")
	cfg.ConnpassBaseURL = getEnv("CONNPASS_BASE_URL", "https://connpass.com/api/v2/events/")
	cfg.ConnpassAPIKey = os.Getenv("CONNPASS_API_KEY")

	requestIntervalStr := getEnv("CONNPASS_REQUEST_INTERVAL", "1s")
	requestInterval, err := time.ParseDuration(requestIntervalStr)
	if err != nil {
		return cfg, fmt.Errorf("invalid CONNPASS_REQUEST_INTERVAL: %w", err)
	}
	cfg.ConnpassRequestInterval = requestInterval

	notificationLimitStr := getEnv("NOTIFICATION_DEFAULT_THRESHOLD", "80")
	notificationLimit, err := strconv.Atoi(notificationLimitStr)
	if err != nil {
		return cfg, fmt.Errorf("invalid NOTIFICATION_DEFAULT_THRESHOLD: %w", err)
	}
	cfg.NotificationDefaultLimit = notificationLimit

	schedulerIntervalStr := getEnv("SCHEDULER_POLL_INTERVAL", "30m")
	schedulerInterval, err := time.ParseDuration(schedulerIntervalStr)
	if err != nil {
		return cfg, fmt.Errorf("invalid SCHEDULER_POLL_INTERVAL: %w", err)
	}
	cfg.SchedulerInterval = schedulerInterval

	// セッションモード: develop=1分, production=3ヶ月
	cfg.SessionMode = getEnv("SESSION_MODE", "production")
	if cfg.SessionMode == "develop" {
		cfg.SessionDuration = 1 * time.Minute
	} else {
		cfg.SessionDuration = 90 * 24 * time.Hour // 3ヶ月
	}

	if cfg.DatabaseURL == "" {
		return cfg, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return cfg, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.DiscordClientID == "" || cfg.DiscordClientSecret == "" {
		return cfg, fmt.Errorf("discord client credentials are required")
	}
	if cfg.DiscordRedirectURI == "" {
		return cfg, fmt.Errorf("DISCORD_REDIRECT_URI is required")
	}
	if cfg.ConnpassAPIKey == "" {
		return cfg, fmt.Errorf("CONNPASS_API_KEY is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
