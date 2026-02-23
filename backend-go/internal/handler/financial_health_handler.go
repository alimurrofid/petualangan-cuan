package handler

import (
	"cuan-backend/internal/service"
	"cuan-backend/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type FinancialHealthHandler struct {
	service service.FinancialHealthService
}

func NewFinancialHealthHandler(service service.FinancialHealthService) *FinancialHealthHandler {
	return &FinancialHealthHandler{service: service}
}

// GetFinancialHealth godoc
// @Summary Get financial health analysis
// @Description Get comprehensive financial health check including savings rate, liquidity, and debt ratio
// @Tags financial_health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/financial-health [get]
func (h *FinancialHealthHandler) GetFinancialHealth(c *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		reqID, _ := c.Locals("requestid").(string)
		log.Warn().Str("request_id", reqID).Err(err).Msg("Failed to get user ID from context")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status": "error",
			"error":  "Unauthorized",
		})
	}

	data, err := h.service.GetFinancialHealth(userID)
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
