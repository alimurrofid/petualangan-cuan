package service

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"
	"fmt"
	"strings"
	"time"
)

type ChatbotService struct {
	walletRepo      repository.WalletRepository
	categoryRepo    repository.CategoryRepository
	transactionSvc  TransactionService
	transactionRepo repository.TransactionRepository
	debtRepo        repository.DebtRepository
	savingGoalRepo  repository.SavingGoalRepository
	dashboardSvc    DashboardService
	financialHealth FinancialHealthService
}

func NewChatbotService(
	walletRepo repository.WalletRepository,
	categoryRepo repository.CategoryRepository,
	transactionSvc TransactionService,
	transactionRepo repository.TransactionRepository,
	debtRepo repository.DebtRepository,
	savingGoalRepo repository.SavingGoalRepository,
	dashboardSvc DashboardService,
	financialHealth FinancialHealthService,
) *ChatbotService {
	return &ChatbotService{
		walletRepo:      walletRepo,
		categoryRepo:    categoryRepo,
		transactionSvc:  transactionSvc,
		transactionRepo: transactionRepo,
		debtRepo:        debtRepo,
		savingGoalRepo:  savingGoalRepo,
		dashboardSvc:    dashboardSvc,
		financialHealth: financialHealth,
	}
}

func (s *ChatbotService) GetUserContext(userID uint) string {
	var sb strings.Builder
	sb.WriteString("\n--- DATA KEUANGAN USER ---\n")

	// Dashboard summary
	if dashboard, err := s.dashboardSvc.GetDashboardData(userID); err == nil {
		sb.WriteString(fmt.Sprintf("Total Saldo: Rp%.0f\n", dashboard.TotalBalance))
		sb.WriteString(fmt.Sprintf("Saldo Tersedia: Rp%.0f\n", dashboard.TotalAvailableBalance))
		sb.WriteString(fmt.Sprintf("Pemasukan Bulan Ini: Rp%.0f\n", dashboard.TotalIncomeMonth))
		sb.WriteString(fmt.Sprintf("Pengeluaran Bulan Ini: Rp%.0f\n", dashboard.TotalExpenseMonth))
	}

	// Wallets with type info (critical for AI to match wallet names)
	if wallets, err := s.walletRepo.FindByUserID(userID); err == nil && len(wallets) > 0 {
		sb.WriteString(fmt.Sprintf("\nDaftar Wallet (%d):\n", len(wallets)))
		for _, w := range wallets {
			sb.WriteString(fmt.Sprintf("- %s (%s): Rp%.0f\n", w.Name, w.Type, w.Balance))
		}
	}

	// Recent transactions (5 items, with category for pattern recognition)
	if txns, err := s.transactionRepo.GetRecentTransactions(userID, 5); err == nil && len(txns) > 0 {
		sb.WriteString("\nTransaksi Terakhir:\n")
		for _, t := range txns {
			walletName := ""
			if t.Wallet.Name != "" {
				walletName = t.Wallet.Name
			}
			categoryName := ""
			if t.Category.Name != "" {
				categoryName = t.Category.Name
			}
			sb.WriteString(fmt.Sprintf("- %s: Rp%.0f (%s, %s, %s)\n",
				t.Description, t.Amount, t.Type, walletName, categoryName))
		}
	}

	// Use WIB timezone for date calculations (container may run in UTC)
	wib, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(wib)
	today := now.Format("2006-01-02")
	todayEnd := today + " 23:59:59"

	// Today summary
	if summary, err := s.transactionRepo.FindSummaryByDateRange(userID, today, todayEnd, nil, nil, ""); err == nil {
		expToday, incToday := 0.0, 0.0
		for _, s := range summary {
			expToday += s.Expense
			incToday += s.Income
		}
		sb.WriteString(fmt.Sprintf("\nHari Ini (%s): Pengeluaran Rp%.0f, Pemasukan Rp%.0f\n", today, expToday, incToday))
	}

	// Weekly summary
	offset := (int(now.Weekday()) + 6) % 7
	startOfWeek := now.AddDate(0, 0, -offset)
	weekStart := startOfWeek.Format("2006-01-02")
	if summary, err := s.transactionRepo.FindSummaryByDateRange(userID, weekStart, todayEnd, nil, nil, ""); err == nil {
		expWeek, incWeek := 0.0, 0.0
		for _, s := range summary {
			expWeek += s.Expense
			incWeek += s.Income
		}
		sb.WriteString(fmt.Sprintf("Minggu Ini (%s s/d %s): Pengeluaran Rp%.0f, Pemasukan Rp%.0f\n", weekStart, today, expWeek, incWeek))
	}
	if debts, err := s.debtRepo.FindByUserID(userID, ""); err == nil && len(debts) > 0 {
		hasActive := false
		for _, d := range debts {
			if d.IsPaid {
				continue
			}
			if !hasActive {
				sb.WriteString("\nUtang/Piutang Aktif:\n")
				hasActive = true
			}
			typeLabel := "Utang"
			if d.Type == entity.DebtTypeReceivable {
				typeLabel = "Piutang"
			}
			sb.WriteString(fmt.Sprintf("- %s [%s]: Sisa Rp%.0f dari Rp%.0f\n", d.Name, typeLabel, d.Remaining, d.Amount))
		}
	}

	// Saving goals
	if goals, err := s.savingGoalRepo.FindAll(userID); err == nil && len(goals) > 0 {
		hasActive := false
		for _, g := range goals {
			if g.IsFinished {
				continue
			}
			if !hasActive {
				sb.WriteString("\nTarget Tabungan:\n")
				hasActive = true
			}
			progress := 0.0
			if g.TargetAmount > 0 {
				progress = (g.CurrentAmount / g.TargetAmount) * 100
			}
			sb.WriteString(fmt.Sprintf("- %s: Rp%.0f/Rp%.0f (%.0f%%)\n", g.Name, g.CurrentAmount, g.TargetAmount, progress))
		}
	}

	// Financial health score only
	if health, err := s.financialHealth.GetFinancialHealth(userID); err == nil {
		sb.WriteString(fmt.Sprintf("\nSkor Keuangan: %.0f/100 (%s)\n", health.OverallScore, health.OverallStatus))
	}

	sb.WriteString("--- AKHIR DATA ---")
	return sb.String()
}

func (s *ChatbotService) SaveTransactions(userID uint, items []entity.TransactionItemAI) ([]entity.SavedTransaction, error) {
	var results []entity.SavedTransaction
	for _, item := range items {
		if item.Amount <= 0 {
			continue
		}
		saved, err := s.saveOne(userID, &item)
		if err != nil {
			fmt.Printf("[ERROR] SaveTransaction item '%s' failed: %v\n", item.Description, err)
			continue
		}
		results = append(results, *saved)
	}
	return results, nil
}

func (s *ChatbotService) saveOne(userID uint, tx *entity.TransactionItemAI) (*entity.SavedTransaction, error) {
	walletID, walletName, err := s.resolveWallet(userID, tx.WalletName)
	if err != nil {
		return nil, fmt.Errorf("wallet '%s' tidak ditemukan: %w", tx.WalletName, err)
	}

	categoryID, categoryName, err := s.resolveCategory(userID, tx.CategoryName, tx.Type)
	if err != nil {
		return nil, fmt.Errorf("kategori '%s' tidak ditemukan: %w", tx.CategoryName, err)
	}

	input := CreateTransactionInput{
		WalletID:    walletID,
		CategoryID:  categoryID,
		Amount:      tx.Amount,
		Type:        tx.Type,
		Description: tx.Description,
		Date:        time.Now(),
	}

	created, err := s.transactionSvc.CreateTransaction(userID, input)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat transaksi: %w", err)
	}

	return &entity.SavedTransaction{
		ID:           created.ID,
		Description:  tx.Description,
		Amount:       tx.Amount,
		Type:         tx.Type,
		CategoryName: categoryName,
		WalletName:   walletName,
	}, nil
}

func (s *ChatbotService) resolveWallet(userID uint, name string) (uint, string, error) {
	wallets, err := s.walletRepo.FindByUserID(userID)
	if err != nil {
		return 0, "", err
	}

	nameLower := strings.ToLower(strings.TrimSpace(name))
	if nameLower == "" {
		nameLower = "tunai"
	}

	for _, w := range wallets {
		if strings.ToLower(w.Name) == nameLower {
			return w.ID, w.Name, nil
		}
	}

	for _, w := range wallets {
		wLower := strings.ToLower(w.Name)
		if strings.Contains(wLower, nameLower) || strings.Contains(nameLower, wLower) {
			return w.ID, w.Name, nil
		}
	}

	if len(wallets) > 0 {
		return wallets[0].ID, wallets[0].Name, nil
	}

	return 0, "", fmt.Errorf("user has no wallets")
}

func (s *ChatbotService) resolveCategory(userID uint, name string, txType string) (uint, string, error) {
	categories, err := s.categoryRepo.FindAll(userID)
	if err != nil {
		return 0, "", err
	}

	nameLower := strings.ToLower(strings.TrimSpace(name))
	if nameLower == "" {
		nameLower = "lainnya"
	}

	var filtered []entity.Category
	for _, c := range categories {
		if strings.ToLower(c.Type) == txType {
			filtered = append(filtered, c)
		}
	}
	if len(filtered) == 0 {
		filtered = categories
	}

	for _, c := range filtered {
		if strings.ToLower(c.Name) == nameLower {
			return c.ID, c.Name, nil
		}
	}

	for _, c := range filtered {
		cLower := strings.ToLower(c.Name)
		if strings.Contains(cLower, nameLower) || strings.Contains(nameLower, cLower) {
			return c.ID, c.Name, nil
		}
	}

	if len(filtered) > 0 {
		return filtered[0].ID, filtered[0].Name, nil
	}

	return 0, "", fmt.Errorf("user has no categories")
}
