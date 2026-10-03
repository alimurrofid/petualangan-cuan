package service

import (
	"bytes"
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"
	"fmt"
	"testing"
	"time"

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

type mockWishlistRepository struct{ mock.Mock }

func (m *mockWishlistRepository) Create(item *entity.WishlistItem) error { return nil }
func (m *mockWishlistRepository) FindAllByUserID(userID uint) ([]entity.WishlistItem, error) {
	args := m.Called(userID)
	return args.Get(0).([]entity.WishlistItem), args.Error(1)
}
func (m *mockWishlistRepository) FindByID(id, userID uint) (*entity.WishlistItem, error) {
	return nil, nil
}
func (m *mockWishlistRepository) Update(item *entity.WishlistItem) error { return nil }
func (m *mockWishlistRepository) Delete(id, userID uint) error          { return nil }
func (m *mockWishlistRepository) MarkAsBought(id, userID uint) error    { return nil }
func (m *mockWishlistRepository) HasRelatedTransactions(categoryID uint) (bool, error) {
	return false, nil
}

type mockWishlistService struct{ mock.Mock }

func (m *mockWishlistService) Create(userID uint, req *StoreWishlistRequest) error {
	args := m.Called(userID, req)
	return args.Error(0)
}
func (m *mockWishlistService) FindAll(userID uint) ([]entity.WishlistItem, error) {
	args := m.Called(userID)
	return args.Get(0).([]entity.WishlistItem), args.Error(1)
}
func (m *mockWishlistService) FindByID(id uint, userID uint) (*entity.WishlistItem, error) {
	return nil, nil
}
func (m *mockWishlistService) Update(id uint, userID uint, req *StoreWishlistRequest) error {
	return nil
}
func (m *mockWishlistService) Delete(id uint, userID uint) error { return nil }
func (m *mockWishlistService) MarkAsBought(id uint, userID uint) error {
	return nil
}

type mockDebtService struct{ mock.Mock }

func (m *mockDebtService) CreateDebt(userID uint, input CreateDebtInput) (*entity.Debt, error) {
	return nil, nil
}
func (m *mockDebtService) GetDebts(userID uint, debtType string) ([]entity.Debt, error) {
	return nil, nil
}
func (m *mockDebtService) GetDebt(id uint, userID uint) (*entity.Debt, error) {
	return nil, nil
}
func (m *mockDebtService) PayDebt(id uint, userID uint, input PayDebtInput) (*entity.Debt, error) {
	args := m.Called(id, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Debt), args.Error(1)
}
func (m *mockDebtService) UpdateDebt(id uint, userID uint, input UpdateDebtInput) (*entity.Debt, error) {
	return nil, nil
}
func (m *mockDebtService) DeleteDebt(id uint, userID uint) error { return nil }
func (m *mockDebtService) DeletePayment(id uint, userID uint) error {
	return nil
}

type mockSavingGoalService struct{ mock.Mock }

func (m *mockSavingGoalService) CreateGoal(userID uint, input CreateGoalInput) (*entity.SavingGoal, error) {
	return nil, nil
}
func (m *mockSavingGoalService) GetGoals(userID uint) ([]entity.SavingGoal, error) {
	return nil, nil
}
func (m *mockSavingGoalService) AddContribution(userID uint, goalID uint, input ContributionInput) (*entity.SavingContribution, error) {
	args := m.Called(userID, goalID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.SavingContribution), args.Error(1)
}
func (m *mockSavingGoalService) UpdateGoal(userID uint, goalID uint, input CreateGoalInput) (*entity.SavingGoal, error) {
	return nil, nil
}
func (m *mockSavingGoalService) DeleteGoal(userID uint, goalID uint) error { return nil }
func (m *mockSavingGoalService) DeleteContribution(userID uint, contributionID uint) error {
	return nil
}
func (m *mockSavingGoalService) FinishGoal(userID uint, goalID uint) error {
	return nil
}

type mockTransactionService struct{ mock.Mock }

func (m *mockTransactionService) CreateTransaction(userID uint, input CreateTransactionInput) (*entity.Transaction, error) {
	args := m.Called(userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Transaction), args.Error(1)
}
func (m *mockTransactionService) GetTransactions(userID uint, params entity.TransactionFilterParams) ([]entity.Transaction, int64, error) {
	return nil, 0, nil
}
func (m *mockTransactionService) GetTransaction(id, userID uint) (*entity.Transaction, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Transaction), args.Error(1)
}
func (m *mockTransactionService) UpdateTransaction(id, userID uint, input CreateTransactionInput) (*entity.Transaction, error) {
	args := m.Called(id, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Transaction), args.Error(1)
}
func (m *mockTransactionService) DeleteTransaction(id, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}
func (m *mockTransactionService) TransferTransaction(userID uint, input TransferTransactionInput) error {
	args := m.Called(userID, input)
	return args.Error(0)
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

type mockUserRepository struct{ mock.Mock }

func (m *mockUserRepository) Create(user *entity.User) error { return nil }
func (m *mockUserRepository) FindByEmail(email string) (*entity.User, error) { return nil, nil }
func (m *mockUserRepository) FindByPhone(phone string) (*entity.User, error) { return nil, nil }
func (m *mockUserRepository) Update(user *entity.User) error                 { return nil }
func (m *mockUserRepository) FindByID(id uint) (*entity.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func TestChatbotService_DetectIntent(t *testing.T) {
	assert.Equal(t, IntentSmallTalk, DetectIntent("halo selamat pagi"))
	assert.Equal(t, IntentWishlist, DetectIntent("tampilkan wishlist saya"))
	assert.Equal(t, IntentWishlist, DetectIntent("saya mau beli ps5"))
	assert.Equal(t, IntentDebt, DetectIntent("berapa sisa utang saya"))
	assert.Equal(t, IntentGoal, DetectIntent("progres target tabungan"))
	assert.Equal(t, IntentHealth, DetectIntent("cek kesehatan keuangan saya"))
	assert.Equal(t, IntentReport, DetectIntent("rekap pengeluaran minggu ini"))
	assert.Equal(t, IntentTransaction, DetectIntent("beli bensin 20rb"))
}

func TestChatbotService_GetUserContext(t *testing.T) {
	mockWalletRepo := new(mockWalletRepository)
	mockCategoryRepo := new(mockCategoryRepository)
	mockTransactionRepo := new(mockTransactionRepository)
	mockDebtRepo := new(mockDebtRepository)
	mockDebtSvc := new(mockDebtService)
	mockGoalRepo := new(mockSavingGoalRepository)
	mockGoalSvc := new(mockSavingGoalService)
	mockWishlistRepo := new(mockWishlistRepository)
	mockWishlistSvc := new(mockWishlistService)
	mockTxSvc := new(mockTransactionService)
	mockDashSvc := new(mockDashboardService)
	mockHealthSvc := new(mockFinancialHealthService)

	mockUserRepo := &mockUserRepository{}
	mockUserRepo.On("FindByID", uint(1)).Return((*entity.User)(nil), fmt.Errorf("not found"))

	service := NewChatbotService(
		mockWalletRepo, mockCategoryRepo, mockTxSvc, mockTransactionRepo,
		mockDebtRepo, mockDebtSvc, mockGoalRepo, mockGoalSvc,
		mockWishlistRepo, mockWishlistSvc, mockDashSvc, mockHealthSvc, mockUserRepo,
	)

	mockDashSvc.On("GetDashboardData", uint(1)).Return(&entity.DashboardData{
		TotalBalance: 1000000,
		ExpenseBreakdown: []entity.CategoryBreakdown{
			{CategoryName: "Makan", TotalAmount: 500000, Percentage: 50.0},
		},
	}, nil)

	mockWalletRepo.On("FindByUserID", uint(1)).Return([]entity.Wallet{
		{ID: 1, Name: "Cash", Balance: 500000, Type: "cash"},
	}, nil)

	mockTransactionRepo.On("GetRecentTransactions", uint(1), 5).Return([]entity.Transaction{}, nil)
	mockTransactionRepo.On("FindSummaryByDateRange", uint(1), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entity.TransactionSummary{}, nil)
	
	overdueDate := time.Now().AddDate(0, 0, -2)
	mockDebtRepo.On("FindByUserID", uint(1), "").Return([]entity.Debt{
		{ID: 3, Name: "Utang Budi", Amount: 200000, Remaining: 100000, Type: entity.DebtTypePayable, DueDate: &overdueDate},
	}, nil)

	mockGoalRepo.On("FindAll", uint(1)).Return([]entity.SavingGoal{
		{ID: 5, Name: "Laptop", TargetAmount: 10000000, CurrentAmount: 2000000},
	}, nil)

	mockWishlistRepo.On("FindAllByUserID", uint(1)).Return([]entity.WishlistItem{
		{ID: 10, Name: "AirPods Pro", EstimatedPrice: 3500000, Priority: entity.WishlistPriorityHigh, IsBought: false},
	}, nil)

	mockHealthSvc.On("GetFinancialHealth", uint(1)).Return(entity.FinancialHealthResponse{
		OverallStatus: "Sehat",
		OverallScore:  85,
		Ratios: []entity.FinancialHealthRatio{
			{Name: "Dana Darurat", FormattedValue: "3x pengeluaran", Target: "3-6x", Status: "Sehat"},
		},
	}, nil)

	contextStr := service.GetUserContext(1, "cek data keuangan saya")

	assert.Contains(t, contextStr, "Daftar Wallet (1):")
	assert.Contains(t, contextStr, "Cash")
	assert.Contains(t, contextStr, "85/100 (Sehat)")
	assert.Contains(t, contextStr, "Dana Darurat: 3x pengeluaran")
	assert.Contains(t, contextStr, "Utang Budi")
	assert.Contains(t, contextStr, "OVERDUE!")
	assert.Contains(t, contextStr, "[ID: 5] Laptop")
	assert.Contains(t, contextStr, "AirPods Pro [high]")
}

func TestChatbotService_SaveTransactions_Actions(t *testing.T) {
	mockWalletRepo := new(mockWalletRepository)
	mockCategoryRepo := new(mockCategoryRepository)
	mockTransactionRepo := new(mockTransactionRepository)
	mockDebtRepo := new(mockDebtRepository)
	mockDebtSvc := new(mockDebtService)
	mockGoalRepo := new(mockSavingGoalRepository)
	mockGoalSvc := new(mockSavingGoalService)
	mockWishlistRepo := new(mockWishlistRepository)
	mockWishlistSvc := new(mockWishlistService)
	mockTxSvc := new(mockTransactionService)
	mockDashSvc := new(mockDashboardService)
	mockHealthSvc := new(mockFinancialHealthService)
	mockUserRepo := new(mockUserRepository)

	service := NewChatbotService(
		mockWalletRepo, mockCategoryRepo, mockTxSvc, mockTransactionRepo,
		mockDebtRepo, mockDebtSvc, mockGoalRepo, mockGoalSvc,
		mockWishlistRepo, mockWishlistSvc, mockDashSvc, mockHealthSvc, mockUserRepo,
	)

	mockWalletRepo.On("FindByUserID", uint(1)).Return([]entity.Wallet{
		{ID: 1, Name: "BCA"},
		{ID: 2, Name: "GoPay"},
	}, nil)

	mockCategoryRepo.On("FindAll", uint(1)).Return([]entity.Category{
		{ID: 1, Name: "Makan", Type: "expense"},
		{ID: 2, Name: "Belanja", Type: "expense"},
	}, nil)

	// Test 1: Action Transfer
	mockTxSvc.On("TransferTransaction", uint(1), mock.MatchedBy(func(input TransferTransactionInput) bool {
		return input.FromWalletID == 1 && input.ToWalletID == 2 && input.Amount == 100000
	})).Return(nil)

	resTransfer, err := service.SaveTransactions(1, []entity.TransactionItemAI{
		{
			Action:       "transfer",
			Amount:       100000,
			WalletName:   "BCA",
			ToWalletName: "GoPay",
			Description:  "Transfer ke GoPay",
		},
	})
	assert.NoError(t, err)
	assert.Len(t, resTransfer, 1)
	assert.Equal(t, "transfer", resTransfer[0].Action)
	assert.Equal(t, "BCA", resTransfer[0].WalletName)
	assert.Equal(t, "GoPay", resTransfer[0].ToWalletName)

	// Test 2: Action Pay Debt
	mockDebtSvc.On("PayDebt", uint(3), uint(1), mock.MatchedBy(func(input PayDebtInput) bool {
		return input.WalletID == 1 && input.Amount == 50000
	})).Return(&entity.Debt{ID: 3, Name: "Utang Budi"}, nil)

	resDebt, err := service.SaveTransactions(1, []entity.TransactionItemAI{
		{
			Action:      "pay_debt",
			ID:          3,
			Amount:      50000,
			WalletName:  "BCA",
			Description: "Bayar utang Budi",
		},
	})
	assert.NoError(t, err)
	assert.Len(t, resDebt, 1)
	assert.Equal(t, "pay_debt", resDebt[0].Action)

	// Test 3: Action Save Goal
	mockGoalSvc.On("AddContribution", uint(1), uint(7), mock.MatchedBy(func(input ContributionInput) bool {
		return input.WalletID == 1 && input.Amount == 200000
	})).Return(&entity.SavingContribution{ID: 99}, nil)

	resGoal, err := service.SaveTransactions(1, []entity.TransactionItemAI{
		{
			Action:      "save_goal",
			ID:          7,
			Amount:      200000,
			WalletName:  "BCA",
			Description: "Setor Tabungan Laptop",
		},
	})
	assert.NoError(t, err)
	assert.Len(t, resGoal, 1)
	assert.Equal(t, "save_goal", resGoal[0].Action)

	// Test 4: Action Create Wishlist
	mockWishlistSvc.On("Create", uint(1), mock.MatchedBy(func(req *StoreWishlistRequest) bool {
		return req.Name == "Sepatu Nike" && req.EstimatedPrice == 1500000 && req.Priority == "medium"
	})).Return(nil)

	resWishlist, err := service.SaveTransactions(1, []entity.TransactionItemAI{
		{
			Action:       "create_wishlist",
			Amount:       1500000,
			Description:  "Sepatu Nike",
			CategoryName: "Belanja",
			Priority:     "medium",
		},
	})
	assert.NoError(t, err)
	assert.Len(t, resWishlist, 1)
	assert.Equal(t, "create_wishlist", resWishlist[0].Action)
}
