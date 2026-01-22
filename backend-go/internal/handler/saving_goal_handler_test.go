package handler_test

import (
	"bytes"
	"cuan-backend/internal/entity"
	"cuan-backend/internal/handler"
	"cuan-backend/internal/service"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSavingGoalService is a mock implementation of service.SavingGoalService
type MockSavingGoalService struct {
	mock.Mock
}

func (m *MockSavingGoalService) CreateGoal(userID uint, input service.CreateGoalInput) (*entity.SavingGoal, error) {
	args := m.Called(userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.SavingGoal), args.Error(1)
}

func (m *MockSavingGoalService) GetGoals(userID uint) ([]entity.SavingGoal, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.SavingGoal), args.Error(1)
}

func (m *MockSavingGoalService) AddContribution(userID uint, goalID uint, input service.ContributionInput) (*entity.SavingContribution, error) {
	args := m.Called(userID, goalID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.SavingContribution), args.Error(1)
}

func (m *MockSavingGoalService) UpdateGoal(userID uint, goalID uint, input service.CreateGoalInput) (*entity.SavingGoal, error) {
	args := m.Called(userID, goalID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.SavingGoal), args.Error(1)
}

func (m *MockSavingGoalService) DeleteGoal(userID uint, goalID uint) error {
	args := m.Called(userID, goalID)
	return args.Error(0)
}

func (m *MockSavingGoalService) DeleteContribution(userID uint, contributionID uint) error {
	args := m.Called(userID, contributionID)
	return args.Error(0)
}

// Handler Test Suite

func TestGetGoals_Handler(t *testing.T) {
	mockService := new(MockSavingGoalService)
	h := handler.NewSavingGoalHandler(mockService)

	app := fiber.New()
	app.Get("/api/saving-goals", mockAuthMiddleware(1), h.GetGoals)

	expectedGoals := []entity.SavingGoal{
		{UserID: 1, Name: "Goal 1", TargetAmount: 1000},
	}

	mockService.On("GetGoals", uint(1)).Return(expectedGoals, nil)

	req := httptest.NewRequest("GET", "/api/saving-goals", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetGoals_Handler_Error(t *testing.T) {
	mockService := new(MockSavingGoalService)
	h := handler.NewSavingGoalHandler(mockService)

	app := fiber.New()
	app.Get("/api/saving-goals", mockAuthMiddleware(1), h.GetGoals)

	mockService.On("GetGoals", uint(1)).Return(nil, errors.New("service error"))

	req := httptest.NewRequest("GET", "/api/saving-goals", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestCreateGoal_Handler(t *testing.T) {
	mockService := new(MockSavingGoalService)
	h := handler.NewSavingGoalHandler(mockService)

	app := fiber.New()
	app.Post("/api/saving-goals", mockAuthMiddleware(1), h.CreateGoal)

	input := service.CreateGoalInput{
		Name:         "New Goal",
		TargetAmount: 5000,
		CategoryID:   1,
	}
	body, _ := json.Marshal(input)

	expectedGoal := &entity.SavingGoal{
		UserID: 1, Name: "New Goal", TargetAmount: 5000, CategoryID: 1,
	}

	mockService.On("CreateGoal", uint(1), input).Return(expectedGoal, nil)

	req := httptest.NewRequest("POST", "/api/saving-goals", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestUpdateGoal_Handler(t *testing.T) {
	mockService := new(MockSavingGoalService)
	h := handler.NewSavingGoalHandler(mockService)

	app := fiber.New()
	app.Put("/api/saving-goals/:id", mockAuthMiddleware(1), h.UpdateGoal)

	goalID := 1
	input := service.CreateGoalInput{
		Name:         "Updated Goal",
		TargetAmount: 6000,
		CategoryID:   2,
	}
	body, _ := json.Marshal(input)

	expectedGoal := &entity.SavingGoal{
		ID: uint(goalID), UserID: 1, Name: "Updated Goal", TargetAmount: 6000, CategoryID: 2,
	}

	mockService.On("UpdateGoal", uint(1), uint(goalID), input).Return(expectedGoal, nil)

	req := httptest.NewRequest("PUT", "/api/saving-goals/"+strconv.Itoa(goalID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestDeleteGoal_Handler(t *testing.T) {
	mockService := new(MockSavingGoalService)
	h := handler.NewSavingGoalHandler(mockService)

	app := fiber.New()
	app.Delete("/api/saving-goals/:id", mockAuthMiddleware(1), h.DeleteGoal)

	goalID := 1

	mockService.On("DeleteGoal", uint(1), uint(goalID)).Return(nil)

	req := httptest.NewRequest("DELETE", "/api/saving-goals/"+strconv.Itoa(goalID), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestAddContribution_Handler(t *testing.T) {
	mockService := new(MockSavingGoalService)
	h := handler.NewSavingGoalHandler(mockService)

	app := fiber.New()
	app.Post("/api/saving-goals/:id/contributions", mockAuthMiddleware(1), h.AddContribution)

	goalID := 1
	input := service.ContributionInput{
		WalletID: 1,
		Amount:   100,
		Date:     time.Now(),
	}
	// Need to fix time comparison in mock, usually issues with monotonic clock or parsing
	// Easier to use Matcher or just ensure unmarshal works similarly
	// For this test, let's construct body and matcher carefully
	
	body, _ := json.Marshal(input)

	expectedContribution := &entity.SavingContribution{
		GoalID: uint(goalID), WalletID: 1, Amount: 100,
	}

	// Use MatchedBy to ignore slight time differences during JSON Marshal/Unmarshal
	mockService.On("AddContribution", uint(1), uint(goalID), mock.MatchedBy(func(i service.ContributionInput) bool {
		return i.WalletID == input.WalletID && i.Amount == input.Amount
	})).Return(expectedContribution, nil)

	req := httptest.NewRequest("POST", "/api/saving-goals/"+strconv.Itoa(goalID)+"/contributions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	mockService.AssertExpectations(t)
}
