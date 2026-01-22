package service_test

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository/mock"
	"cuan-backend/internal/service"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
)

func TestGetFinancialHealth(t *testing.T) {
	mockRepo := new(mock.TransactionRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	mockDebtRepo := new(mock.DebtRepositoryMock)

	svc := service.NewFinancialHealthService(mockRepo, mockWalletRepo, mockDebtRepo)
	userID := uint(1)

	// Mock Date Range Logic (Current Month)
	now := time.Now()
	_ = now
	// currentYear, currentMonth, _ := now.Date()
	// currentLocation := now.Location()
	// firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	// lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	// 1. Savings Rate Mocks
	// Income: 1000, Expense: 500 => Savings: 500 (50%) -> Healthy
	mockSummary := []entity.TransactionSummary{
		{Income: 1000, Expense: 500},
	}
	mockRepo.On("FindSummaryByDateRange", userID, testMock.Anything, testMock.Anything, (*uint)(nil), (*uint)(nil), "").Return(mockSummary, nil)

	// 2. Liquidity Mocks
	// Wallet Balance: 3000
	// Avg Expense (3 months): 500 (derived from trend)
	// Liquidity: 3000 / 500 = 6 months -> Healthy
	mockWallets := []entity.Wallet{
		{Balance: 3000},
	}
	mockWalletRepo.On("FindByUserID", userID).Return(mockWallets, nil)

	mockTrend := []entity.MonthlyTrend{
		{Date: "2023-01", Expense: 500},
		{Date: "2023-02", Expense: 500},
		{Date: "2023-03", Expense: 500},
	}
	mockRepo.On("GetMonthlyTrend", userID, testMock.Anything, testMock.Anything).Return(mockTrend, nil)

	// 3. Debt-to-Income Mocks
	// Total Payments: 100
	// Income: 1000
	// DTI: 100 / 1000 = 10% -> Healthy
	mockDebtRepo.On("GetTotalPayments", userID, testMock.Anything, testMock.Anything).Return(100.0, nil)

	// Execute
	response, err := svc.GetFinancialHealth(userID)

	assert.NoError(t, err)
	
	// Assertions
	assert.Equal(t, 100.0, response.OverallScore) // All Healthy
	assert.Equal(t, entity.StatusHealthy, response.OverallStatus)
	
	assert.Len(t, response.Ratios, 3)
	
	// Savings
	assert.Equal(t, "Savings Rate", response.Ratios[0].Name)
	assert.Equal(t, 0.5, response.Ratios[0].Value)
	assert.Equal(t, entity.StatusHealthy, response.Ratios[0].Status)

	// Liquidity
	assert.Equal(t, "Dana Darurat", response.Ratios[1].Name)
	assert.Equal(t, 6.0, response.Ratios[1].Value)
	assert.Equal(t, entity.StatusHealthy, response.Ratios[1].Status)

	// DTI
	assert.Equal(t, "Debt-to-Income", response.Ratios[2].Name)
	assert.Equal(t, 0.1, response.Ratios[2].Value)
	assert.Equal(t, entity.StatusHealthy, response.Ratios[2].Status)

	mockRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
	mockDebtRepo.AssertExpectations(t)
}

func TestGetFinancialHealth_Warning(t *testing.T) {
	mockRepo := new(mock.TransactionRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	mockDebtRepo := new(mock.DebtRepositoryMock)

	svc := service.NewFinancialHealthService(mockRepo, mockWalletRepo, mockDebtRepo)
	userID := uint(1)

	// 1. Savings Rate Mocks
	// Income: 1000, Expense: 900 => Savings: 100 (10%) -> Warning
	mockSummary := []entity.TransactionSummary{
		{Income: 1000, Expense: 900},
	}
	mockRepo.On("FindSummaryByDateRange", userID, testMock.Anything, testMock.Anything, (*uint)(nil), (*uint)(nil), "").Return(mockSummary, nil)

	// 2. Liquidity Mocks
	// Wallet Balance: 1000
	// Avg Expense: 900
	// Liquidity: 1.1 months -> Warning
	mockWallets := []entity.Wallet{
		{Balance: 1000},
	}
	mockWalletRepo.On("FindByUserID", userID).Return(mockWallets, nil)

	mockTrend := []entity.MonthlyTrend{
		{Date: "2023-01", Expense: 900},
	}
	mockRepo.On("GetMonthlyTrend", userID, testMock.Anything, testMock.Anything).Return(mockTrend, nil)

	// 3. Debt-to-Income Mocks
	// Total Payments: 400
	// Income: 1000
	// DTI: 40% -> Warning
	mockDebtRepo.On("GetTotalPayments", userID, testMock.Anything, testMock.Anything).Return(400.0, nil)

	// Execute
	response, err := svc.GetFinancialHealth(userID)

	assert.NoError(t, err)
	
	// Score: (50 + 50 + 50) / 3 = 50 -> Warning
	assert.Equal(t, 50.0, response.OverallScore) 
	assert.Equal(t, entity.StatusWarning, response.OverallStatus)
	
	assert.Equal(t, entity.StatusWarning, response.Ratios[0].Status)
	assert.Equal(t, entity.StatusWarning, response.Ratios[1].Status)
	assert.Equal(t, entity.StatusWarning, response.Ratios[2].Status)
}
