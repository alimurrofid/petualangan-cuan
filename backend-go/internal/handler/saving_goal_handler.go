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

// GetGoals godoc
// @Summary Get all saving goals
// @Description Get list of saving goals for the authenticated user
// @Tags saving_goals
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/saving-goals [get]
func (h *SavingGoalHandler) GetGoals(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	goals, err := h.service.GetGoals(userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"data": goals})
}

// CreateGoal godoc
// @Summary Create a new saving goal
// @Description Create a new saving goal record
// @Tags saving_goals
// @Accept json
// @Produce json
// @Param goal body service.CreateGoalInput true "Goal Input"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/saving-goals [post]
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

// AddContribution godoc
// @Summary Add contribution to a saving goal
// @Description Add money to a saving goal
// @Tags saving_goals
// @Accept json
// @Produce json
// @Param id path int true "Goal ID"
// @Param contribution body service.ContributionInput true "Contribution Input"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/saving-goals/{id}/contributions [post]
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

// UpdateGoal godoc
// @Summary Update a saving goal
// @Description Update details of a saving goal
// @Tags saving_goals
// @Accept json
// @Produce json
// @Param id path int true "Goal ID"
// @Param goal body service.CreateGoalInput true "Goal Input"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/saving-goals/{id} [put]
func (h *SavingGoalHandler) UpdateGoal(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid goal ID"})
	}

	var input service.CreateGoalInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	goal, err := h.service.UpdateGoal(userID, uint(id), input)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"data": goal})
}

// DeleteGoal godoc
// @Summary Delete a saving goal
// @Description Delete a saving goal
// @Tags saving_goals
// @Accept json
// @Produce json
// @Param id path int true "Goal ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/saving-goals/{id} [delete]
func (h *SavingGoalHandler) DeleteGoal(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid goal ID"})
	}

	if err := h.service.DeleteGoal(userID, uint(id)); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "Saving goal deleted successfully"})
}

// DeleteContribution godoc
// @Summary Delete a saving contribution
// @Description Delete a saving contribution
// @Tags saving_goals
// @Accept json
// @Produce json
// @Param id path int true "Goal ID"
// @Param contribution_id path int true "Contribution ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/saving-goals/{id}/contributions/{contribution_id} [delete]
func (h *SavingGoalHandler) DeleteContribution(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	// Goal ID not strictly needed if we look up by contribution ID but good for URL structure consistency
	// id, _ := strconv.Atoi(c.Params("id")) 
	contributionID, err := strconv.Atoi(c.Params("contribution_id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid contribution ID"})
	}

	if err := h.service.DeleteContribution(userID, uint(contributionID)); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "Contribution deleted successfully"})
}

// FinishGoal godoc
// @Summary Finish and cash out a saving goal
// @Description Mark a saving goal as finished and release funds to available balance
// @Tags saving_goals
// @Accept json
// @Produce json
// @Param id path int true "Goal ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/saving-goals/{id}/finish [put]
func (h *SavingGoalHandler) FinishGoal(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid goal ID"})
	}

	if err := h.service.FinishGoal(userID, uint(id)); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "Saving goal finished and funds released successfully"})
}
