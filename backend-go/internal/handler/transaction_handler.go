package handler

import (
	"cuan-backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type TransactionHandler interface {
	CreateTransaction(c *fiber.Ctx) error
	GetTransactions(c *fiber.Ctx) error
	DeleteTransaction(c *fiber.Ctx) error
	TransferTransaction(c *fiber.Ctx) error
	GetCalendarData(c *fiber.Ctx) error
	WebhookReceiver(c *fiber.Ctx) error
}

type transactionHandler struct {
	service service.TransactionService
}

func NewTransactionHandler(service service.TransactionService) TransactionHandler {
	return &transactionHandler{service}
}

// CreateTransaction godoc
// @Summary Create a new transaction
// @Description Create a new income or expense transaction
// @Tags transactions
// @Accept json
// @Produce json
// @Param transaction body service.CreateTransactionInput true "Transaction Input"
// @Success 201 {object} entity.Transaction
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/transactions [post]
func (h *transactionHandler) CreateTransaction(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var input service.CreateTransactionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	transaction, err := h.service.CreateTransaction(userID, input)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(transaction)
}

// GetTransactions godoc
// @Summary Get all transactions
// @Description Get all transactions for the logged in user
// @Tags transactions
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/transactions [get]
func (h *transactionHandler) GetTransactions(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	transactions, err := h.service.GetTransactions(userID)
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

// DeleteTransaction godoc
// @Summary Delete a transaction
// @Description Delete a transaction by ID and revert balance
// @Tags transactions
// @Accept json
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/transactions/{id} [delete]
func (h *transactionHandler) DeleteTransaction(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	id, _ := strconv.Atoi(c.Params("id"))

	err := h.service.DeleteTransaction(uint(id), userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Transaction deleted"})
}

// TransferTransaction godoc
// @Summary Transfer money between wallets
// @Description Create a transfer comprising an expense and an income
// @Tags transactions
// @Accept json
// @Produce json
// @Param transfer body service.TransferTransactionInput true "Transfer Input"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/transactions/transfer [post]
func (h *transactionHandler) TransferTransaction(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var input service.TransferTransactionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	if err := h.service.TransferTransaction(userID, input); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Transfer successful"})
}

// GetCalendarData godoc
// @Summary Get aggregated calendar data
// @Description Get total income and expense per day for a specific date range
// @Tags transactions
// @Accept json
// @Produce json
// @Param start_date query string true "Start Date (YYYY-MM-DD)"
// @Param end_date query string true "End Date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/transactions/calendar [get]
func (h *transactionHandler) GetCalendarData(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "start_date and end_date are required"})
	}

	summary, err := h.service.GetCalendarData(userID, startDate, endDate)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   summary,
	})
}

func (h *transactionHandler) WebhookReceiver(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Webhook receiver is ready!",
	})
}
