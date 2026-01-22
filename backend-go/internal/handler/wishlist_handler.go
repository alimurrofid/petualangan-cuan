package handler

import (
	"cuan-backend/internal/service"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type WishlistHandler struct {
	wishlistService service.WishlistService
}

func NewWishlistHandler(wishlistService service.WishlistService) *WishlistHandler {
	return &WishlistHandler{wishlistService}
}

// Create godoc
// @Summary Create a new wishlist item
// @Description Create a new prospective purchase item
// @Tags wishlist
// @Accept json
// @Produce json
// @Param item body service.StoreWishlistRequest true "Wishlist Item"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/wishlists [post]
func (h *WishlistHandler) Create(c *fiber.Ctx) error {
	var req service.StoreWishlistRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	
	// User ID is set directly in middleware
	userID := c.Locals("userID").(uint)

	if err := h.wishlistService.Create(userID, &req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Wishlist item created successfully"})
}

// FindAll godoc
// @Summary Get all wishlist items
// @Description Get list of all wishlist items for the user
// @Tags wishlist
// @Accept json
// @Produce json
// @Success 200 {object} []entity.WishlistItem
// @Router /api/wishlists [get]
func (h *WishlistHandler) FindAll(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	items, err := h.wishlistService.FindAll(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(items)
}

// FindByID godoc
// @Summary Get a wishlist item
// @Description Get details of a specific wishlist item
// @Tags wishlist
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Success 200 {object} entity.WishlistItem
// @Failure 404 {object} map[string]interface{}
// @Router /api/wishlists/{id} [get]
func (h *WishlistHandler) FindByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	userID := c.Locals("userID").(uint)

	item, err := h.wishlistService.FindByID(uint(id), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Item not found"})
	}

	return c.JSON(item)
}

// Update godoc
// @Summary Update a wishlist item
// @Description Update details of a wishlist item
// @Tags wishlist
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Param item body service.StoreWishlistRequest true "Wishlist Item"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/wishlists/{id} [put]
func (h *WishlistHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var req service.StoreWishlistRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	userID := c.Locals("userID").(uint)

	if err := h.wishlistService.Update(uint(id), userID, &req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Wishlist item updated successfully"})
}

// Delete godoc
// @Summary Delete a wishlist item
// @Description Delete a wishlist item
// @Tags wishlist
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/wishlists/{id} [delete]
func (h *WishlistHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	userID := c.Locals("userID").(uint)

	if err := h.wishlistService.Delete(uint(id), userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Wishlist item deleted successfully"})
}

// MarkAsBought godoc
// @Summary Mark a wishlist item as bought
// @Description Mark item as purchased
// @Tags wishlist
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/wishlists/{id}/bought [patch]
func (h *WishlistHandler) MarkAsBought(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	userID := c.Locals("userID").(uint)

	if err := h.wishlistService.MarkAsBought(uint(id), userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Wishlist item marked as bought"})
}
