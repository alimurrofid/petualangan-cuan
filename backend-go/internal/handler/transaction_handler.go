package handler

import (
	"cuan-backend/internal/entity"
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
	GetReport(c *fiber.Ctx) error
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
// @Description Get all transactions for the logged in user with pagination and filtering
// @Tags transactions
// @Accept json
// @Produce json
// @Param page query int false "Page Number" default(1)
// @Param limit query int false "Items per Page" default(10)
// @Param start_date query string false "Start Date"
// @Param end_date query string false "End Date"
// @Param wallet_id query int false "Wallet ID"
// @Param category_id query int false "Category ID"
// @Param search query string false "Search Term"
// @Param type query string false "Transaction Type"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/transactions [get]
func (h *transactionHandler) GetTransactions(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	// Parse Query Params
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	search := c.Query("search")
	transType := c.Query("type")
	
	walletID, _ := strconv.Atoi(c.Query("wallet_id", "0"))
	categoryID, _ := strconv.Atoi(c.Query("category_id", "0"))

	params := entity.TransactionFilterParams{
		Page:       page,
		Limit:      limit,
		StartDate:  startDate,
		EndDate:    endDate,
		Search:     search,
		Type:       transType,
		WalletID:   uint(walletID),
		CategoryID: uint(categoryID),
	}

	transactions, total, err := h.service.GetTransactions(userID, params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   transactions,
		"meta": fiber.Map{
			"total": total,
			"page":  page,
			"limit": limit,
		},
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
// @Description Get total income and expense per day for a specific date range with filters
// @Tags transactions
// @Accept json
// @Produce json
// @Param start_date query string true "Start Date (YYYY-MM-DD)"
// @Param end_date query string true "End Date (YYYY-MM-DD)"
// @Param wallet_id query int false "Wallet ID"
// @Param category_id query int false "Category ID"
// @Param search query string false "Search Term"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/transactions/calendar [get]
func (h *transactionHandler) GetCalendarData(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	walletIDStr := c.Query("wallet_id")
	categoryIDStr := c.Query("category_id")

	if startDate == "" || endDate == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "start_date and end_date are required"})
	}

	var walletID *uint
	if walletIDStr != "" && walletIDStr != "all" && walletIDStr != "0" {
		if id, err := strconv.ParseUint(walletIDStr, 10, 32); err == nil {
			uid := uint(id)
			walletID = &uid
		}
	}

	var categoryID *uint
	if categoryIDStr != "" && categoryIDStr != "all" && categoryIDStr != "0" {
		if id, err := strconv.ParseUint(categoryIDStr, 10, 32); err == nil {
			cid := uint(id)
			categoryID = &cid
		}
	}

	search := c.Query("search")

	summary, err := h.service.GetCalendarData(userID, startDate, endDate, walletID, categoryID, search)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   summary,
	})
}

// GetReport godoc
// @Summary Get category breakdown for report
// @Description Get comprehensive report of expenses/income by category
// @Tags transactions
// @Accept json
// @Produce json
// @Param start_date query string true "Start Date (YYYY-MM-DD)"
// @Param end_date query string true "End Date (YYYY-MM-DD)"
// @Param wallet_id query int false "Wallet ID"
// @Param type query string false "Transaction Type (income, expense, all)"
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Router /transactions/report [get]
func (h *transactionHandler) GetReport(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	walletIDStr := c.Query("wallet_id")
	filterType := c.Query("type")

	if startDate == "" || endDate == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "start_date and end_date are required",
		})
	}

	var walletID *uint
	if walletIDStr != "" && walletIDStr != "all" {
		id, err := strconv.ParseUint(walletIDStr, 10, 32)
		if err == nil {
			uid := uint(id)
			walletID = &uid
		}
	}

	var fType *string
	if filterType != "" && filterType != "all" {
		fType = &filterType
	}

	report, err := h.service.GetReport(userID, startDate, endDate, walletID, fType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   report,
	})
}

func (h *transactionHandler) WebhookReceiver(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Webhook receiver is ready!",
	})
}
