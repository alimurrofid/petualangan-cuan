package handler

import (
	"cuan-backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type CategoryHandler struct {
	service service.CategoryService
}

func NewCategoryHandler(service service.CategoryService) *CategoryHandler {
	return &CategoryHandler{service}
}

// CreateCategory godoc
// @Summary Create a new category
// @Description Create a new category for transactions
// @Tags categories
// @Accept json
// @Produce json
// @Param category body service.CreateCategoryInput true "Category Input"
// @Success 201 {object} entity.Category
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/categories [post]
func (h *CategoryHandler) CreateCategory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var input service.CreateCategoryInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	category, err := h.service.CreateCategory(userID, input)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(category)
}

// GetCategories godoc
// @Summary Get all categories
// @Description Get all categories for the logged in user
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {array} entity.Category
// @Failure 500 {object} map[string]interface{}
// @Router /api/categories [get]
func (h *CategoryHandler) GetCategories(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	categories, err := h.service.GetCategories(userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(categories)
}

// GetCategory godoc
// @Summary Get a category by ID
// @Description Get a specific category
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} entity.Category
// @Failure 404 {object} map[string]interface{}
// @Router /api/categories/{id} [get]
func (h *CategoryHandler) GetCategory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	id, _ := strconv.Atoi(c.Params("id"))

	category, err := h.service.GetCategory(uint(id), userID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Category not found"})
	}

	return c.JSON(category)
}

// UpdateCategory godoc
// @Summary Update a category
// @Description Update category details
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param category body service.UpdateCategoryInput true "Update Category Input"
// @Success 200 {object} entity.Category
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	id, _ := strconv.Atoi(c.Params("id"))

	var input service.UpdateCategoryInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	category, err := h.service.UpdateCategory(uint(id), userID, input)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(category)
}

// DeleteCategory godoc
// @Summary Delete a category
// @Description Delete a category by ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	id, _ := strconv.Atoi(c.Params("id"))

	err := h.service.DeleteCategory(uint(id), userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Category deleted"})
}
