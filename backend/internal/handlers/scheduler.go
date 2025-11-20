package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"connpass-requirement/internal/services"
)

type SchedulerHandler struct {
	scheduler *services.SchedulerService
}

func NewSchedulerHandler(scheduler *services.SchedulerService) *SchedulerHandler {
	return &SchedulerHandler{
		scheduler: scheduler,
	}
}

func (h *SchedulerHandler) RunNow(c echo.Context) error {
	ctx := c.Request().Context()

	if err := h.scheduler.Run(ctx); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"message": "スケジューラーの実行に失敗しました",
			"error":   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "スケジューラーを実行しました",
	})
}

func RegisterSchedulerRoutes(g *echo.Group, h *SchedulerHandler) {
	g.POST("/scheduler/run", h.RunNow)
}
