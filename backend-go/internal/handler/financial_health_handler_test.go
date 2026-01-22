package handler_test

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/handler"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockFinancialHealthService
type MockFinancialHealthService struct {
	mock.Mock
}

func (m *MockFinancialHealthService) GetFinancialHealth(userID uint) (entity.FinancialHealthResponse, error) {
	args := m.Called(userID)
	return args.Get(0).(entity.FinancialHealthResponse), args.Error(1)
}

func TestGetFinancialHealth_Handler(t *testing.T) {
	mockService := new(MockFinancialHealthService)
	h := handler.NewFinancialHealthHandler(mockService)

	app := fiber.New()
	app.Get("/api/financial-health", mockAuthMiddleware(1), h.GetFinancialHealth)

	expectedResponse := entity.FinancialHealthResponse{
		OverallStatus: entity.StatusHealthy,
		OverallScore:  100,
	}

	mockService.On("GetFinancialHealth", uint(1)).Return(expectedResponse, nil)

	req := httptest.NewRequest("GET", "/api/financial-health", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetFinancialHealth_Handler_Error(t *testing.T) {
	mockService := new(MockFinancialHealthService)
	h := handler.NewFinancialHealthHandler(mockService)

	app := fiber.New()
	app.Get("/api/financial-health", mockAuthMiddleware(1), h.GetFinancialHealth)

	mockService.On("GetFinancialHealth", uint(1)).Return(entity.FinancialHealthResponse{}, errors.New("service error"))

	req := httptest.NewRequest("GET", "/api/financial-health", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	mockService.AssertExpectations(t)
}
