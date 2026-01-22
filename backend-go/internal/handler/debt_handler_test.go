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

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDebtService is a mock implementation of service.DebtService
type MockDebtService struct {
	mock.Mock
}

func (m *MockDebtService) CreateDebt(userID uint, input service.CreateDebtInput) (*entity.Debt, error) {
	args := m.Called(userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Debt), args.Error(1)
}

func (m *MockDebtService) GetDebts(userID uint, debtType string) ([]entity.Debt, error) {
	args := m.Called(userID, debtType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Debt), args.Error(1)
}

func (m *MockDebtService) GetDebt(id uint, userID uint) (*entity.Debt, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Debt), args.Error(1)
}

func (m *MockDebtService) PayDebt(id uint, userID uint, input service.PayDebtInput) (*entity.Debt, error) {
	args := m.Called(id, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Debt), args.Error(1)
}

func (m *MockDebtService) UpdateDebt(id uint, userID uint, input service.UpdateDebtInput) (*entity.Debt, error) {
	args := m.Called(id, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Debt), args.Error(1)
}

func (m *MockDebtService) DeleteDebt(id uint, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

func (m *MockDebtService) DeletePayment(id uint, userID uint) error {
    args := m.Called(id, userID)
    return args.Error(0)
}


// Handler Tests

func TestCreateDebt_Handler(t *testing.T) {
	mockService := new(MockDebtService)
	h := handler.NewDebtHandler(mockService)

	app := fiber.New()
	app.Post("/api/debts", mockAuthMiddleware(1), h.CreateDebt)

	input := service.CreateDebtInput{
		Name:     "Test Debt",
		Amount:   5000,
		WalletID: 1,
		Type:     "debt",
	}
	body, _ := json.Marshal(input)

	mockService.On("CreateDebt", uint(1), mock.MatchedBy(func(i service.CreateDebtInput) bool {
		return i.Name == "Test Debt"
	})).Return(&entity.Debt{ID: 1, Name: "Test Debt", Amount: 5000}, nil)

	req := httptest.NewRequest("POST", "/api/debts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetDebts_Handler(t *testing.T) {
	mockService := new(MockDebtService)
	h := handler.NewDebtHandler(mockService)

	app := fiber.New()
	app.Get("/api/debts", mockAuthMiddleware(1), h.GetDebts)

	mockService.On("GetDebts", uint(1), "").Return([]entity.Debt{{ID: 1, Name: "Debt 1"}}, nil)

	req := httptest.NewRequest("GET", "/api/debts", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetDebt_Handler(t *testing.T) {
	mockService := new(MockDebtService)
	h := handler.NewDebtHandler(mockService)

	app := fiber.New()
	app.Get("/api/debts/:id", mockAuthMiddleware(1), h.GetDebt)

	id := 1

	mockService.On("GetDebt", uint(id), uint(1)).Return(&entity.Debt{ID: uint(id)}, nil)

	req := httptest.NewRequest("GET", "/api/debts/"+strconv.Itoa(id), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetDebt_Handler_NotFound(t *testing.T) {
	mockService := new(MockDebtService)
	h := handler.NewDebtHandler(mockService)

	app := fiber.New()
	app.Get("/api/debts/:id", mockAuthMiddleware(1), h.GetDebt)

	id := 99

	mockService.On("GetDebt", uint(id), uint(1)).Return(nil, errors.New("not found"))

	req := httptest.NewRequest("GET", "/api/debts/"+strconv.Itoa(id), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestPayDebt_Handler(t *testing.T) {
	mockService := new(MockDebtService)
	h := handler.NewDebtHandler(mockService)

	app := fiber.New()
	app.Post("/api/debts/:id/pay", mockAuthMiddleware(1), h.PayDebt)

	id := 1
	input := service.PayDebtInput{
		WalletID: 1,
		Amount:   100,
	}
	body, _ := json.Marshal(input)

	mockService.On("PayDebt", uint(id), uint(1), mock.MatchedBy(func(i service.PayDebtInput) bool {
		return i.Amount == 100
	})).Return(&entity.Debt{ID: uint(id), Remaining: 900}, nil)

	req := httptest.NewRequest("POST", "/api/debts/"+strconv.Itoa(id)+"/pay", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestUpdateDebt_Handler(t *testing.T) {
	mockService := new(MockDebtService)
	h := handler.NewDebtHandler(mockService)

	app := fiber.New()
	app.Put("/api/debts/:id", mockAuthMiddleware(1), h.UpdateDebt)

	id := 1
	input := service.UpdateDebtInput{
		Name:     "Updated Debt",
		Amount:   6000,
		WalletID: 1,
	}
	body, _ := json.Marshal(input)

	mockService.On("UpdateDebt", uint(1), uint(1), mock.MatchedBy(func(i service.UpdateDebtInput) bool {
		return i.Name == "Updated Debt"
	})).Return(&entity.Debt{ID: uint(id), Name: "Updated Debt"}, nil)

	req := httptest.NewRequest("PUT", "/api/debts/"+strconv.Itoa(id), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestDeleteDebt_Handler(t *testing.T) {
	mockService := new(MockDebtService)
	h := handler.NewDebtHandler(mockService)

	app := fiber.New()
	app.Delete("/api/debts/:id", mockAuthMiddleware(1), h.DeleteDebt)

	id := 1

	mockService.On("DeleteDebt", uint(id), uint(1)).Return(nil)

	req := httptest.NewRequest("DELETE", "/api/debts/"+strconv.Itoa(id), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}
