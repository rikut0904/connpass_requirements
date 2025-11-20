package services

import (
	"context"
	"encoding/json"
	"time"

	"connpass-requirement/internal/models"
	"connpass-requirement/internal/repository"
)

// LoggerService は重要ログ記録を統一的に扱う。
type LoggerService struct {
	repo *repository.LogRepository
}

func NewLoggerService(repo *repository.LogRepository) *LoggerService {
	return &LoggerService{repo: repo}
}

func (l *LoggerService) log(ctx context.Context, level, eventType, message string, metadata any) {
	metaJSON, err := json.Marshal(metadata)
	if err != nil {
		metaJSON = []byte(`{"error":"metadata marshal failed"}`)
	}

	_ = l.repo.Save(ctx, models.ImportantLog{
		Level:     level,
		EventType: eventType,
		Message:   message,
		Metadata:  string(metaJSON),
	})
}

func (l *LoggerService) Info(ctx context.Context, eventType, message string, metadata any) {
	l.log(ctx, "INFO", eventType, message, metadata)
}

func (l *LoggerService) Warn(ctx context.Context, eventType, message string, metadata any) {
	l.log(ctx, "WARNING", eventType, message, metadata)
}

func (l *LoggerService) Error(ctx context.Context, eventType, message string, metadata any) {
	l.log(ctx, "ERROR", eventType, message, metadata)
}

func (l *LoggerService) UpdateSchedulerStatus(ctx context.Context, lastRun time.Time, lastError string) {
	_ = l.repo.UpdateSchedulerStatus(ctx, models.SchedulerStatus{
		LastRunAt: lastRun,
		LastError: lastError,
	})
}
