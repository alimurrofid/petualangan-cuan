package service

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"
	"time"
)

type DashboardService interface {
	GetDashboardData(userID uint) (*entity.DashboardData, error)
}

type dashboardService struct {
	transactionRepo repository.TransactionRepository
	walletRepo      repository.WalletRepository
	savingGoalRepo  repository.SavingGoalRepository
}

func NewDashboardService(transactionRepo repository.TransactionRepository, walletRepo repository.WalletRepository, savingGoalRepo repository.SavingGoalRepository) DashboardService {
	return &dashboardService{
		transactionRepo: transactionRepo,
		walletRepo:      walletRepo,
		savingGoalRepo:  savingGoalRepo,
	}
}

func (s *dashboardService) GetDashboardData(userID uint) (*entity.DashboardData, error) {
	// 1. Get Wallets and Calculate Total Balance
	wallets, err := s.walletRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	var totalBalance float64
	var totalAvailableBalance float64

	for i := range wallets {
		// Calculate Active Contributions for each wallet
		activeContributions, err := s.savingGoalRepo.GetActiveContributions(wallets[i].ID)
		if err != nil {
			return nil, err
		}
		
		wallets[i].AvailableBalance = wallets[i].Balance - activeContributions
		
		totalBalance += wallets[i].Balance
		totalAvailableBalance += wallets[i].AvailableBalance
	}

	// 2. Dates for Current Month
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local).Format("2006-01-02")
	endOfMonth := time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, time.Local).Format("2006-01-02")

	// 3. Get Monthly Summary (Income/Expense) - Reusing GetCategoryBreakdown or FindSummaryByDateRange?
	// FindSummaryByDateRange gives daily. GetCategoryBreakdown gives breakdown.
	// Let's use GetCategoryBreakdown for expense breakdown AND total calculation
	expenseFilter := "expense"
	expenseBreakdown, err := s.transactionRepo.GetCategoryBreakdown(userID, startOfMonth, endOfMonth, nil, &expenseFilter)
	if err != nil {
		return nil, err
	}

	var totalExpenseMonth float64
	for _, item := range expenseBreakdown {
		totalExpenseMonth += item.TotalAmount
	}

	incomeFilter := "income"
	incomeBreakdown, err := s.transactionRepo.GetCategoryBreakdown(userID, startOfMonth, endOfMonth, nil, &incomeFilter)
	if err != nil {
		return nil, err
	}
	var totalIncomeMonth float64
	for _, item := range incomeBreakdown {
		totalIncomeMonth += item.TotalAmount
	}

	// 4. Get Recent Transactions
	recentTransactions, err := s.transactionRepo.GetRecentTransactions(userID, 5)
	if err != nil {
		return nil, err
	}

	// 5. Get Monthly Trend (Last 6 Months)
	startOfTrend := now.AddDate(0, -5, 0) // 5 months ago + current month = 6 months
	startOfTrendStr := time.Date(startOfTrend.Year(), startOfTrend.Month(), 1, 0, 0, 0, 0, time.Local).Format("2006-01-02")
	monthlyTrend, err := s.transactionRepo.GetMonthlyTrend(userID, startOfTrendStr, endOfMonth)
	if err != nil {
		return nil, err
	}

	return &entity.DashboardData{
		TotalBalance:          totalBalance,
		TotalAvailableBalance: totalAvailableBalance,
		TotalIncomeMonth:      totalIncomeMonth,
		TotalExpenseMonth:  totalExpenseMonth,
		Wallets:            wallets,
		RecentTransactions: recentTransactions,
		MonthlyTrend:       monthlyTrend,
		ExpenseBreakdown:   expenseBreakdown,
	}, nil
}
