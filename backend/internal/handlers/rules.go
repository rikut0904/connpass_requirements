package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"connpass-requirement/internal/models"
	"connpass-requirement/internal/repository"
	"connpass-requirement/internal/services"
)

// RuleHandler は通知ルール操作API。
type RuleHandler struct {
	rules   *repository.RuleRepository
	users   *repository.UserRepository
	logger  *services.LoggerService
	discord *services.DiscordService
}

func NewRuleHandler(rules *repository.RuleRepository, users *repository.UserRepository, logger *services.LoggerService, discord *services.DiscordService) *RuleHandler {
	return &RuleHandler{rules: rules, users: users, logger: logger, discord: discord}
}

// RegisterRuleRoutes はルール関連のルートを登録する。
func RegisterRuleRoutes(g *echo.Group, handler *RuleHandler) {
	g.GET("/rules", handler.List)
	g.POST("/rules", handler.Create)
	g.GET("/rules/:id", handler.Get)
	g.PUT("/rules/:id", handler.Update)
	g.DELETE("/rules/:id", handler.Delete)
	g.POST("/rules/:id/test", handler.Test)
}

func (h *RuleHandler) List(c echo.Context) error {
	userID := MustUserID(c)
	guildID := c.QueryParam("guild_id")
	if guildID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "guild_id is required")
	}

	rules, err := h.rules.ListByUserAndGuild(c.Request().Context(), userID, guildID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch rules")
	}

	return c.JSON(http.StatusOK, rules)
}

type rulePayload struct {
	GuildID        string   `json:"guildId"`
	ChannelID      string   `json:"channelId"`
	ChannelName    string   `json:"channelName"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Location       string   `json:"location"`
	CapacityThresh int      `json:"capacityThreshold"`
	Keywords       []string `json:"keywords"`
	NotifyTypes    []string `json:"notifyTypes"`
	IsActive       bool     `json:"isActive"`
}

func (h *RuleHandler) Create(c echo.Context) error {
	userID := MustUserID(c)
	var payload rulePayload
	if err := c.Bind(&payload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	if err := h.ensureGuildPermission(c, userID, payload.GuildID); err != nil {
		return err
	}

	rule := models.Rule{
		UserID:         userID,
		GuildID:        payload.GuildID,
		ChannelID:      payload.ChannelID,
		ChannelName:    payload.ChannelName,
		Name:           strings.TrimSpace(payload.Name),
		Description:    strings.TrimSpace(payload.Description),
		Location:       strings.TrimSpace(payload.Location),
		CapacityThresh: payload.CapacityThresh,
		Keywords:       payload.Keywords,
		NotifyTypes:    payload.NotifyTypes,
		IsActive:       payload.IsActive,
	}

	if err := h.rules.Create(c.Request().Context(), &rule); err != nil {
		h.logger.Error(c.Request().Context(), "database_error", "ルール作成に失敗", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create rule")
	}

	return c.JSON(http.StatusCreated, rule)
}

func (h *RuleHandler) Get(c echo.Context) error {
	ruleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	rule, err := h.rules.Get(c.Request().Context(), ruleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch rule")
	}
	if rule == nil {
		return echo.NewHTTPError(http.StatusNotFound, "rule not found")
	}

	return c.JSON(http.StatusOK, rule)
}

func (h *RuleHandler) Update(c echo.Context) error {
	userID := MustUserID(c)
	ruleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	var payload rulePayload
	if err := c.Bind(&payload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	rule, err := h.rules.Get(c.Request().Context(), ruleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch rule")
	}
	if rule == nil {
		return echo.NewHTTPError(http.StatusNotFound, "rule not found")
	}
	if rule.UserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "permission denied")
	}

	if err := h.ensureGuildPermission(c, userID, payload.GuildID); err != nil {
		return err
	}

	rule.GuildID = payload.GuildID
	rule.ChannelID = payload.ChannelID
	rule.ChannelName = payload.ChannelName
	rule.Name = strings.TrimSpace(payload.Name)
	rule.Description = strings.TrimSpace(payload.Description)
	rule.Location = strings.TrimSpace(payload.Location)
	rule.CapacityThresh = payload.CapacityThresh
	rule.Keywords = payload.Keywords
	rule.NotifyTypes = payload.NotifyTypes
	rule.IsActive = payload.IsActive

	if err := h.rules.Update(c.Request().Context(), rule); err != nil {
		h.logger.Error(c.Request().Context(), "database_error", "ルール更新に失敗", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update rule")
	}

	return c.JSON(http.StatusOK, rule)
}

func (h *RuleHandler) Delete(c echo.Context) error {
	userID := MustUserID(c)
	ruleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	rule, err := h.rules.Get(c.Request().Context(), ruleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch rule")
	}
	if rule == nil {
		return echo.NewHTTPError(http.StatusNotFound, "rule not found")
	}
	if rule.UserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "permission denied")
	}

	if err := h.rules.Delete(c.Request().Context(), ruleID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete rule")
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *RuleHandler) Test(c echo.Context) error {
	if h.discord == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "discord integration is disabled")
	}
	userID := MustUserID(c)
	ruleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	rule, err := h.rules.Get(c.Request().Context(), ruleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch rule")
	}
	if rule == nil {
		return echo.NewHTTPError(http.StatusNotFound, "rule not found")
	}
	if rule.UserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "permission denied")
	}

	message := "テスト通知です。Discord Botの接続とチャンネル権限を確認しました。"
	if err := h.discord.SendMessage(c.Request().Context(), rule.ChannelID, message); err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "failed to send test notification")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "テスト通知を送信しました"})
}

func (h *RuleHandler) ensureGuildPermission(c echo.Context, userID int64, guildID string) error {
	if guildID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "guildId is required")
	}

	guilds, err := h.users.ListGuildPermissions(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to verify guild")
	}

	for _, guild := range guilds {
		if guild.GuildID == guildID {
			if guild.CanManage {
				return nil
			}
			return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
		}
	}

	return echo.NewHTTPError(http.StatusForbidden, "guild access denied")
}
