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

type MockDashboardService struct {
	mock.Mock
}

func (m *MockDashboardService) GetDashboardData(userID uint) (*entity.DashboardData, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.DashboardData), args.Error(1)
}

func TestGetDashboard(t *testing.T) {
	mockService := new(MockDashboardService)
	dashboardHandler := handler.NewDashboardHandler(mockService)

	app := fiber.New()
	app.Get("/api/dashboard", mockAuthMiddleware(1), dashboardHandler.GetDashboard)

	mockData := &entity.DashboardData{
		TotalBalance:      1000000,
		TotalIncomeMonth:  500000,
		TotalExpenseMonth: 200000,
		Wallets:           []entity.Wallet{},
		RecentTransactions: []entity.Transaction{},
		MonthlyTrend:      []entity.MonthlyTrend{},
		ExpenseBreakdown:  []entity.CategoryBreakdown{},
	}
	mockService.On("GetDashboardData", uint(1)).Return(mockData, nil)

	req := httptest.NewRequest("GET", "/api/dashboard", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetDashboard_Error(t *testing.T) {
	mockService := new(MockDashboardService)
	dashboardHandler := handler.NewDashboardHandler(mockService)

	app := fiber.New()
	app.Get("/api/dashboard", mockAuthMiddleware(1), dashboardHandler.GetDashboard)

	mockService.On("GetDashboardData", uint(1)).Return(nil, errors.New("database error"))

	req := httptest.NewRequest("GET", "/api/dashboard", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	mockService.AssertExpectations(t)
}
