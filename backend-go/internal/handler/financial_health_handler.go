package handler

import (
	"cuan-backend/internal/service"

	"github.com/gofiber/fiber/v2"
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
	// Assumes JWT middleware sets "userID" in Locals
	userID := c.Locals("userID").(uint)

	data, err := h.service.GetFinancialHealth(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": "error",
			"error":  err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   data,
	})
}
