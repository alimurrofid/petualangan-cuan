package service

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"
	"fmt"
	"math"
	"time"
)

type FinancialHealthService interface {
	GetFinancialHealth(userID uint) (entity.FinancialHealthResponse, error)
}

type financialHealthService struct {
	transactionRepo repository.TransactionRepository
	walletRepo      repository.WalletRepository
	debtRepo        repository.DebtRepository
}

func NewFinancialHealthService(
	transactionRepo repository.TransactionRepository,
	walletRepo repository.WalletRepository,
	debtRepo repository.DebtRepository,
) FinancialHealthService {
	return &financialHealthService{
		transactionRepo: transactionRepo,
		walletRepo:      walletRepo,
		debtRepo:        debtRepo,
	}
}

func (s *financialHealthService) GetFinancialHealth(userID uint) (entity.FinancialHealthResponse, error) {
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	
	startDate := firstOfMonth.Format("2006-01-02")
	endDate := lastOfMonth.Format("2006-01-02")

	// --- 1. Savings Rate ---
	// Formula: (Total Income - Total Expense) / Total Income (Current Month)
	summary, err := s.transactionRepo.FindSummaryByDateRange(userID, startDate, endDate, nil, nil, "")
	if err != nil {
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
	}

    // Cap at 0 if negative savings (spending > income)
    // Actually finding negative savings rate is valid (debt increase), but visually we might handle it.
    // Let's keep the raw value but handle status.

	savingsRatio := entity.FinancialHealthRatio{
		Name:   "Savings Rate",
		Value:  savingsRate,
		Target: "> 20%",
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
		savingsRatio.Description = "Hati-hati, tabungan Anda terlalu sedikit (atau minus)."
	}

	// --- 2. Liquidity Ratio (Emergency Fund) ---
	// Formula: Total Wallet Balance / Avg Monthly Expense (Last 3 Months)
	wallets, err := s.walletRepo.FindByUserID(userID)
	if err != nil {
		return entity.FinancialHealthResponse{}, err
	}

	totalWalletBalance := 0.0
	for _, w := range wallets {
		totalWalletBalance += w.Balance
	}

	// Get 3 previous months
	startOf3MonthsAgo := firstOfMonth.AddDate(0, -3, 0)
	endOfLastMonth := firstOfMonth.AddDate(0, 0, -1) // End of previous month
    
    // If user is new, they might not have 3 months of data. Handle gracefully.
	trend, err := s.transactionRepo.GetMonthlyTrend(userID, startOf3MonthsAgo.Format("2006-01-02"), endOfLastMonth.Format("2006-01-02"))
	if err != nil {
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
    
    // If no past data, use current month expense if available project it (risky), or just use 0 -> Infinite
    avgMonthlyExpense := 0.0
    if monthsCount > 0 {
        avgMonthlyExpense = totalExpense3Months / float64(monthsCount)
    } else if totalExpenseMonth > 0 {
        // Fallback to current month if no history
        avgMonthlyExpense = totalExpenseMonth
    }

	liquidityScore := 0.0 // In Months
	if avgMonthlyExpense > 0 {
		liquidityScore = totalWalletBalance / avgMonthlyExpense
	} else if totalWalletBalance > 0 {
        liquidityScore = 999 // Infinite liquidity
    }
    
    // Cap strictly for display if needed, but raw value is better
    
	liquidityRatio := entity.FinancialHealthRatio{
		Name:   "Dana Darurat",
		Value:  liquidityScore,
		Target: "3 - 6 Bulan",
        FormattedValue: fmt.Sprintf("%.1f Bulan", liquidityScore),
	}

	if liquidityScore >= 3 && liquidityScore <= 12 { // Allow up to 12 months as healthy
		liquidityRatio.Status = entity.StatusHealthy
		liquidityRatio.Description = "Dana darurat Anda aman untuk menutupi pengeluaran mendadak."
	} else if liquidityScore >= 1 {
		liquidityRatio.Status = entity.StatusWarning
		liquidityRatio.Description = "Dana darurat ada, namun perlu ditingkatkan untuk keamanan ekstra."
	} else if liquidityScore > 12 {
        // Too much cash is also "Warning" in investment terms but here acts as Healthy/Passive
        liquidityRatio.Status = entity.StatusHealthy 
        liquidityRatio.Description = "Dana darurat sangat berlimpah."
    } else {
		liquidityRatio.Status = entity.StatusDanger
		liquidityRatio.Description = "Bahaya! Segera sisihkan uang untuk dana darurat minimal 1 bulan pengeluaran."
	}

	// --- 3. Debt-to-Income Ratio ---
	// Formula: Total Debt Installments / Total Income (Current Month)
	debtPayments, err := s.debtRepo.GetTotalPayments(userID, startDate, endDate)
	if err != nil {
		return entity.FinancialHealthResponse{}, err
	}

	dtiRatio := 0.0
	if totalIncomeMonth > 0 {
		dtiRatio = debtPayments / totalIncomeMonth
	}

	dtiRatioStruct := entity.FinancialHealthRatio{
		Name:   "Debt-to-Income",
		Value:  dtiRatio,
		Target: "< 35%",
        FormattedValue: fmt.Sprintf("%.1f%%", dtiRatio*100),
	}

	if dtiRatio == 0 {
		dtiRatioStruct.Status = entity.StatusHealthy
		dtiRatioStruct.Description = "Bebas utang! Kondisi yang sangat ideal."
	} else if dtiRatio <= 0.35 {
		dtiRatioStruct.Status = entity.StatusHealthy
		dtiRatioStruct.Description = "Porsi utang masih dalam batas aman."
	} else if dtiRatio <= 0.50 {
		dtiRatioStruct.Status = entity.StatusWarning
		dtiRatioStruct.Description = "Waspada, utang mulai memakan porsi besar pendapatan Anda."
	} else {
		dtiRatioStruct.Status = entity.StatusDanger
		dtiRatioStruct.Description = "Bahaya! Utang Anda sudah melebih batas wajar (over-leveraged)."
	}

	// --- Overall Score Calculation ---
    // Simple point system: Healthy=100, Warning=50, Danger=0
    score := 0.0
    
    // Savings
    if savingsRatio.Status == entity.StatusHealthy { score += 100 }
    if savingsRatio.Status == entity.StatusWarning { score += 50 }
    
    // Liquidity
    if liquidityRatio.Status == entity.StatusHealthy { score += 100 }
    if liquidityRatio.Status == entity.StatusWarning { score += 50 }
    
    // Debt
    if dtiRatioStruct.Status == entity.StatusHealthy { score += 100 }
    if dtiRatioStruct.Status == entity.StatusWarning { score += 50 }
    
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
			dtiRatioStruct,
		},
	}, nil
}
