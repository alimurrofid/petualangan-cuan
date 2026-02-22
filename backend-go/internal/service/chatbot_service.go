package service

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"
	"fmt"
	"math"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

func formatRupiah(amount float64) string {
	n := int64(math.Round(amount))
	if n < 0 {
		return "-" + formatRupiah(-amount)
	}
	str := fmt.Sprintf("%d", n)
	result := make([]byte, 0, len(str)+len(str)/3)
	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result = append(result, '.')
		}
		result = append(result, byte(c))
	}
	return "Rp" + string(result)
}

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

func (s *ChatbotService) GetUserContext(userID uint, message string) string {
	intent := DetectIntent(message)

	if intent == IntentSmallTalk {
		return ""
	}

	var eg errgroup.Group

	var dashboard *entity.DashboardData
	var wallets []entity.Wallet
	var txns []entity.Transaction
	var summaryToday []entity.TransactionSummary
	var summaryWeek []entity.TransactionSummary
	var summaryMonth []entity.TransactionSummary
	var debts []entity.Debt
	var goals []entity.SavingGoal
	var health entity.FinancialHealthResponse

	wib, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(wib)
	today := now.Format("2006-01-02")
	todayEnd := today + " 23:59:59"

	offset := (int(now.Weekday()) + 6) % 7
	startOfWeek := now.AddDate(0, 0, -offset)
	weekStart := startOfWeek.Format("2006-01-02")

	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, wib).Format("2006-01-02")
	endOfMonth := today + " 23:59:59"

	// Dashboard — selalu dibutuhkan (saldo, income, expense bulan ini).
	if needsDashboardContext(intent) {
		eg.Go(func() error {
			if d, err := s.dashboardSvc.GetDashboardData(userID); err == nil {
				dashboard = d
			}
			return nil
		})
	}

	// Wallet — dibutuhkan untuk semua intent transaksi & umum.
	if needsWalletContext(intent) {
		eg.Go(func() error {
			if w, err := s.walletRepo.FindByUserID(userID); err == nil {
				wallets = w
			}
			return nil
		})
	}

	// Transaksi terakhir — selalu berguna untuk transaksi & laporan.
	if intent != IntentDebt && intent != IntentGoal && intent != IntentHealth {
		eg.Go(func() error {
			if t, err := s.transactionRepo.GetRecentTransactions(userID, MaxRecentTxns); err == nil {
				txns = t
			}
			return nil
		})
	}

	// Ringkasan harian & mingguan — hanya untuk report / general / transaksi.
	if needsReportContext(intent) {
		eg.Go(func() error {
			if st, err := s.transactionRepo.FindSummaryByDateRange(userID, today, todayEnd, nil, nil, ""); err == nil {
				summaryToday = st
			}
			return nil
		})
		eg.Go(func() error {
			if sw, err := s.transactionRepo.FindSummaryByDateRange(userID, weekStart, todayEnd, nil, nil, ""); err == nil {
				summaryWeek = sw
			}
			return nil
		})
	}

	// Breakdown per-hari bulan ini — HANYA untuk IntentReport agar AI bisa menjawab
	// pertanyaan detail seperti "hari mana pengeluaran terbanyak?".
	if intent == IntentReport {
		eg.Go(func() error {
			if sm, err := s.transactionRepo.FindSummaryByDateRange(userID, startOfMonth, endOfMonth, nil, nil, ""); err == nil {
				summaryMonth = sm
			}
			return nil
		})
	}

	// Utang / piutang — hanya untuk intent utang & general.
	if needsDebtContext(intent) {
		eg.Go(func() error {
			if d, err := s.debtRepo.FindByUserID(userID, ""); err == nil {
				debts = d
			}
			return nil
		})
	}

	// Target tabungan — hanya untuk intent goal & general.
	if needsGoalContext(intent) {
		eg.Go(func() error {
			if g, err := s.savingGoalRepo.FindAll(userID); err == nil {
				goals = g
			}
			return nil
		})
	}

	// Skor keuangan — hanya untuk intent health & general.
	if needsHealthContext(intent) {
		eg.Go(func() error {
			if h, err := s.financialHealth.GetFinancialHealth(userID); err == nil {
				health = h
			}
			return nil
		})
	}

	eg.Wait()

	var sb strings.Builder
	sb.WriteString("\n--- DATA KEUANGAN USER ---\n")

	// Dashboard summary
	if dashboard != nil {
		sb.WriteString("Total Saldo: " + formatRupiah(dashboard.TotalBalance) + "\n")
		sb.WriteString("Saldo Tersedia: " + formatRupiah(dashboard.TotalAvailableBalance) + "\n")
		sb.WriteString("Pemasukan Bulan Ini: " + formatRupiah(dashboard.TotalIncomeMonth) + "\n")
		sb.WriteString("Pengeluaran Bulan Ini: " + formatRupiah(dashboard.TotalExpenseMonth) + "\n")
	}

	// Daftar wallet — kritis agar AI tahu ke mana transaksi disimpan.
	if len(wallets) > 0 {
		sb.WriteString(fmt.Sprintf("\nDaftar Wallet (%d):\n", len(wallets)))
		for _, w := range wallets {
			sb.WriteString(fmt.Sprintf("- %s (%s): %s\n", w.Name, w.Type, formatRupiah(w.Balance)))
		}
	}

	// Transaksi terakhir (maks MaxRecentTxns)
	if len(txns) > 0 {
		sb.WriteString("\nTransaksi Terakhir:\n")
		for _, t := range txns {
			walletName := t.Wallet.Name
			categoryName := t.Category.Name
			sb.WriteString(fmt.Sprintf("- %s: %s (%s, %s, %s)\n",
				t.Description, formatRupiah(t.Amount), t.Type, walletName, categoryName))
		}
	}

	// Ringkasan hari ini
	if len(summaryToday) > 0 {
		expToday, incToday := 0.0, 0.0
		for _, s := range summaryToday {
			expToday += s.Expense
			incToday += s.Income
		}
		sb.WriteString(fmt.Sprintf("\nHari Ini (%s): Pengeluaran %s, Pemasukan %s\n",
			today, formatRupiah(expToday), formatRupiah(incToday)))
	}

	// Ringkasan minggu ini
	if len(summaryWeek) > 0 {
		expWeek, incWeek := 0.0, 0.0
		for _, s := range summaryWeek {
			expWeek += s.Expense
			incWeek += s.Income
		}
		sb.WriteString(fmt.Sprintf("Minggu Ini (%s s/d %s): Pengeluaran %s, Pemasukan %s\n",
			weekStart, today, formatRupiah(expWeek), formatRupiah(incWeek)))
	}

	// Breakdown per-hari bulan ini — pre-computed agar AI tidak perlu hitung sendiri.
	// LLM lokal tidak handal untuk operasi max/min dari data mentah.
	if len(summaryMonth) > 0 {
		var maxExpDate, minExpDate string
		var maxExp, minExp float64
		var totalExpMonth, totalIncMonth float64
		activeDays := 0
		minExp = -1 // sentinel untuk deteksi hari pertama

		type dayData struct {
			date    string
			expense float64
			income  float64
		}
		var days []dayData

		for _, s := range summaryMonth {
			if s.Expense == 0 && s.Income == 0 {
				continue
			}
			totalExpMonth += s.Expense
			totalIncMonth += s.Income
			activeDays++
			days = append(days, dayData{s.Date, s.Expense, s.Income})

			if s.Expense > maxExp {
				maxExp = s.Expense
				maxExpDate = s.Date
			}
			if minExp < 0 || (s.Expense > 0 && s.Expense < minExp) {
				minExp = s.Expense
				minExpDate = s.Date
			}
		}

		if activeDays > 0 {
			sb.WriteString(fmt.Sprintf("\nAnalitik Bulan Ini (%s s/d %s):\n", startOfMonth, today))
			sb.WriteString(fmt.Sprintf("  Total Pengeluaran: %s (%d hari aktif)\n", formatRupiah(totalExpMonth), activeDays))
			sb.WriteString(fmt.Sprintf("  Total Pemasukan: %s\n", formatRupiah(totalIncMonth)))
			if maxExpDate != "" {
				sb.WriteString(fmt.Sprintf("  Pengeluaran TERBANYAK: %s sebesar %s\n", maxExpDate, formatRupiah(maxExp)))
			}
			if minExpDate != "" && minExpDate != maxExpDate {
				sb.WriteString(fmt.Sprintf("  Pengeluaran TERKECIL: %s sebesar %s\n", minExpDate, formatRupiah(minExp)))
			}

			// Urutkan top-5 hari pengeluaran terbesar (sort sederhana, data kecil)
			for i := 0; i < len(days)-1; i++ {
				for j := i + 1; j < len(days); j++ {
					if days[j].expense > days[i].expense {
						days[i], days[j] = days[j], days[i]
					}
				}
			}
			limit := 5
			if len(days) < limit {
				limit = len(days)
			}
			if limit > 1 {
				sb.WriteString("  Top pengeluaran per-hari:\n")
				for i := 0; i < limit; i++ {
					sb.WriteString(fmt.Sprintf("    %d. %s — %s\n", i+1, days[i].date, formatRupiah(days[i].expense)))
				}
			}
		}
	}

	// Utang / piutang aktif (maks MaxDebtsInContext)
	if len(debts) > 0 {
		count := 0
		hasActive := false
		for _, d := range debts {
			if d.IsPaid || count >= MaxDebtsInContext {
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
			sb.WriteString(fmt.Sprintf("- %s [%s]: Sisa %s dari %s\n",
				d.Name, typeLabel, formatRupiah(d.Remaining), formatRupiah(d.Amount)))
			count++
		}
	}

	// Target tabungan aktif (maks MaxGoalsInContext)
	if len(goals) > 0 {
		count := 0
		hasActive := false
		for _, g := range goals {
			if g.IsFinished || count >= MaxGoalsInContext {
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
			sb.WriteString(fmt.Sprintf("- %s: %s/%s (%.0f%%)\n",
				g.Name, formatRupiah(g.CurrentAmount), formatRupiah(g.TargetAmount), progress))
			count++
		}
	}

	// Skor kesehatan keuangan
	if health.OverallStatus != "" {
		sb.WriteString(fmt.Sprintf("\nSkor Keuangan: %.0f/100 (%s)\n", health.OverallScore, health.OverallStatus))
	}

	sb.WriteString("--- AKHIR DATA ---")

	// Terapkan token budget: potong jika melebihi MaxContextChars.
	result := sb.String()
	if len(result) > MaxContextChars {
		return result[:MaxContextChars] + "\n[...context dipotong karena melebihi batas token]\n--- AKHIR DATA ---"
	}
	return result
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

// resolveWallet mencari wallet berdasarkan nama. Ia memuat ulang daftar wallet dari DB
// hanya sebagai fallback jika list kosong (misalnya dipanggil dari path non-context).
func (s *ChatbotService) resolveWallet(userID uint, name string) (uint, string, error) {
	wallets, err := s.walletRepo.FindByUserID(userID)
	if err != nil {
		return 0, "", err
	}
	return resolveWalletFromList(wallets, name)
}

// resolveWalletFromList mencari wallet dari list yang sudah ada di memori,
// sehingga tidak perlu query DB ulang.
func resolveWalletFromList(wallets []entity.Wallet, name string) (uint, string, error) {
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

// resolveCategory mencari kategori berdasarkan nama dan tipe transaksi.
func (s *ChatbotService) resolveCategory(userID uint, name string, txType string) (uint, string, error) {
	categories, err := s.categoryRepo.FindAll(userID)
	if err != nil {
		return 0, "", err
	}
	return resolveCategoryFromList(categories, name, txType)
}

// resolveCategoryFromList mencari kategori dari list yang sudah ada di memori.
func resolveCategoryFromList(categories []entity.Category, name string, txType string) (uint, string, error) {
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
