package handler

import (
	"cuan-backend/internal/service"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type WalletHandler interface {
	CreateWallet(c *fiber.Ctx) error
	GetWallets(c *fiber.Ctx) error
	GetWallet(c *fiber.Ctx) error
	UpdateWallet(c *fiber.Ctx) error
	DeleteWallet(c *fiber.Ctx) error
}

type walletHandler struct {
	walletService service.WalletService
}

func NewWalletHandler(walletService service.WalletService) WalletHandler {
	return &walletHandler{walletService}
}

// CreateWallet godoc
// @Summary Create a new wallet
// @Description Create a new wallet for the authenticated user
// @Tags wallets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreateWalletInput true "Create Wallet Request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/wallets [post]
func (h *walletHandler) CreateWallet(c *fiber.Ctx) error {
	var input service.CreateWalletInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userID := c.Locals("user_id").(uint)

	input.UserID = userID

	wallet, err := h.walletService.CreateWallet(input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(wallet)
}

// GetWallets godoc
// @Summary Get all wallets
// @Description Get all wallets for the authenticated user
// @Tags wallets
// @Produce json
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/wallets [get]
func (h *walletHandler) GetWallets(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	wallets, err := h.walletService.GetUserWallets(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(wallets)
}

// GetWallet godoc
// @Summary Get a wallet by ID
// @Description Get a specific wallet by ID (must belong to user)
// @Tags wallets
// @Produce json
// @Security BearerAuth
// @Param id path int true "Wallet ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/wallets/{id} [get]
func (h *walletHandler) GetWallet(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	userID := c.Locals("user_id").(uint)

	wallet, err := h.walletService.GetWalletByID(uint(id), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(wallet)
}

// UpdateWallet godoc
// @Summary Update a wallet
// @Description Update a specific wallet details
// @Tags wallets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Wallet ID"
// @Param request body service.UpdateWalletInput true "Update Wallet Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/wallets/{id} [put]
func (h *walletHandler) UpdateWallet(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var input service.UpdateWalletInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userID := c.Locals("user_id").(uint)

	wallet, err := h.walletService.UpdateWallet(uint(id), userID, input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(wallet)
}

// DeleteWallet godoc
// @Summary Delete a wallet
// @Description Delete a specific wallet
// @Tags wallets
// @Security BearerAuth
// @Param id path int true "Wallet ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/wallets/{id} [delete]
func (h *walletHandler) DeleteWallet(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	userID := c.Locals("user_id").(uint)

	err = h.walletService.DeleteWallet(uint(id), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Wallet deleted successfully"})
}
