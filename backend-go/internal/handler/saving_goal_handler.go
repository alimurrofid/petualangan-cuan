package handler

import (
	"cuan-backend/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type SavingGoalHandler struct {
	service service.SavingGoalService
}

func NewSavingGoalHandler(service service.SavingGoalService) *SavingGoalHandler {
	return &SavingGoalHandler{service}
}

func (h *SavingGoalHandler) GetGoals(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	goals, err := h.service.GetGoals(userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"data": goals})
}

func (h *SavingGoalHandler) CreateGoal(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	var input service.CreateGoalInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	goal, err := h.service.CreateGoal(userID, input)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": goal})
}

func (h *SavingGoalHandler) AddContribution(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid goal ID"})
	}

	var input service.ContributionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	
	// Ensure date is set if not provided (though binding required it)
	if input.Date.IsZero() {
		input.Date = time.Now()
	}

	contribution, err := h.service.AddContribution(userID, uint(id), input)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": contribution})
}
