package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"connpass-requirement/internal/models"
	"connpass-requirement/internal/repository"
)

// NotifierService は通知条件の判定とDiscord送信を担当する。
type NotifierService struct {
	notificationRepo *repository.NotificationRepository
	eventRepo        *repository.EventRepository
	discord          *DiscordService
	logger           *LoggerService
	defaultThreshold int
}

func NewNotifierService(
	notificationRepo *repository.NotificationRepository,
	eventRepo *repository.EventRepository,
	discord *DiscordService,
	logger *LoggerService,
	defaultThreshold int,
) *NotifierService {
	return &NotifierService{
		notificationRepo: notificationRepo,
		eventRepo:        eventRepo,
		discord:          discord,
		logger:           logger,
		defaultThreshold: defaultThreshold,
	}
}

// Evaluate は通知対象のトリガーを返す。
func (n *NotifierService) Evaluate(rule models.Rule, event models.Event, prev *models.Event) []string {
	var targets []string
	for _, notifyType := range rule.NotifyTypes {
		switch notifyType {
		case "open":
			if prev == nil {
				targets = append(targets, notifyType)
			}
		case "start":
			if withinWindow(event.StartedAt, 30*time.Minute) {
				targets = append(targets, notifyType)
			}
		case "almost_full":
			threshold := rule.CapacityThresh
			if threshold == 0 {
				threshold = n.defaultThreshold
			}
			if event.Limit > 0 {
				rate := float64(event.Accepted) / float64(event.Limit) * 100
				if int(rate+0.5) >= threshold {
					targets = append(targets, notifyType)
				}
			}
		case "before_deadline":
			deadline := event.EndedAt.Add(-1 * time.Hour)
			if withinWindow(deadline, 30*time.Minute) {
				targets = append(targets, notifyType)
			}
		}
	}
	return targets
}

// Notify はDiscordへの通知と履歴登録を行う。
func (n *NotifierService) Notify(ctx context.Context, rule models.Rule, event models.Event, notifyKey string) error {
	exists, err := n.notificationRepo.Exists(ctx, rule.ID, event.EventID, notifyKey)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	message := buildMessage(rule, event, notifyKey)
	if err := n.discord.SendMessage(ctx, rule.ChannelID, message); err != nil {
		n.logger.Error(ctx, "discord_send_failed", err.Error(), map[string]any{
			"ruleId":    rule.ID,
			"eventId":   event.EventID,
			"notifyKey": notifyKey,
		})
		return err
	}

	if err := n.notificationRepo.Record(ctx, rule.ID, event.EventID, notifyKey); err != nil {
		return err
	}

	n.logger.Info(ctx, "notification_sent", "Discord通知を送信しました", map[string]any{
		"ruleId":    rule.ID,
		"eventId":   event.EventID,
		"notifyKey": notifyKey,
	})

	return nil
}

func withinWindow(target time.Time, window time.Duration) bool {
	if target.IsZero() {
		return false
	}
	now := time.Now()
	return target.After(now.Add(-window)) && target.Before(now.Add(window))
}

func buildMessage(rule models.Rule, event models.Event, notifyKey string) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("**%s**\n", event.Title))
	builder.WriteString(fmt.Sprintf("イベントURL: %s\n", event.EventURL))
	builder.WriteString(fmt.Sprintf("開始: %s\n終了: %s\n", event.StartedAt.Format(time.RFC1123), event.EndedAt.Format(time.RFC1123)))
	builder.WriteString(fmt.Sprintf("参加者: %d / %d (待機 %d)\n", event.Accepted, event.Limit, event.Waiting))
	builder.WriteString(fmt.Sprintf("トリガー: %s\n", notifyKey))
	builder.WriteString(fmt.Sprintf("ルール: %s\n", rule.Name))
	if rule.Description != "" {
		builder.WriteString(rule.Description + "\n")
	}
	return builder.String()
}
