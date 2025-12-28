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

func (h *CategoryHandler) GetCategories(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	categories, err := h.service.GetCategories(userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(categories)
}

func (h *CategoryHandler) GetCategory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	id, _ := strconv.Atoi(c.Params("id"))

	category, err := h.service.GetCategory(uint(id), userID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Category not found"})
	}

	return c.JSON(category)
}

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

func (h *CategoryHandler) DeleteCategory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	id, _ := strconv.Atoi(c.Params("id"))

	err := h.service.DeleteCategory(uint(id), userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Category deleted"})
}
