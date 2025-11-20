package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"connpass-requirement/internal/repository"
)

// LogHandler は重要ログ一覧API。
type LogHandler struct {
	logs *repository.LogRepository
}

func NewLogHandler(logs *repository.LogRepository) *LogHandler {
	return &LogHandler{logs: logs}
}

// RegisterLogRoutes はログ関連ルートを登録。
func RegisterLogRoutes(g *echo.Group, handler *LogHandler) {
	g.GET("/logs", handler.List)
}

func (h *LogHandler) List(c echo.Context) error {
	limitStr := c.QueryParam("limit")
	limit := 50
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
			limit = v
		}
	}

	logs, err := h.logs.ListRecent(c.Request().Context(), limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch logs")
	}

	return c.JSON(http.StatusOK, logs)
}
