package handler

import (
	"cuan-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type DashboardHandler interface {
	GetDashboard(c *fiber.Ctx) error
}

type dashboardHandler struct {
	service service.DashboardService
}

func NewDashboardHandler(service service.DashboardService) DashboardHandler {
	return &dashboardHandler{service}
}

// GetDashboard godoc
// @Summary Get dashboard data
// @Description Get total balance, monthly summary, recent transactions, trend, and breakdown
// @Tags dashboard
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Router /dashboard [get]
func (h *dashboardHandler) GetDashboard(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	data, err := h.service.GetDashboardData(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   data,
	})
}
