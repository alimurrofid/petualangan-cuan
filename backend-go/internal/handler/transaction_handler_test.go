package handler_test

import (
	"bytes"
	"cuan-backend/internal/entity"
	"cuan-backend/internal/handler"
	"cuan-backend/internal/service"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTransactionService struct {
	mock.Mock
}

func (m *MockTransactionService) CreateTransaction(userID uint, input service.CreateTransactionInput) (*entity.Transaction, error) {
	args := m.Called(userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Transaction), args.Error(1)
}

func (m *MockTransactionService) GetTransactions(userID uint) ([]entity.Transaction, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Transaction), args.Error(1)
}

func (m *MockTransactionService) DeleteTransaction(id uint, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

func (m *MockTransactionService) TransferTransaction(userID uint, input service.TransferTransactionInput) error {
	args := m.Called(userID, input)
	return args.Error(0)
}

func (m *MockTransactionService) GetCalendarData(userID uint, startDate, endDate string) ([]entity.TransactionSummary, error) {
	args := m.Called(userID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.TransactionSummary), args.Error(1)
}

func (m *MockTransactionService) GetReport(userID uint, startDate, endDate string, walletID *uint, filterType *string) ([]entity.CategoryBreakdown, error) {
	args := m.Called(userID, startDate, endDate, walletID, filterType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.CategoryBreakdown), args.Error(1)
}

func TestCreateTransaction(t *testing.T) {
	mockService := new(MockTransactionService)
	transactionHandler := handler.NewTransactionHandler(mockService)

	app := fiber.New()
	app.Post("/api/transactions", mockAuthMiddleware(1), transactionHandler.CreateTransaction)

	now := time.Now()
	input := service.CreateTransactionInput{
		WalletID:   1,
		CategoryID: 1,
		Amount:     10000,
		Type:       "expense",
		Date:       now,
	}
	body, _ := json.Marshal(input)
	
	mockTransaction := &entity.Transaction{
		ID: 1, UserID: 1, WalletID: 1, CategoryID: 1, Amount: 10000, Type: "expense", Date: now,
	}
	
	mockService.On("CreateTransaction", uint(1), mock.MatchedBy(func(i service.CreateTransactionInput) bool {
		return i.WalletID == input.WalletID && i.Amount == input.Amount
	})).Return(mockTransaction, nil)

	req := httptest.NewRequest("POST", "/api/transactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetTransactions(t *testing.T) {
	mockService := new(MockTransactionService)
	transactionHandler := handler.NewTransactionHandler(mockService)

	app := fiber.New()
	app.Get("/api/transactions", mockAuthMiddleware(1), transactionHandler.GetTransactions)

	mockTransactions := []entity.Transaction{
		{ID: 1, UserID: 1, Amount: 10000, Type: "expense"},
	}
	mockService.On("GetTransactions", uint(1)).Return(mockTransactions, nil)

	req := httptest.NewRequest("GET", "/api/transactions", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestDeleteTransaction(t *testing.T) {
	mockService := new(MockTransactionService)
	transactionHandler := handler.NewTransactionHandler(mockService)

	app := fiber.New()
	app.Delete("/api/transactions/:id", mockAuthMiddleware(1), transactionHandler.DeleteTransaction)

	mockService.On("DeleteTransaction", uint(1), uint(1)).Return(nil)

	req := httptest.NewRequest("DELETE", "/api/transactions/1", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestTransferTransaction(t *testing.T) {
	mockService := new(MockTransactionService)
	transactionHandler := handler.NewTransactionHandler(mockService)

	app := fiber.New()
	app.Post("/api/transactions/transfer", mockAuthMiddleware(1), transactionHandler.TransferTransaction)

	input := service.TransferTransactionInput{
		FromWalletID: 1,
		ToWalletID:   2,
		Amount:       50000,
		Date:         time.Now(),
	}
	body, _ := json.Marshal(input)

	mockService.On("TransferTransaction", uint(1), mock.MatchedBy(func(i service.TransferTransactionInput) bool {
		return i.Amount == input.Amount
	})).Return(nil)

	req := httptest.NewRequest("POST", "/api/transactions/transfer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetCalendarData(t *testing.T) {
	mockService := new(MockTransactionService)
	transactionHandler := handler.NewTransactionHandler(mockService)

	app := fiber.New()
	app.Get("/api/transactions/calendar", mockAuthMiddleware(1), transactionHandler.GetCalendarData)

	startDate := "2023-01-01"
	endDate := "2023-01-31"
	
	mockData := []entity.TransactionSummary{}
	mockService.On("GetCalendarData", uint(1), startDate, endDate).Return(mockData, nil)

	req := httptest.NewRequest("GET", "/api/transactions/calendar?start_date=2023-01-01&end_date=2023-01-31", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetReport(t *testing.T) {
	mockService := new(MockTransactionService)
	transactionHandler := handler.NewTransactionHandler(mockService)

	app := fiber.New()
	app.Get("/transactions/report", mockAuthMiddleware(1), transactionHandler.GetReport)

	startDate := "2023-01-01"
	endDate := "2023-01-31"
	
	mockData := []entity.CategoryBreakdown{}
	mockService.On("GetReport", uint(1), startDate, endDate, (*uint)(nil), (*string)(nil)).Return(mockData, nil)

	req := httptest.NewRequest("GET", "/transactions/report?start_date=2023-01-01&end_date=2023-01-31", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetReport_WithFilters(t *testing.T) {
	mockService := new(MockTransactionService)
	transactionHandler := handler.NewTransactionHandler(mockService)

	app := fiber.New()
	app.Get("/transactions/report", mockAuthMiddleware(1), transactionHandler.GetReport)

	startDate := "2023-01-01"
	endDate := "2023-01-31"
	walletID := uint(5)
	filterType := "income"
	
	mockData := []entity.CategoryBreakdown{}

	mockService.On("GetReport", uint(1), startDate, endDate, mock.MatchedBy(func(id *uint) bool {
		return id != nil && *id == walletID
	}), mock.MatchedBy(func(ft *string) bool {
		return ft != nil && *ft == filterType
	})).Return(mockData, nil)

	req := httptest.NewRequest("GET", "/transactions/report?start_date=2023-01-01&end_date=2023-01-31&wallet_id=5&type=income", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}
