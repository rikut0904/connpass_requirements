package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/labstack/echo/v4"

	"connpass-requirement/internal/models"
	"connpass-requirement/internal/repository"
	"connpass-requirement/internal/services"
)

// GuildHandler はギルド／チャンネル関連API。
type GuildHandler struct {
	users   *repository.UserRepository
	discord *services.DiscordService
}

func NewGuildHandler(users *repository.UserRepository, discord *services.DiscordService) *GuildHandler {
	return &GuildHandler{users: users, discord: discord}
}

func RegisterGuildRoutes(g *echo.Group, handler *GuildHandler) {
	g.GET("/me/guilds", handler.ListGuilds)
	g.GET("/guilds/:guildId/channels", handler.ListChannels)
	g.POST("/guilds/:guildId/channels", handler.CreateChannel)
}

func (h *GuildHandler) ListGuilds(c echo.Context) error {
	userID := MustUserID(c)
	guilds, err := h.users.ListGuildPermissions(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load guilds")
	}
	if h.discord != nil {
		filtered := make([]models.GuildPermission, 0, len(guilds))
		for _, guild := range guilds {
			ok, err := h.discord.IsBotInGuild(c.Request().Context(), guild.GuildID)
			if err != nil {
				continue
			}
			if ok {
				filtered = append(filtered, guild)
			}
		}
		guilds = filtered
	}
	return c.JSON(http.StatusOK, guilds)
}

func (h *GuildHandler) ListChannels(c echo.Context) error {
	if h.discord == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "discord integration is disabled")
	}
	userID := MustUserID(c)
	guildID := c.Param("guildId")
	if guildID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "guildId is required")
	}

	guilds, err := h.users.ListGuildPermissions(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to verify guild access")
	}

	if !canManageGuild(guilds, guildID) {
		return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
	}

	textChannels, categories, err := h.discord.ListTextChannelsWithCategories(c.Request().Context(), guildID)
	if err != nil {
		if errors.Is(err, services.ErrMissingAccess) {
			return echo.NewHTTPError(http.StatusForbidden, "bot does not have access to this guild")
		}
		return echo.NewHTTPError(http.StatusBadGateway, "failed to fetch channels")
	}

	type channelView struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		CategoryID   string `json:"categoryId,omitempty"`
		CategoryName string `json:"categoryName,omitempty"`
	}

	type categoryView struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	resp := struct {
		Channels   []channelView  `json:"channels"`
		Categories []categoryView `json:"categories"`
	}{
		Channels:   make([]channelView, 0, len(textChannels)),
		Categories: make([]categoryView, 0, len(categories)),
	}

	categoryMap := make(map[string]string, len(categories))
	for _, cat := range categories {
		resp.Categories = append(resp.Categories, categoryView{ID: cat.ID, Name: cat.Name})
		categoryMap[cat.ID] = cat.Name
	}

	for _, ch := range textChannels {
		view := channelView{ID: ch.ID, Name: ch.Name}
		if ch.ParentID != "" {
			view.CategoryID = ch.ParentID
			if name, ok := categoryMap[ch.ParentID]; ok {
				view.CategoryName = name
			}
		}
		resp.Channels = append(resp.Channels, view)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *GuildHandler) CreateChannel(c echo.Context) error {
	if h.discord == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "discord integration is disabled")
	}

	userID := MustUserID(c)
	guildID := c.Param("guildId")
	if strings.TrimSpace(guildID) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "guildId is required")
	}

	var req struct {
		Name         string `json:"name"`
		CategoryID   string `json:"categoryId"`
		CategoryName string `json:"categoryName"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "channel name is required")
	}
	categoryID := strings.TrimSpace(req.CategoryID)
	categoryName := strings.TrimSpace(req.CategoryName)

	guilds, err := h.users.ListGuildPermissions(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to verify guild access")
	}

	if !canManageGuild(guilds, guildID) {
		return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
	}

	var createdCategory *discordgo.Channel
	if categoryName != "" {
		createdCategory, err = h.discord.CreateCategory(c.Request().Context(), guildID, categoryName)
		if err != nil {
			if errors.Is(err, services.ErrMissingAccess) {
				return echo.NewHTTPError(http.StatusForbidden, "bot does not have permissions to create channels")
			}
			return echo.NewHTTPError(http.StatusBadGateway, "failed to create category")
		}
		categoryID = createdCategory.ID
	}

	channel, err := h.discord.CreateTextChannel(c.Request().Context(), guildID, name, categoryID)
	if err != nil {
		if errors.Is(err, services.ErrMissingAccess) {
			return echo.NewHTTPError(http.StatusForbidden, "bot does not have permissions to create channels")
		}
		return echo.NewHTTPError(http.StatusBadGateway, "failed to create channel")
	}

	resp := map[string]any{
		"id":         channel.ID,
		"name":       channel.Name,
		"categoryId": channel.ParentID,
	}
	if channel.ParentID == "" {
		delete(resp, "categoryId")
	}
	if createdCategory != nil {
		resp["category"] = map[string]any{
			"id":   createdCategory.ID,
			"name": createdCategory.Name,
		}
	}

	return c.JSON(http.StatusCreated, resp)
}

func canManageGuild(guilds []models.GuildPermission, guildID string) bool {
	for _, guild := range guilds {
		if guild.GuildID == guildID {
			return guild.CanManage
		}
	}
	return false
}
