package services

import (
	"context"
	"fmt"
	"time"

	"connpass-requirement/internal/repository"
)

// SchedulerService は30分毎に実行されるジョブを実装する。
type SchedulerService struct {
	ruleRepo         *repository.RuleRepository
	notificationRepo *repository.NotificationRepository
	eventRepo        *repository.EventRepository
	logRepo          *repository.LogRepository
	connpass         *ConnpassService
	notifier         *NotifierService
	logger           *LoggerService
}

func NewSchedulerService(
	ruleRepo *repository.RuleRepository,
	notificationRepo *repository.NotificationRepository,
	eventRepo *repository.EventRepository,
	logRepo *repository.LogRepository,
	connpass *ConnpassService,
	notifier *NotifierService,
	logger *LoggerService,
) *SchedulerService {
	return &SchedulerService{
		ruleRepo:         ruleRepo,
		notificationRepo: notificationRepo,
		eventRepo:        eventRepo,
		logRepo:          logRepo,
		connpass:         connpass,
		notifier:         notifier,
		logger:           logger,
	}
}

// Run はスケジュール処理を実行する。
func (s *SchedulerService) Run(ctx context.Context) error {
	s.logger.Info(ctx, "scheduler_start", "スケジューラを開始", nil)
	start := time.Now()

	rules, err := s.ruleRepo.ListActive(ctx)
	if err != nil {
		s.logger.Error(ctx, "database_error", "ルール一覧の取得に失敗", err)
		s.logger.UpdateSchedulerStatus(ctx, start, err.Error())
		return err
	}

	s.logger.Info(ctx, "scheduler_processing", fmt.Sprintf("処理するルール数: %d", len(rules)), nil)

	for _, rule := range rules {
		if len(rule.Keywords) == 0 {
			s.logger.Info(ctx, "rule_skip", "キーワードが未設定のためスキップ", map[string]any{"ruleId": rule.ID, "ruleName": rule.Name})
			continue
		}
		s.logger.Info(ctx, "rule_process", "ルールを処理中", map[string]any{"ruleId": rule.ID, "ruleName": rule.Name, "keywords": rule.Keywords})

		for _, keyword := range rule.Keywords {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			events, err := s.connpass.FetchEvents(ctx, keyword, rule.Location)
			if err != nil {
				s.logger.Error(ctx, "connpass_api_error", "connpass API取得に失敗", map[string]any{"keyword": keyword, "error": err.Error()})
				continue
			}

			s.logger.Info(ctx, "connpass_fetch", fmt.Sprintf("connpassから%d件のイベントを取得", len(events)), map[string]any{"keyword": keyword, "location": rule.Location})

			for _, event := range events {
				prev, err := s.eventRepo.FindByEventID(ctx, event.EventID)
				if err != nil {
					s.logger.Error(ctx, "database_error", "イベントキャッシュ取得に失敗", err)
					continue
				}

				if err := s.eventRepo.Upsert(ctx, &event); err != nil {
					s.logger.Error(ctx, "database_error", "イベントキャッシュ保存に失敗", err)
					continue
				}

				triggers := s.notifier.Evaluate(rule, event, prev)
				if len(triggers) > 0 {
					s.logger.Info(ctx, "notification_trigger", fmt.Sprintf("%d件の通知トリガーを検出", len(triggers)), map[string]any{
						"eventId":  event.EventID,
						"title":    event.Title,
						"triggers": triggers,
					})
				}
				for _, notifyKey := range triggers {
					if err := s.notifier.Notify(ctx, rule, event, notifyKey); err != nil {
						continue
					}
				}
			}
		}
	}

	cleanupBefore := time.Now().Add(-14 * 24 * time.Hour)
	_ = s.eventRepo.Cleanup(ctx, cleanupBefore)
	_ = s.notificationRepo.Cleanup(ctx, cleanupBefore)
	_ = s.logRepo.Cleanup(ctx, time.Now().Add(-90*24*time.Hour))

	s.logger.UpdateSchedulerStatus(ctx, time.Now(), "")
	s.logger.Info(ctx, "scheduler_complete", "スケジューラが正常終了", map[string]any{"duration": time.Since(start).String()})

	return nil
}
