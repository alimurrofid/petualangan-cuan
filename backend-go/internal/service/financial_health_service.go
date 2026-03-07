package service

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"
	pkgutils "cuan-backend/pkg/utils"
	"fmt"
	"math"
	"time"

	"github.com/rs/zerolog/log"
)

type FinancialHealthService interface {
	GetFinancialHealth(userID uint) (entity.FinancialHealthResponse, error)
}

type financialHealthService struct {
	transactionRepo repository.TransactionRepository
	walletRepo      repository.WalletRepository
	debtRepo        repository.DebtRepository
	userRepo        repository.UserRepository
	savingGoalRepo  repository.SavingGoalRepository // TAMBAHAN: Inject Saving Goal Repo
}

func NewFinancialHealthService(
	transactionRepo repository.TransactionRepository,
	walletRepo repository.WalletRepository,
	debtRepo repository.DebtRepository,
	userRepo repository.UserRepository,
	savingGoalRepo repository.SavingGoalRepository, // TAMBAHAN: Inject Saving Goal Repo
) FinancialHealthService {
	return &financialHealthService{
		transactionRepo: transactionRepo,
		walletRepo:      walletRepo,
		debtRepo:        debtRepo,
		userRepo:        userRepo,
		savingGoalRepo:  savingGoalRepo, // TAMBAHAN
	}
}

func (s *financialHealthService) GetFinancialHealth(userID uint) (entity.FinancialHealthResponse, error) {
	// Resolve payday with safe fallback to 1
	payday := 1
	if user, err := s.userRepo.FindByID(userID); err == nil && user.Payday != nil {
		payday = *user.Payday
	}

	now := time.Now()
	startCycle, endCycle := pkgutils.GetBillingCycle(now, payday)

	startDate := startCycle.Format("2006-01-02")
	endDate := endCycle.Format("2006-01-02")

	// 1. SAVINGS RATE : (Total Income - Total Expense) / Total Income
	summary, err := s.transactionRepo.FindSummaryByDateRange(userID, startDate, endDate, nil, nil, "")
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Msg("Failed to fetch transaction summary")
		return entity.FinancialHealthResponse{}, err
	}

	var totalIncomeMonth, totalExpenseMonth float64
	for _, item := range summary {
		totalIncomeMonth += item.Income
		totalExpenseMonth += item.Expense
	}

	savingsRate := 0.0
	if totalIncomeMonth > 0 {
		savingsRate = (totalIncomeMonth - totalExpenseMonth) / totalIncomeMonth
		// PERBAIKAN: Batasi persentase minimal di -100% agar tidak muncul -9340%
		if savingsRate < -1.0 {
			savingsRate = -1.0
		}
	} else if totalExpenseMonth > 0 {
		savingsRate = -1.0
	}

	savingsRatio := entity.FinancialHealthRatio{
		Name:           "Rasio Tabungan",
		Value:          savingsRate,
		Target:         "> 20%",
		FormattedValue: fmt.Sprintf("%.1f%%", savingsRate*100),
	}

	if savingsRate >= 0.20 {
		savingsRatio.Status = entity.StatusHealthy
		savingsRatio.Description = "Hebat! Anda menabung dengan porsi yang sehat."
	} else if savingsRate >= 0.10 {
		savingsRatio.Status = entity.StatusWarning
		savingsRatio.Description = "Cukup baik, tapi coba tingkatkan lagi tabungan Anda."
	} else {
		savingsRatio.Status = entity.StatusDanger
		savingsRatio.Description = "Hati-hati, pengeluaran Anda melebihi pendapatan di siklus ini."
	}

	// 2. LIQUIDITY RATIO & TOTAL ASSETS
	wallets, err := s.walletRepo.FindByUserID(userID)
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Msg("Failed to fetch wallets")
		return entity.FinancialHealthResponse{}, err
	}

	totalAssets := 0.0
	for _, w := range wallets {
		totalAssets += w.Balance
	}

	// PERBAIKAN: Tambahkan saldo Target Menabung sebagai bagian dari Total Aset Anda
	savingGoals, err := s.savingGoalRepo.FindAll(userID)
	if err == nil {
		for _, sg := range savingGoals {
			totalAssets += sg.CurrentAmount
		}
	}

	// 3-month trend: from 3 cycles ago up to (but not including) the current cycle start
	startOf3CyclesAgo := startCycle.AddDate(0, -3, 0)
	endOfLastCycle := startCycle.Add(-time.Second)

	trend, err := s.transactionRepo.GetMonthlyTrend(userID, startOf3CyclesAgo.Format("2006-01-02"), endOfLastCycle.Format("2006-01-02"))
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Msg("Failed to fetch monthly trend")
		return entity.FinancialHealthResponse{}, err
	}

	totalExpense3Months := 0.0
	monthsCount := 0
	for _, item := range trend {
		if item.Expense > 0 {
			totalExpense3Months += item.Expense
			monthsCount++
		}
	}

	avgMonthlyExpense := 0.0
	if monthsCount > 0 {
		avgMonthlyExpense = totalExpense3Months / float64(monthsCount)
	} else if totalExpenseMonth > 0 {
		avgMonthlyExpense = totalExpenseMonth
	}

	liquidityScore := 0.0 // In Months
	if avgMonthlyExpense > 0 {
		liquidityScore = totalAssets / avgMonthlyExpense // Menggunakan TotalAssets
	} else if totalAssets > 0 {
		liquidityScore = 999 // Infinite liquidity
	}

	liquidityRatio := entity.FinancialHealthRatio{
		Name:           "Dana Darurat",
		Value:          liquidityScore,
		Target:         "3 - 6 Bulan",
		FormattedValue: fmt.Sprintf("%.1f Bulan", liquidityScore),
	}

	if liquidityScore >= 3 && liquidityScore <= 12 {
		liquidityRatio.Status = entity.StatusHealthy
		liquidityRatio.Description = "Dana darurat Anda aman untuk menutupi pengeluaran mendadak."
	} else if liquidityScore > 12 {
		if avgMonthlyExpense < 500000 {
			liquidityRatio.Status = entity.StatusWarning
			liquidityRatio.Description = "Saldo aman, namun data pengeluaran bulanan Anda belum lengkap untuk kalkulasi akurat."
		} else {
			liquidityRatio.Status = entity.StatusHealthy
			liquidityRatio.Description = "Dana darurat sangat berlimpah."
		}
	} else if liquidityScore >= 1 {
		liquidityRatio.Status = entity.StatusWarning
		liquidityRatio.Description = "Dana darurat ada, namun perlu ditingkatkan untuk keamanan ekstra."
	} else {
		liquidityRatio.Status = entity.StatusDanger
		liquidityRatio.Description = "Bahaya! Segera sisihkan uang untuk dana darurat minimal 1 bulan pengeluaran."
	}

	// 3. DEBT-TO-ASSET RATIO
	allDebts, err := s.debtRepo.FindByUserID(userID, "")
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Msg("Failed to fetch debts")
		return entity.FinancialHealthResponse{}, err
	}

	totalSisaHutang := 0.0
	for _, debt := range allDebts {
		if !debt.IsPaid && debt.Type != entity.DebtTypeReceivable {
			totalSisaHutang += debt.Remaining
		}
	}

	debtRatio := 0.0
	if totalAssets > 0 {
		debtRatio = totalSisaHutang / totalAssets // Menggunakan TotalAssets
	} else if totalSisaHutang > 0 {
		debtRatio = 1.0 // 100% (all debt, no assets)
	}

	debtRatioStruct := entity.FinancialHealthRatio{
		Name:           "Rasio Hutang Terhadap Aset",
		Value:          debtRatio,
		Target:         "< 35%",
		FormattedValue: fmt.Sprintf("%.1f%%", debtRatio*100),
	}

	if totalSisaHutang == 0 {
		debtRatioStruct.Status = entity.StatusHealthy
		debtRatioStruct.Description = "Bebas utang! Kondisi sangat ideal."
	} else if debtRatio <= 0.35 {
		debtRatioStruct.Status = entity.StatusHealthy
		debtRatioStruct.Description = "Porsi utang aman dibandingkan aset."
	} else if debtRatio <= 0.50 {
		debtRatioStruct.Status = entity.StatusWarning
		debtRatioStruct.Description = "Waspada, saldo terancam habis jika semua utang ditagih."
	} else {
		debtRatioStruct.Status = entity.StatusDanger
		debtRatioStruct.Description = "Bahaya! Sisa hutang terlalu besar dibanding uang yang Anda miliki."
	}

	// Overall Score Calculation
	score := 0.0
	if savingsRatio.Status == entity.StatusHealthy {
		score += 100
	}
	if savingsRatio.Status == entity.StatusWarning {
		score += 50
	}
	if liquidityRatio.Status == entity.StatusHealthy {
		score += 100
	}
	if liquidityRatio.Status == entity.StatusWarning {
		score += 50
	}
	if debtRatioStruct.Status == entity.StatusHealthy {
		score += 100
	}
	if debtRatioStruct.Status == entity.StatusWarning {
		score += 50
	}

	overallScore := score / 3.0
	overallStatus := entity.StatusWarning
	if overallScore >= 80 {
		overallStatus = entity.StatusHealthy
	} else if overallScore < 40 {
		overallStatus = entity.StatusDanger
	}

	return entity.FinancialHealthResponse{
		OverallScore:  math.Round(overallScore),
		OverallStatus: overallStatus,
		Ratios: []entity.FinancialHealthRatio{
			savingsRatio,
			liquidityRatio,
			debtRatioStruct,
		},
	}, nil
}
