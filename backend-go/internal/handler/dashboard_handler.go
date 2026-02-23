package handler

import (
	"cuan-backend/internal/service"
	"cuan-backend/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
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
// @Router /api/dashboard [get]
func (h *dashboardHandler) GetDashboard(c *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		reqID, _ := c.Locals("requestid").(string)
		log.Warn().Str("request_id", reqID).Err(err).Msg("Failed to get user ID from context")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	data, err := h.service.GetDashboardData(userID)
	if err != nil {
		reqID, _ := c.Locals("requestid").(string)
		log.Error().Str("request_id", reqID).Err(err).Msg("Internal server error")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   data,
	})
}
