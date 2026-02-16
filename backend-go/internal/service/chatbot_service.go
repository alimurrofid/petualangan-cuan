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
	sb.WriteString("\n\n--- DATA KEUANGAN USER (REAL-TIME) ---\n")

	if dashboard, err := s.dashboardSvc.GetDashboardData(userID); err == nil {
		sb.WriteString("\n📊 RINGKASAN BULAN INI:\n")
		sb.WriteString(fmt.Sprintf("- Total Saldo: Rp%.0f\n", dashboard.TotalBalance))
		sb.WriteString(fmt.Sprintf("- Saldo Tersedia: Rp%.0f\n", dashboard.TotalAvailableBalance))
		sb.WriteString(fmt.Sprintf("- Pemasukan Bulan Ini: Rp%.0f\n", dashboard.TotalIncomeMonth))
		sb.WriteString(fmt.Sprintf("- Pengeluaran Bulan Ini: Rp%.0f\n", dashboard.TotalExpenseMonth))
	}

	if wallets, err := s.walletRepo.FindByUserID(userID); err == nil && len(wallets) > 0 {
		sb.WriteString(fmt.Sprintf("\n💳 DAFTAR WALLET (%d):\n", len(wallets)))
		for _, w := range wallets {
			sb.WriteString(fmt.Sprintf("- %s (%s): Rp%.0f\n", w.Name, w.Type, w.Balance))
		}
	}

	if txns, err := s.transactionRepo.GetRecentTransactions(userID, 10); err == nil && len(txns) > 0 {
		sb.WriteString(fmt.Sprintf("\n📝 TRANSAKSI TERAKHIR (%d):\n", len(txns)))
		for _, t := range txns {
			prefix := "🔴"
			if t.Type == "income" {
				prefix = "🟢"
			}
			walletName := ""
			if t.Wallet.Name != "" {
				walletName = t.Wallet.Name
			}
			categoryName := ""
			if t.Category.Name != "" {
				categoryName = t.Category.Name
			}
			sb.WriteString(fmt.Sprintf("- %s %s: Rp%.0f (%s, %s, %s)\n",
				prefix, t.Description, t.Amount, t.Type, walletName, categoryName))
		}
	}

	today := time.Now().Format("2006-01-02")
	if summary, err := s.transactionRepo.FindSummaryByDateRange(userID, today, today, nil, nil, ""); err == nil {
		totalExpenseToday := 0.0
		totalIncomeToday := 0.0
		for _, s := range summary {
			totalExpenseToday += s.Expense
			totalIncomeToday += s.Income
		}
		sb.WriteString(fmt.Sprintf("\n📅 HARI INI (%s):\n", today))
		sb.WriteString(fmt.Sprintf("- Pengeluaran: Rp%.0f\n", totalExpenseToday))
		sb.WriteString(fmt.Sprintf("- Pemasukan: Rp%.0f\n", totalIncomeToday))
	}

	now := time.Now()
	offset := (int(now.Weekday()) + 6) % 7
	startOfWeek := now.AddDate(0, 0, -offset)
	weekStart := startOfWeek.Format("2006-01-02")
	if summary, err := s.transactionRepo.FindSummaryByDateRange(userID, weekStart, today, nil, nil, ""); err == nil {
		totalExpenseWeek := 0.0
		totalIncomeWeek := 0.0
		for _, s := range summary {
			totalExpenseWeek += s.Expense
			totalIncomeWeek += s.Income
		}
		sb.WriteString(fmt.Sprintf("\n📅 MINGGU INI (%s s/d %s):\n", weekStart, today))
		sb.WriteString(fmt.Sprintf("- Pengeluaran: Rp%.0f\n", totalExpenseWeek))
		sb.WriteString(fmt.Sprintf("- Pemasukan: Rp%.0f\n", totalIncomeWeek))
	}

	if debts, err := s.debtRepo.FindByUserID(userID, ""); err == nil && len(debts) > 0 {
		activeDebts := 0
		for _, d := range debts {
			if !d.IsPaid {
				activeDebts++
			}
		}
		if activeDebts > 0 {
			sb.WriteString(fmt.Sprintf("\n💸 UTANG/PIUTANG AKTIF (%d):\n", activeDebts))
			for _, d := range debts {
				if d.IsPaid {
					continue
				}
				typeLabel := "Utang"
				if d.Type == entity.DebtTypeReceivable {
					typeLabel = "Piutang"
				}
				sb.WriteString(fmt.Sprintf("- %s [%s]: Sisa Rp%.0f dari Rp%.0f", d.Name, typeLabel, d.Remaining, d.Amount))
				if d.DueDate != nil {
					sb.WriteString(fmt.Sprintf(" (jatuh tempo: %s)", d.DueDate.Format("02 Jan 2006")))
				}
				sb.WriteString("\n")
			}
		}
	}

	if goals, err := s.savingGoalRepo.FindAll(userID); err == nil && len(goals) > 0 {
		activeGoals := 0
		for _, g := range goals {
			if !g.IsFinished {
				activeGoals++
			}
		}
		if activeGoals > 0 {
			sb.WriteString(fmt.Sprintf("\n🎯 TARGET TABUNGAN AKTIF (%d):\n", activeGoals))
			for _, g := range goals {
				if g.IsFinished {
					continue
				}
				progress := 0.0
				if g.TargetAmount > 0 {
					progress = (g.CurrentAmount / g.TargetAmount) * 100
				}
				status := "⏳"
				if g.IsAchieved {
					status = "✅"
				}
				sb.WriteString(fmt.Sprintf("- %s %s: Rp%.0f / Rp%.0f (%.0f%%)",
					status, g.Name, g.CurrentAmount, g.TargetAmount, progress))
				if g.Deadline != nil {
					sb.WriteString(fmt.Sprintf(" deadline: %s", g.Deadline.Format("02 Jan 2006")))
				}
				sb.WriteString("\n")
			}
		}
	}

	if health, err := s.financialHealth.GetFinancialHealth(userID); err == nil {
		sb.WriteString("\n🏥 KESEHATAN KEUANGAN:\n")
		sb.WriteString(fmt.Sprintf("- Skor: %.0f/100 (%s)\n", health.OverallScore, health.OverallStatus))
		for _, r := range health.Ratios {
			sb.WriteString(fmt.Sprintf("- %s: %s (%s) - %s\n", r.Name, r.FormattedValue, r.Status, r.Description))
		}
	}

	sb.WriteString("\n--- AKHIR DATA KEUANGAN ---")
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
