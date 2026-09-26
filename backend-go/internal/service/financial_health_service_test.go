package service_test

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository/mock"
	"cuan-backend/internal/service"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
)

func TestGetFinancialHealth_Healthy(t *testing.T) {
	mockRepo := new(mock.TransactionRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	mockDebtRepo := new(mock.DebtRepositoryMock)
	mockUserRepo := new(mock.UserRepositoryMock)
	mockSavingGoalRepo := new(mock.SavingGoalRepositoryMock)

	mockUserRepo.On("FindByID", uint(1)).Return((*entity.User)(nil), fmt.Errorf("not found"))
	mockSavingGoalRepo.On("FindAll", uint(1)).Return([]entity.SavingGoal{}, nil)

	svc := service.NewFinancialHealthService(mockRepo, mockWalletRepo, mockDebtRepo, mockUserRepo, mockSavingGoalRepo)
	userID := uint(1)

	mockSummary := []entity.TransactionSummary{
		{Income: 1000, Expense: 500},
	}
	mockRepo.On("FindSummaryByDateRange", userID, testMock.Anything, testMock.Anything, (*uint)(nil), (*uint)(nil), "").Return(mockSummary, nil)

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

	mockDebtRepo.On("GetTotalPayments", userID, testMock.Anything, testMock.Anything).Return(0.0, nil)
	mockDebtRepo.On("FindByUserID", userID, string(entity.DebtTypePayable)).Return([]entity.Debt{}, nil)

	response, err := svc.GetFinancialHealth(userID)

	assert.NoError(t, err)
	assert.Equal(t, 100.0, response.OverallScore)
	assert.Equal(t, entity.StatusHealthy, response.OverallStatus)

	assert.Len(t, response.Ratios, 3)

	// Rasio Tabungan
	assert.Equal(t, "Rasio Tabungan", response.Ratios[0].Name)
	assert.Equal(t, 0.5, response.Ratios[0].Value)
	assert.Equal(t, entity.StatusHealthy, response.Ratios[0].Status)

	// Dana Darurat
	assert.Equal(t, "Dana Darurat", response.Ratios[1].Name)
	assert.Equal(t, 6.0, response.Ratios[1].Value)
	assert.Equal(t, entity.StatusHealthy, response.Ratios[1].Status)

	// Rasio Utang Terhadap Pendapatan (DSR)
	assert.Equal(t, "Rasio Utang Terhadap Pendapatan", response.Ratios[2].Name)
	assert.Equal(t, 0.0, response.Ratios[2].Value)
	assert.Equal(t, entity.StatusHealthy, response.Ratios[2].Status)

	mockRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
	mockDebtRepo.AssertExpectations(t)
}

func TestGetFinancialHealth_Warning(t *testing.T) {
	mockRepo := new(mock.TransactionRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	mockDebtRepo := new(mock.DebtRepositoryMock)
	mockUserRepo := new(mock.UserRepositoryMock)
	mockSavingGoalRepo := new(mock.SavingGoalRepositoryMock)

	mockUserRepo.On("FindByID", uint(1)).Return((*entity.User)(nil), fmt.Errorf("not found"))
	mockSavingGoalRepo.On("FindAll", uint(1)).Return([]entity.SavingGoal{}, nil)

	svc := service.NewFinancialHealthService(mockRepo, mockWalletRepo, mockDebtRepo, mockUserRepo, mockSavingGoalRepo)
	userID := uint(1)

	// Savings rate = (1000 - 850) / 1000 = 15% (Warning)
	mockSummary := []entity.TransactionSummary{
		{Income: 1000, Expense: 850},
	}
	mockRepo.On("FindSummaryByDateRange", userID, testMock.Anything, testMock.Anything, (*uint)(nil), (*uint)(nil), "").Return(mockSummary, nil)

	// Liquidity = 1700 / 850 = 2.0 bulan (Warning)
	mockWallets := []entity.Wallet{
		{Balance: 1700},
	}
	mockWalletRepo.On("FindByUserID", userID).Return(mockWallets, nil)

	mockTrend := []entity.MonthlyTrend{
		{Date: "2023-01", Expense: 850},
	}
	mockRepo.On("GetMonthlyTrend", userID, testMock.Anything, testMock.Anything).Return(mockTrend, nil)

	// DSR = 350 / 1000 = 35% (Warning)
	mockDebtRepo.On("GetTotalPayments", userID, testMock.Anything, testMock.Anything).Return(350.0, nil)
	mockDebts := []entity.Debt{
		{Remaining: 500, IsPaid: false, Type: entity.DebtTypePayable},
	}
	mockDebtRepo.On("FindByUserID", userID, string(entity.DebtTypePayable)).Return(mockDebts, nil)

	response, err := svc.GetFinancialHealth(userID)

	assert.NoError(t, err)
	assert.Equal(t, 66.0, response.OverallScore)
	assert.Equal(t, entity.StatusWarning, response.OverallStatus)

	assert.Equal(t, entity.StatusWarning, response.Ratios[0].Status)
	assert.Equal(t, entity.StatusWarning, response.Ratios[1].Status)
	assert.Equal(t, entity.StatusWarning, response.Ratios[2].Status)
}

func TestGetFinancialHealth_Danger(t *testing.T) {
	mockRepo := new(mock.TransactionRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	mockDebtRepo := new(mock.DebtRepositoryMock)
	mockUserRepo := new(mock.UserRepositoryMock)
	mockSavingGoalRepo := new(mock.SavingGoalRepositoryMock)

	mockUserRepo.On("FindByID", uint(1)).Return((*entity.User)(nil), fmt.Errorf("not found"))
	mockSavingGoalRepo.On("FindAll", uint(1)).Return([]entity.SavingGoal{}, nil)

	svc := service.NewFinancialHealthService(mockRepo, mockWalletRepo, mockDebtRepo, mockUserRepo, mockSavingGoalRepo)
	userID := uint(1)

	// Arus kas defisit: Savings rate = (1000 - 1200) / 1000 = -20% (Danger)
	mockSummary := []entity.TransactionSummary{
		{Income: 1000, Expense: 1200},
	}
	mockRepo.On("FindSummaryByDateRange", userID, testMock.Anything, testMock.Anything, (*uint)(nil), (*uint)(nil), "").Return(mockSummary, nil)

	// Dana darurat sangat rendah: 200 / 1000 = 0.2 bulan (Danger)
	mockWallets := []entity.Wallet{
		{Balance: 200},
	}
	mockWalletRepo.On("FindByUserID", userID).Return(mockWallets, nil)

	mockTrend := []entity.MonthlyTrend{
		{Date: "2023-01", Expense: 1000},
	}
	mockRepo.On("GetMonthlyTrend", userID, testMock.Anything, testMock.Anything).Return(mockTrend, nil)

	// Cicilan utang membengkak: 500 / 1000 = 50% (Danger)
	mockDebtRepo.On("GetTotalPayments", userID, testMock.Anything, testMock.Anything).Return(500.0, nil)
	mockDebts := []entity.Debt{
		{Remaining: 2000, IsPaid: false, Type: entity.DebtTypePayable},
	}
	mockDebtRepo.On("FindByUserID", userID, string(entity.DebtTypePayable)).Return(mockDebts, nil)

	response, err := svc.GetFinancialHealth(userID)

	assert.NoError(t, err)
	assert.True(t, response.OverallScore < 50.0)
	assert.Equal(t, entity.StatusDanger, response.OverallStatus)

	assert.Equal(t, entity.StatusDanger, response.Ratios[0].Status)
	assert.Equal(t, entity.StatusDanger, response.Ratios[1].Status)
	assert.Equal(t, entity.StatusDanger, response.Ratios[2].Status)
}

func TestGetFinancialHealth_CashDrag(t *testing.T) {
	mockRepo := new(mock.TransactionRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	mockDebtRepo := new(mock.DebtRepositoryMock)
	mockUserRepo := new(mock.UserRepositoryMock)
	mockSavingGoalRepo := new(mock.SavingGoalRepositoryMock)

	mockUserRepo.On("FindByID", uint(1)).Return((*entity.User)(nil), fmt.Errorf("not found"))
	mockSavingGoalRepo.On("FindAll", uint(1)).Return([]entity.SavingGoal{}, nil)

	svc := service.NewFinancialHealthService(mockRepo, mockWalletRepo, mockDebtRepo, mockUserRepo, mockSavingGoalRepo)
	userID := uint(1)

	mockSummary := []entity.TransactionSummary{
		{Income: 10000000, Expense: 2000000},
	}
	mockRepo.On("FindSummaryByDateRange", userID, testMock.Anything, testMock.Anything, (*uint)(nil), (*uint)(nil), "").Return(mockSummary, nil)

	// Dana darurat 20 bulan (> 12 bulan)
	mockWallets := []entity.Wallet{
		{Balance: 40000000},
	}
	mockWalletRepo.On("FindByUserID", userID).Return(mockWallets, nil)

	mockTrend := []entity.MonthlyTrend{
		{Date: "2023-01", Expense: 2000000},
	}
	mockRepo.On("GetMonthlyTrend", userID, testMock.Anything, testMock.Anything).Return(mockTrend, nil)

	mockDebtRepo.On("GetTotalPayments", userID, testMock.Anything, testMock.Anything).Return(0.0, nil)
	mockDebtRepo.On("FindByUserID", userID, string(entity.DebtTypePayable)).Return([]entity.Debt{}, nil)

	response, err := svc.GetFinancialHealth(userID)

	assert.NoError(t, err)
	assert.Equal(t, entity.StatusHealthy, response.OverallStatus)
	assert.Equal(t, entity.StatusHealthy, response.Ratios[1].Status)
	assert.Contains(t, response.Ratios[1].Description, "investasi")
}

func TestGetFinancialHealth_DebtFree(t *testing.T) {
	mockRepo := new(mock.TransactionRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	mockDebtRepo := new(mock.DebtRepositoryMock)
	mockUserRepo := new(mock.UserRepositoryMock)
	mockSavingGoalRepo := new(mock.SavingGoalRepositoryMock)

	mockUserRepo.On("FindByID", uint(1)).Return((*entity.User)(nil), fmt.Errorf("not found"))
	mockSavingGoalRepo.On("FindAll", uint(1)).Return([]entity.SavingGoal{}, nil)

	svc := service.NewFinancialHealthService(mockRepo, mockWalletRepo, mockDebtRepo, mockUserRepo, mockSavingGoalRepo)
	userID := uint(1)

	mockSummary := []entity.TransactionSummary{
		{Income: 5000000, Expense: 2000000},
	}
	mockRepo.On("FindSummaryByDateRange", userID, testMock.Anything, testMock.Anything, (*uint)(nil), (*uint)(nil), "").Return(mockSummary, nil)

	mockWallets := []entity.Wallet{
		{Balance: 10000000},
	}
	mockWalletRepo.On("FindByUserID", userID).Return(mockWallets, nil)

	mockTrend := []entity.MonthlyTrend{
		{Date: "2023-01", Expense: 2000000},
	}
	mockRepo.On("GetMonthlyTrend", userID, testMock.Anything, testMock.Anything).Return(mockTrend, nil)

	mockDebtRepo.On("GetTotalPayments", userID, testMock.Anything, testMock.Anything).Return(0.0, nil)
	mockDebtRepo.On("FindByUserID", userID, string(entity.DebtTypePayable)).Return([]entity.Debt{}, nil)

	response, err := svc.GetFinancialHealth(userID)

	assert.NoError(t, err)
	assert.Equal(t, "Bebas utang! Kondisi keuangan Anda sangat ideal.", response.Ratios[2].Description)
	assert.Equal(t, 0.0, response.Ratios[2].Value)
	assert.Equal(t, entity.StatusHealthy, response.Ratios[2].Status)
}
