package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"cuan-backend/internal/entity"
	"cuan-backend/internal/handler"
	"cuan-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockWalletService struct {
	mock.Mock
}

func (m *MockWalletService) CreateWallet(input service.CreateWalletInput) (*entity.Wallet, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Wallet), args.Error(1)
}

func (m *MockWalletService) GetWalletByID(id uint, userID uint) (*entity.Wallet, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Wallet), args.Error(1)
}

func (m *MockWalletService) GetUserWallets(userID uint) ([]entity.Wallet, error) {
	args := m.Called(userID)
	return args.Get(0).([]entity.Wallet), args.Error(1)
}

func (m *MockWalletService) UpdateWallet(id uint, userID uint, input service.UpdateWalletInput) (*entity.Wallet, error) {
	args := m.Called(id, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Wallet), args.Error(1)
}

func (m *MockWalletService) DeleteWallet(id uint, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

func mockAuthMiddleware(userID uint) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return c.Next()
	}
}

func TestCreateWalletHandler(t *testing.T) {
	mockService := new(MockWalletService)
	walletHandler := handler.NewWalletHandler(mockService)

	app := fiber.New()
	app.Post("/api/wallets", mockAuthMiddleware(1), walletHandler.CreateWallet)

	input := service.CreateWalletInput{
		Name:    "Test Wallet",
		Type:    "Bank",
		Balance: 100000,
		Icon:    "Landmark",
	}
	body, _ := json.Marshal(input)

	expectedInput := input
	expectedInput.UserID = 1
	
	mockWallet := &entity.Wallet{ID: 1, UserID: 1, Name: "Test Wallet"}
	mockService.On("CreateWallet", expectedInput).Return(mockWallet, nil)

	req := httptest.NewRequest("POST", "/api/wallets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetWalletsHandler(t *testing.T) {
	mockService := new(MockWalletService)
	walletHandler := handler.NewWalletHandler(mockService)

	app := fiber.New()
	app.Get("/api/wallets", mockAuthMiddleware(1), walletHandler.GetWallets)

	wallets := []entity.Wallet{{ID: 1, UserID: 1, Name: "Wallet 1"}}
	mockService.On("GetUserWallets", uint(1)).Return(wallets, nil)

	req := httptest.NewRequest("GET", "/api/wallets", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetWalletHandler_NotFound(t *testing.T) {
	mockService := new(MockWalletService)
	walletHandler := handler.NewWalletHandler(mockService)

	app := fiber.New()
	app.Get("/api/wallets/:id", mockAuthMiddleware(1), walletHandler.GetWallet)

	mockService.On("GetWalletByID", uint(999), uint(1)).Return(nil, errors.New("record not found"))

	req := httptest.NewRequest("GET", "/api/wallets/999", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	mockService.AssertExpectations(t)
}
