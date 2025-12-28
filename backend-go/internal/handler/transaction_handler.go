package handler

import (
	"cuan-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type TransactionHandler interface {
	GetTransactions(c *fiber.Ctx) error
	WebhookReceiver(c *fiber.Ctx) error
}

type transactionHandler struct {
	service service.TransactionService
}

func NewTransactionHandler(service service.TransactionService) TransactionHandler {
	return &transactionHandler{service}
}

// GetTransactions godoc
// @Summary Get all transactions
// @Description Get a list of all transactions ordered by creation date
// @Tags transactions
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/transactions [get]
func (h *transactionHandler) GetTransactions(c *fiber.Ctx) error {
	transactions, err := h.service.GetTransactions()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   transactions,
	})
}

// WebhookReceiver godoc
// @Summary Webhook Receiver
// @Description Receive incoming webhooks
// @Tags webhook
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/webhook [post]
func (h *transactionHandler) WebhookReceiver(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Webhook receiver is ready!",
	})
}
