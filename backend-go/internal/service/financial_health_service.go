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
	savingGoalRepo  repository.SavingGoalRepository
}

func NewFinancialHealthService(
	transactionRepo repository.TransactionRepository,
	walletRepo repository.WalletRepository,
	debtRepo repository.DebtRepository,
	userRepo repository.UserRepository,
	savingGoalRepo repository.SavingGoalRepository,
) FinancialHealthService {
	return &financialHealthService{
		transactionRepo: transactionRepo,
		walletRepo:      walletRepo,
		debtRepo:        debtRepo,
		userRepo:        userRepo,
		savingGoalRepo:  savingGoalRepo,
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

	savingsScore := 0.0
	if savingsRate >= 0.20 {
		// Nilai 80 s/d 100
		bonus := ((savingsRate - 0.20) / 0.20) * 20.0
		if bonus > 20.0 {
			bonus = 20.0
		}
		savingsScore = 80.0 + bonus
		savingsRatio.Status = entity.StatusHealthy
		savingsRatio.Description = "Hebat! Anda berhasil menyisihkan lebih dari 20% pendapatan untuk masa depan."
	} else if savingsRate >= 0.10 {
		// Nilai 50 s/d 79
		savingsScore = 50.0 + ((savingsRate-0.10)/0.10)*29.0
		savingsRatio.Status = entity.StatusWarning
		savingsRatio.Description = "Cukup baik, namun usahakan tingkatkan tabungan hingga minimal 20% penghasilan."
	} else if savingsRate >= 0.0 {
		// Nilai 25 s/d 49
		savingsScore = 25.0 + (savingsRate/0.10)*24.0
		savingsRatio.Status = entity.StatusDanger
		savingsRatio.Description = "Porsi tabungan sangat minim (< 10%). Waspada terhadap risiko pengeluaran tak terduga."
	} else {
		// Arus kas defisit: Nilai 0 s/d 24
		def := 25.0 + savingsRate*25.0
		if def < 0 {
			def = 0
		}
		savingsScore = def
		savingsRatio.Status = entity.StatusDanger
		savingsRatio.Description = "Hati-hati, pengeluaran Anda melebihi pendapatan di siklus ini (arus kas defisit)."
	}

	// 2. LIQUIDITY RATIO & TOTAL ASSETS
	wallets, err := s.walletRepo.FindByUserID(userID)
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Msg("Failed to fetch wallets")
		return entity.FinancialHealthResponse{}, err
	}

	// Total Liquid Assets adalah total saldo wallet (uang saving goals sudah berada di dalam saldo wallet)
	totalLiquidAssets := 0.0
	for _, w := range wallets {
		totalLiquidAssets += w.Balance
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
		liquidityScore = totalLiquidAssets / avgMonthlyExpense
	} else if totalLiquidAssets > 0 {
		liquidityScore = 12.0 // Cadangan aman jika belum ada catatan pengeluaran
	}

	liquidityRatio := entity.FinancialHealthRatio{
		Name:           "Dana Darurat",
		Value:          liquidityScore,
		Target:         "3 - 6 Bulan",
		FormattedValue: fmt.Sprintf("%.1f Bulan", liquidityScore),
	}

	liquidityPoints := 0.0
	if liquidityScore >= 3.0 && liquidityScore <= 12.0 {
		liquidityPoints = 100.0
		liquidityRatio.Status = entity.StatusHealthy
		liquidityRatio.Description = "Dana darurat Anda aman dan ideal untuk menutupi kebutuhan mendadak."
	} else if liquidityScore > 12.0 {
		if avgMonthlyExpense < 500000 {
			liquidityPoints = 75.0
			liquidityRatio.Status = entity.StatusWarning
			liquidityRatio.Description = "Saldo aman, namun data pengeluaran bulanan Anda belum lengkap untuk kalkulasi akurat."
		} else {
			liquidityPoints = 95.0
			liquidityRatio.Status = entity.StatusHealthy
			liquidityRatio.Description = "Dana darurat sangat berlimpah. Pertimbangkan mendiversifikasi sebagian ke instrumen investasi agar tidak tergerus inflasi."
		}
	} else if liquidityScore >= 1.0 {
		// Nilai 50 s/d 90
		liquidityPoints = 50.0 + ((liquidityScore-1.0)/2.0)*40.0
		liquidityRatio.Status = entity.StatusWarning
		liquidityRatio.Description = "Dana darurat ada, namun perlu ditingkatkan hingga minimal 3 bulan pengeluaran untuk keamanan ekstra."
	} else {
		// Nilai 0 s/d 50
		liquidityPoints = (liquidityScore / 1.0) * 50.0
		liquidityRatio.Status = entity.StatusDanger
		liquidityRatio.Description = "Bahaya! Segera sisihkan uang untuk dana darurat minimal 1 bulan pengeluaran."
	}

	// 3. DEBT SERVICE RATIO (DSR) & DEBT BURDEN
	// Mengukur beban cicilan utang bulanan terhadap pendapatan
	paidDebtThisMonth, err := s.debtRepo.GetTotalPayments(userID, startDate, endDate)
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Msg("Failed to fetch debt payments")
		paidDebtThisMonth = 0.0
	}

	allDebts, err := s.debtRepo.FindByUserID(userID, string(entity.DebtTypePayable))
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Msg("Failed to fetch debts")
		return entity.FinancialHealthResponse{}, err
	}

	totalRemainingDebt := 0.0
	hasActiveDebt := false
	for _, debt := range allDebts {
		if !debt.IsPaid && debt.Remaining > 0 {
			totalRemainingDebt += debt.Remaining
			hasActiveDebt = true
		}
	}

	dsr := 0.0
	isDebtFree := !hasActiveDebt && paidDebtThisMonth == 0

	if isDebtFree {
		dsr = 0.0
	} else {
		monthlyObligation := paidDebtThisMonth
		if monthlyObligation == 0 && hasActiveDebt {
			// Jika belum ada pembayaran cicilan di siklus ini, gunakan estimasi beban 5% dari sisa utang
			monthlyObligation = totalRemainingDebt * 0.05
		}

		if totalIncomeMonth > 0 {
			dsr = monthlyObligation / totalIncomeMonth
		} else if monthlyObligation > 0 {
			dsr = 1.0 // 100% beban utang tanpa pendapatan
		}
	}

	debtRatioStruct := entity.FinancialHealthRatio{
		Name:           "Rasio Utang Terhadap Pendapatan",
		Value:          dsr,
		Target:         "< 30%",
		FormattedValue: fmt.Sprintf("%.1f%%", dsr*100),
	}

	debtScore := 0.0
	if isDebtFree {
		debtScore = 100.0
		debtRatioStruct.Status = entity.StatusHealthy
		debtRatioStruct.Description = "Bebas utang! Kondisi keuangan Anda sangat ideal."
	} else if dsr <= 0.15 {
		debtScore = 100.0
		debtRatioStruct.Status = entity.StatusHealthy
		debtRatioStruct.Description = "Beban cicilan utang sangat ringan dibandingkan penghasilan Anda."
	} else if dsr <= 0.30 {
		debtScore = 80.0 + ((0.30-dsr)/0.15)*20.0
		debtRatioStruct.Status = entity.StatusHealthy
		debtRatioStruct.Description = "Porsi pembayaran utang masih dalam batas aman perencana keuangan (< 30%)."
	} else if dsr <= 0.40 {
		debtScore = 50.0 + ((0.40-dsr)/0.10)*29.0
		debtRatioStruct.Status = entity.StatusWarning
		debtRatioStruct.Description = "Waspada! Beban cicilan utang mendekati batas maksimal penghasilan Anda (30% - 40%)."
	} else {
		scoreVal := 50.0 - ((dsr-0.40)/0.20)*50.0
		if scoreVal < 0 {
			scoreVal = 0
		}
		debtScore = scoreVal
		debtRatioStruct.Status = entity.StatusDanger
		debtRatioStruct.Description = "Bahaya! Beban utang menyerap lebih dari 40% penghasilan Anda. Rentan gagal bayar."
	}

	// 4. OVERALL SCORE CALCULATION (Weighted Continuous Scoring)
	// Bobot: Likuiditas (35%), Tabungan (35%), Utang (30%)
	overallScore := math.Round(savingsScore*0.35 + liquidityPoints*0.35 + debtScore*0.30)
	if overallScore > 100 {
		overallScore = 100
	} else if overallScore < 0 {
		overallScore = 0
	}

	overallStatus := entity.StatusWarning
	if overallScore >= 80 {
		overallStatus = entity.StatusHealthy
	} else if overallScore < 50 {
		overallStatus = entity.StatusDanger
	}

	return entity.FinancialHealthResponse{
		OverallScore:  overallScore,
		OverallStatus: overallStatus,
		Ratios: []entity.FinancialHealthRatio{
			savingsRatio,
			liquidityRatio,
			debtRatioStruct,
		},
	}, nil
}
