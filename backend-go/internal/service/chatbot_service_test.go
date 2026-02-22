package service

import (
	"bytes"
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"
	"testing"

	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockWalletRepository struct{ mock.Mock }

func (m *mockWalletRepository) Create(wallet *entity.Wallet) error { return nil }
func (m *mockWalletRepository) FindByUserID(userID uint) ([]entity.Wallet, error) {
	args := m.Called(userID)
	return args.Get(0).([]entity.Wallet), args.Error(1)
}
func (m *mockWalletRepository) FindByID(id uint, userID uint) (*entity.Wallet, error) {
	return nil, nil
}
func (m *mockWalletRepository) Update(wallet *entity.Wallet) error             { return nil }
func (m *mockWalletRepository) Delete(id uint, userID uint) error              { return nil }
func (m *mockWalletRepository) WithTx(tx *gorm.DB) repository.WalletRepository { return m }
func (m *mockWalletRepository) AdjustBalance(tx interface{}, id uint, amount float64) error {
	return nil
}

type mockCategoryRepository struct{ mock.Mock }

func (m *mockCategoryRepository) FindByUserID(userID uint) ([]entity.Category, error) {
	return nil, nil
}
func (m *mockCategoryRepository) FindByID(id uint, userID uint) (*entity.Category, error) {
	return nil, nil
}
func (m *mockCategoryRepository) Create(category *entity.Category) error       { return nil }
func (m *mockCategoryRepository) Update(category *entity.Category) error       { return nil }
func (m *mockCategoryRepository) Delete(id uint, userID uint) error            { return nil }
func (m *mockCategoryRepository) HasRelatedTransactions(id uint) (bool, error) { return false, nil }
func (m *mockCategoryRepository) FindAll(userID uint) ([]entity.Category, error) {
	args := m.Called(userID)
	return args.Get(0).([]entity.Category), args.Error(1)
}
func (m *mockCategoryRepository) WithTx(tx *gorm.DB) repository.CategoryRepository { return m }

type mockTransactionRepository struct{ mock.Mock }

func (m *mockTransactionRepository) Create(transaction *entity.Transaction) error { return nil }
func (m *mockTransactionRepository) FindAll(userID uint, params entity.TransactionFilterParams) ([]entity.Transaction, int64, error) {
	return nil, 0, nil
}
func (m *mockTransactionRepository) FindByID(id, userID uint) (*entity.Transaction, error) {
	return nil, nil
}
func (m *mockTransactionRepository) Update(transaction *entity.Transaction) error { return nil }
func (m *mockTransactionRepository) Delete(id uint, userID uint) error            { return nil }
func (m *mockTransactionRepository) FindSummaryByDateRange(userID uint, startDate, endDate string, walletID, categoryID *uint, search string) ([]entity.TransactionSummary, error) {
	args := m.Called(userID, startDate, endDate, walletID, categoryID, search)
	return args.Get(0).([]entity.TransactionSummary), args.Error(1)
}
func (m *mockTransactionRepository) GetCategoryBreakdown(userID uint, startDate, endDate string, walletIDs []uint, filterType *string) ([]entity.CategoryBreakdown, error) {
	return nil, nil
}
func (m *mockTransactionRepository) GetMonthlyTrend(userID uint, startDate, endDate string) ([]entity.MonthlyTrend, error) {
	return nil, nil
}
func (m *mockTransactionRepository) GetRecentTransactions(userID uint, limit int) ([]entity.Transaction, error) {
	args := m.Called(userID, limit)
	return args.Get(0).([]entity.Transaction), args.Error(1)
}
func (m *mockTransactionRepository) WithTx(tx *gorm.DB) repository.TransactionRepository { return m }

type mockDebtRepository struct{ mock.Mock }

func (m *mockDebtRepository) Create(debt *entity.Debt) error             { return nil }
func (m *mockDebtRepository) FindAll(userID uint) ([]entity.Debt, error) { return nil, nil }
func (m *mockDebtRepository) FindByUserID(userID uint, filter string) ([]entity.Debt, error) {
	args := m.Called(userID, filter)
	return args.Get(0).([]entity.Debt), args.Error(1)
}
func (m *mockDebtRepository) FindByID(id, userID uint) (*entity.Debt, error) { return nil, nil }
func (m *mockDebtRepository) Update(debt *entity.Debt) error                 { return nil }
func (m *mockDebtRepository) Delete(id, userID uint) error                   { return nil }
func (m *mockDebtRepository) GetTotalPayments(debtID uint, startDate, endDate string) (float64, error) {
	return 0, nil
}
func (m *mockDebtRepository) WithTx(tx *gorm.DB) repository.DebtRepository { return m }

type mockSavingGoalRepository struct{ mock.Mock }

func (m *mockSavingGoalRepository) Create(goal *entity.SavingGoal) error { return nil }
func (m *mockSavingGoalRepository) FindAll(userID uint) ([]entity.SavingGoal, error) {
	args := m.Called(userID)
	return args.Get(0).([]entity.SavingGoal), args.Error(1)
}
func (m *mockSavingGoalRepository) FindByID(id, userID uint) (*entity.SavingGoal, error) {
	return nil, nil
}
func (m *mockSavingGoalRepository) Update(goal *entity.SavingGoal) error { return nil }
func (m *mockSavingGoalRepository) Delete(goal *entity.SavingGoal) error { return nil }
func (m *mockSavingGoalRepository) AddContribution(contrib *entity.SavingContribution) error {
	return nil
}
func (m *mockSavingGoalRepository) DeleteContribution(contrib *entity.SavingContribution) error {
	return nil
}
func (m *mockSavingGoalRepository) DeleteContributions(goalID uint) error { return nil }
func (m *mockSavingGoalRepository) FindContributionByID(id uint) (*entity.SavingContribution, error) {
	return nil, nil
}
func (m *mockSavingGoalRepository) GetActiveContributions(goalID uint) (float64, error) {
	return 0, nil
}
func (m *mockSavingGoalRepository) WithTx(tx *gorm.DB) repository.SavingGoalRepository { return m }

type mockTransactionService struct{ mock.Mock }

func (m *mockTransactionService) CreateTransaction(userID uint, input CreateTransactionInput) (*entity.Transaction, error) {
	return nil, nil
}
func (m *mockTransactionService) GetTransactions(userID uint, params entity.TransactionFilterParams) ([]entity.Transaction, int64, error) {
	return nil, 0, nil
}
func (m *mockTransactionService) GetTransaction(id, userID uint) (*entity.Transaction, error) {
	return nil, nil
}
func (m *mockTransactionService) UpdateTransaction(id, userID uint, input CreateTransactionInput) (*entity.Transaction, error) {
	return nil, nil
}
func (m *mockTransactionService) DeleteTransaction(id, userID uint) error { return nil }
func (m *mockTransactionService) TransferTransaction(userID uint, input TransferTransactionInput) error {
	return nil
}
func (m *mockTransactionService) GetCalendarData(userID uint, startDate, endDate string, walletID, categoryID *uint, search string) ([]entity.TransactionSummary, error) {
	return nil, nil
}
func (m *mockTransactionService) GetReport(userID uint, startDate, endDate string, walletIDs []uint, transactionType *string) ([]entity.CategoryBreakdown, error) {
	return nil, nil
}
func (m *mockTransactionService) ExportTransactions(userID uint, params entity.TransactionFilterParams) (*bytes.Buffer, error) {
	return nil, nil
}
func (m *mockTransactionService) ExportReport(userID uint, startDate, endDate string, walletIDs []uint, transactionType *string) (*bytes.Buffer, error) {
	return nil, nil
}

type mockDashboardService struct{ mock.Mock }

func (m *mockDashboardService) GetDashboardData(userID uint) (*entity.DashboardData, error) {
	args := m.Called(userID)
	return args.Get(0).(*entity.DashboardData), args.Error(1)
}

type mockFinancialHealthService struct{ mock.Mock }

func (m *mockFinancialHealthService) GetFinancialHealth(userID uint) (entity.FinancialHealthResponse, error) {
	args := m.Called(userID)
	return args.Get(0).(entity.FinancialHealthResponse), args.Error(1)
}

func TestChatbotService_GetUserContext(t *testing.T) {
	mockWalletRepo := new(mockWalletRepository)
	mockCategoryRepo := new(mockCategoryRepository)
	mockTransactionRepo := new(mockTransactionRepository)
	mockDebtRepo := new(mockDebtRepository)
	mockGoalRepo := new(mockSavingGoalRepository)
	mockTxSvc := new(mockTransactionService)
	mockDashSvc := new(mockDashboardService)
	mockHealthSvc := new(mockFinancialHealthService)

	service := NewChatbotService(
		mockWalletRepo, mockCategoryRepo, mockTxSvc, mockTransactionRepo, mockDebtRepo, mockGoalRepo, mockDashSvc, mockHealthSvc,
	)

	mockDashSvc.On("GetDashboardData", uint(1)).Return(&entity.DashboardData{TotalBalance: 1000}, nil)
	mockWalletRepo.On("FindByUserID", uint(1)).Return([]entity.Wallet{
		{ID: 1, Name: "Cash"},
	}, nil)

	mockTransactionRepo.On("GetRecentTransactions", uint(1), 5).Return([]entity.Transaction{}, nil)
	mockTransactionRepo.On("FindSummaryByDateRange", uint(1), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entity.TransactionSummary{}, nil)
	mockDebtRepo.On("FindByUserID", uint(1), "").Return([]entity.Debt{}, nil)
	mockGoalRepo.On("FindAll", uint(1)).Return([]entity.SavingGoal{}, nil)
	mockHealthSvc.On("GetFinancialHealth", uint(1)).Return(entity.FinancialHealthResponse{OverallStatus: "Good", OverallScore: 85}, nil)

	contextStr := service.GetUserContext(1, "cek data keuangan saya")

	assert.Contains(t, contextStr, "Daftar Wallet (1):")
	assert.Contains(t, contextStr, "Cash")
	assert.Contains(t, contextStr, "85/100 (Good)")
}
