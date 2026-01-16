package handler

import (
	"cuan-backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type DebtHandler interface {
	CreateDebt(c *fiber.Ctx) error
	GetDebts(c *fiber.Ctx) error
	GetDebt(c *fiber.Ctx) error
	PayDebt(c *fiber.Ctx) error
	UpdateDebt(c *fiber.Ctx) error
	DeleteDebt(c *fiber.Ctx) error
	DeletePayment(c *fiber.Ctx) error
}

type debtHandler struct {
	service service.DebtService
}

func NewDebtHandler(service service.DebtService) DebtHandler {
	return &debtHandler{service}
}

// CreateDebt godoc
// @Summary Create a new debt or receivable
// @Description Create a new debt record and automatically create associated transaction
// @Tags debts
// @Accept json
// @Produce json
// @Param debt body service.CreateDebtInput true "Debt Input"
// @Success 201 {object} entity.Debt
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/debts [post]
func (h *debtHandler) CreateDebt(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	var input service.CreateDebtInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	debt, err := h.service.CreateDebt(userID, input)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(debt)
}

// GetDebts godoc
// @Summary Get all debts/receivables
// @Description Get list of debts filtered by type
// @Tags debts
// @Accept json
// @Produce json
// @Param type query string false "Type (debt/receivable)"
// @Success 200 {object} []entity.Debt
// @Router /api/debts [get]
func (h *debtHandler) GetDebts(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	debtType := c.Query("type")

	debts, err := h.service.GetDebts(userID, debtType)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": debts})
}

// GetDebt godoc
// @Summary Get a single debt
// @Description Get details of a specific debt
// @Tags debts
// @Accept json
// @Produce json
// @Param id path int true "Debt ID"
// @Success 200 {object} entity.Debt
// @Failure 404 {object} map[string]interface{}
// @Router /api/debts/{id} [get]
func (h *debtHandler) GetDebt(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	id, _ := strconv.Atoi(c.Params("id"))

	debt, err := h.service.GetDebt(uint(id), userID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Debt not found"})
	}

	return c.JSON(debt)
}

// PayDebt godoc
// @Summary Pay a debt installment
// @Description Record a payment for a debt and update remaining amount
// @Tags debts
// @Accept json
// @Produce json
// @Param id path int true "Debt ID"
// @Param payment body service.PayDebtInput true "Payment Input"
// @Success 200 {object} entity.Debt
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/debts/{id}/pay [post]
func (h *debtHandler) PayDebt(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	id, _ := strconv.Atoi(c.Params("id"))

	var input service.PayDebtInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	debt, err := h.service.PayDebt(uint(id), userID, input)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(debt)
}


// UpdateDebt godoc
// @Summary Update a debt record
// @Description Update name, description, and due date of a debt
// @Tags debts
// @Accept json
// @Produce json
// @Param id path int true "Debt ID"
// @Param debt body service.UpdateDebtInput true "Update Input"
// @Success 200 {object} entity.Debt
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/debts/{id} [put]
func (h *debtHandler) UpdateDebt(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	id, _ := strconv.Atoi(c.Params("id"))

	var input service.UpdateDebtInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	debt, err := h.service.UpdateDebt(uint(id), userID, input)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(debt)
}

// DeleteDebt godoc
// @Summary Delete a debt record
// @Description Delete a debt record (CAUTION: does not revert transactions)
// @Tags debts
// @Accept json
// @Produce json
// @Param id path int true "Debt ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/debts/{id} [delete]
func (h *debtHandler) DeleteDebt(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	id, _ := strconv.Atoi(c.Params("id"))

	if err := h.service.DeleteDebt(uint(id), userID); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Debt deleted"})
}

// DeletePayment godoc
// @Summary Delete a debt payment
// @Description Delete a debt payment and revert balance/debt remaining
// @Tags debts
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/debts/payments/{id} [delete]
func (h *debtHandler) DeletePayment(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	id, _ := strconv.Atoi(c.Params("id"))

	if err := h.service.DeletePayment(uint(id), userID); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Payment deleted and balance reverted"})
}
