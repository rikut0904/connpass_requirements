package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"connpass-requirement/internal/repository"
)

// StatusHandler はスケジューラ状態取得API。
type StatusHandler struct {
	logs *repository.LogRepository
}

func NewStatusHandler(logs *repository.LogRepository) *StatusHandler {
	return &StatusHandler{logs: logs}
}

// RegisterStatusRoutes はステータス関連ルートを登録。
func RegisterStatusRoutes(g *echo.Group, handler *StatusHandler) {
	g.GET("/status", handler.GetStatus)
}

func (h *StatusHandler) GetStatus(c echo.Context) error {
	status, err := h.logs.GetSchedulerStatus(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch status")
	}
	if status == nil {
		return c.JSON(http.StatusOK, map[string]any{"status": "unknown"})
	}

	return c.JSON(http.StatusOK, status)
}
