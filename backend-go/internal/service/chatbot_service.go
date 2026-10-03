package service

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"
	pkgutils "cuan-backend/pkg/utils"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

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
	debtSvc         DebtService
	savingGoalRepo  repository.SavingGoalRepository
	savingGoalSvc   SavingGoalService
	wishlistRepo    repository.WishlistRepository
	wishlistSvc     WishlistService
	dashboardSvc    DashboardService
	financialHealth FinancialHealthService
	userRepo        repository.UserRepository
}

func NewChatbotService(
	walletRepo repository.WalletRepository,
	categoryRepo repository.CategoryRepository,
	transactionSvc TransactionService,
	transactionRepo repository.TransactionRepository,
	debtRepo repository.DebtRepository,
	debtSvc DebtService,
	savingGoalRepo repository.SavingGoalRepository,
	savingGoalSvc SavingGoalService,
	wishlistRepo repository.WishlistRepository,
	wishlistSvc WishlistService,
	dashboardSvc DashboardService,
	financialHealth FinancialHealthService,
	userRepo repository.UserRepository,
) *ChatbotService {
	return &ChatbotService{
		walletRepo:      walletRepo,
		categoryRepo:    categoryRepo,
		transactionSvc:  transactionSvc,
		transactionRepo: transactionRepo,
		debtRepo:        debtRepo,
		debtSvc:         debtSvc,
		savingGoalRepo:  savingGoalRepo,
		savingGoalSvc:   savingGoalSvc,
		wishlistRepo:    wishlistRepo,
		wishlistSvc:     wishlistSvc,
		dashboardSvc:    dashboardSvc,
		financialHealth: financialHealth,
		userRepo:        userRepo,
	}
}

func (s *ChatbotService) GetUserContext(userID uint, message string) string {
	intent := DetectIntent(message)

	if intent == IntentSmallTalk {
		return ""
	}

	// Resolve payday for dynamic billing cycle
	payday := 1
	if user, err := s.userRepo.FindByID(userID); err == nil && user.Payday != nil {
		payday = *user.Payday
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
	var wishlists []entity.WishlistItem
	var health entity.FinancialHealthResponse

	wib, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(wib)
	today := now.Format("2006-01-02")
	todayEnd := today + " 23:59:59"

	offset := (int(now.Weekday()) + 6) % 7
	startOfWeek := now.AddDate(0, 0, -offset)
	weekStart := startOfWeek.Format("2006-01-02")

	// Dynamic billing cycle
	startCycle, _ := pkgutils.GetBillingCycle(now, payday)
	startOfMonth := startCycle.Format("2006-01-02")
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
	if intent != IntentDebt && intent != IntentGoal && intent != IntentHealth && intent != IntentWishlist {
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

	// Wishlist — hanya untuk intent wishlist & general.
	if needsWishlistContext(intent) && s.wishlistRepo != nil {
		eg.Go(func() error {
			if w, err := s.wishlistRepo.FindAllByUserID(userID); err == nil {
				wishlists = w
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

		// Top Kategori Pengeluaran Bulan Ini
		if len(dashboard.ExpenseBreakdown) > 0 && needsReportContext(intent) {
			sb.WriteString("\nTop Kategori Pengeluaran Bulan Ini:\n")
			limit := 5
			if len(dashboard.ExpenseBreakdown) < limit {
				limit = len(dashboard.ExpenseBreakdown)
			}
			for i := 0; i < limit; i++ {
				cb := dashboard.ExpenseBreakdown[i]
				sb.WriteString(fmt.Sprintf("  %d. %s: %s (%.1f%%)\n", i+1, cb.CategoryName, formatRupiah(cb.TotalAmount), cb.Percentage))
			}
		}
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
			sb.WriteString(fmt.Sprintf("- [ID: %d] %s: %s (%s, %s, %s)\n",
				t.ID, t.Description, formatRupiah(t.Amount), t.Type, walletName, categoryName))
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
			dueDateStr := ""
			if d.DueDate != nil {
				dueFormatted := d.DueDate.In(wib).Format("2006-01-02")
				if d.DueDate.Before(now) {
					dueDateStr = fmt.Sprintf(" [OVERDUE! Jatuh tempo: %s]", dueFormatted)
				} else {
					dueDateStr = fmt.Sprintf(" [Jatuh tempo: %s]", dueFormatted)
				}
			}
			sb.WriteString(fmt.Sprintf("- [ID: %d] %s [%s]: Sisa %s dari %s%s\n",
				d.ID, d.Name, typeLabel, formatRupiah(d.Remaining), formatRupiah(d.Amount), dueDateStr))
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
			sb.WriteString(fmt.Sprintf("- [ID: %d] %s: %s/%s (%.0f%%)\n",
				g.ID, g.Name, formatRupiah(g.CurrentAmount), formatRupiah(g.TargetAmount), progress))
			count++
		}
	}

	// Wishlist aktif (maks MaxWishlistsInContext)
	if len(wishlists) > 0 {
		count := 0
		hasActive := false
		for _, w := range wishlists {
			if w.IsBought || count >= MaxWishlistsInContext {
				continue
			}
			if !hasActive {
				sb.WriteString("\nDaftar Keinginan (Wishlist):\n")
				hasActive = true
			}
			prio := string(w.Priority)
			if prio == "" {
				prio = "medium"
			}
			catName := ""
			if w.Category.Name != "" {
				catName = fmt.Sprintf(", Kategori: %s", w.Category.Name)
			}
			sb.WriteString(fmt.Sprintf("- [ID: %d] %s [%s]: %s%s\n",
				w.ID, w.Name, prio, formatRupiah(w.EstimatedPrice), catName))
			count++
		}
	}

	// Skor kesehatan keuangan
	if health.OverallStatus != "" {
		sb.WriteString(fmt.Sprintf("\nSkor Keuangan: %.0f/100 (%s)\n", health.OverallScore, health.OverallStatus))
		if len(health.Ratios) > 0 && (intent == IntentHealth || intent == IntentGeneral) {
			sb.WriteString("Rincian Rasio Keuangan:\n")
			for _, r := range health.Ratios {
				sb.WriteString(fmt.Sprintf("- %s: %s (Target: %s, Status: %s)\n", r.Name, r.FormattedValue, r.Target, r.Status))
			}
		}
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
	var errs []string

	for _, item := range items {
		action := strings.ToLower(item.Action)
		
		if item.Amount <= 0 && !strings.Contains(action, "delete") && !strings.Contains(action, "update") {
			continue
		}
		
		var saved *entity.SavedTransaction
		var err error
		
		switch action {
		case "create_transaction":
			saved, err = s.saveOne(userID, &item)
		case "update_transaction":
			saved, err = s.updateOne(userID, &item)
		case "delete_transaction":
			saved, err = s.deleteOne(userID, &item)
		case "transfer":
			saved, err = s.transferOne(userID, &item)
		case "pay_debt":
			saved, err = s.payDebtOne(userID, &item)
		case "save_goal":
			saved, err = s.saveGoalOne(userID, &item)
		case "create_wishlist":
			saved, err = s.createWishlistOne(userID, &item)
		case "update_wishlist":
			saved, err = s.updateWishlistOne(userID, &item)
		case "delete_wishlist":
			saved, err = s.deleteWishlistOne(userID, &item)
		case "create_debt":
			saved, err = s.createDebtOne(userID, &item)
		case "update_debt":
			saved, err = s.updateDebtOne(userID, &item)
		case "delete_debt":
			saved, err = s.deleteDebtOne(userID, &item)
		case "create_goal":
			saved, err = s.createGoalOne(userID, &item)
		case "update_goal":
			saved, err = s.updateGoalOne(userID, &item)
		case "delete_goal":
			saved, err = s.deleteGoalOne(userID, &item)
		default:
			err = fmt.Errorf("aksi '%s' tidak didukung", action)
		}

		if err != nil {
			log.Error().Err(err).Str("action", action).Str("description", item.Description).Msg("Transaction item failed")
			errs = append(errs, fmt.Sprintf("- '%s': %v", item.Description, err))
			continue
		}
		
		if saved != nil {
			saved.Action = action
			results = append(results, *saved)
			log.Info().Uint("user_id", userID).Str("action", action).Uint("transaction_id", saved.ID).Msg("AI transaction processed successfully")
		}
	}

	if len(errs) > 0 {
		return results, fmt.Errorf("Beberapa transaksi gagal diproses:\n%s", strings.Join(errs, "\n"))
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

func (s *ChatbotService) updateOne(userID uint, tx *entity.TransactionItemAI) (*entity.SavedTransaction, error) {
	if tx.ID == 0 {
		return nil, errors.New("ID transaksi tidak valid untuk update")
	}

	existingTx, err := s.transactionSvc.GetTransaction(tx.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("transaksi tidak ditemukan: %w", err)
	}

	walletID := existingTx.WalletID
	walletName := existingTx.Wallet.Name
	if tx.WalletName != "" {
		walletID, walletName, err = s.resolveWallet(userID, tx.WalletName)
		if err != nil {
			return nil, fmt.Errorf("wallet '%s' tidak ditemukan: %w", tx.WalletName, err)
		}
	}

	txType := tx.Type
	if txType == "" {
		txType = existingTx.Type
	}

	categoryID := existingTx.CategoryID
	categoryName := existingTx.Category.Name
	if tx.CategoryName != "" {
		categoryID, categoryName, err = s.resolveCategory(userID, tx.CategoryName, txType)
		if err != nil {
			return nil, fmt.Errorf("kategori '%s' tidak ditemukan: %w", tx.CategoryName, err)
		}
	}

	amount := tx.Amount
	if amount <= 0 {
		amount = existingTx.Amount
	}

	desc := tx.Description
	if desc == "" {
		desc = existingTx.Description
	}

	input := CreateTransactionInput{
		WalletID:    walletID,
		CategoryID:  categoryID,
		Amount:      amount,
		Type:        txType,
		Description: desc,
		Date:        existingTx.Date,
	}

	updated, err := s.transactionSvc.UpdateTransaction(tx.ID, userID, input)
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui transaksi: %w", err)
	}

	return &entity.SavedTransaction{
		ID:           updated.ID,
		Description:  desc,
		Amount:       amount,
		Type:         txType,
		CategoryName: categoryName,
		WalletName:   walletName,
	}, nil
}

func (s *ChatbotService) deleteOne(userID uint, tx *entity.TransactionItemAI) (*entity.SavedTransaction, error) {
	if tx.ID == 0 {
		return nil, errors.New("ID transaksi tidak valid untuk delete")
	}

	existingTx, err := s.transactionSvc.GetTransaction(tx.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("transaksi tidak ditemukan: %w", err)
	}
	
	err = s.transactionSvc.DeleteTransaction(tx.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal menghapus transaksi: %w", err)
	}

	return &entity.SavedTransaction{
		ID:           tx.ID,
		Description:  existingTx.Description,
		Amount:       existingTx.Amount,
		Type:         existingTx.Type,
		CategoryName: existingTx.Category.Name,
		WalletName:   existingTx.Wallet.Name,
	}, nil
}

func (s *ChatbotService) transferOne(userID uint, tx *entity.TransactionItemAI) (*entity.SavedTransaction, error) {
	fromWalletID, fromWalletName, err := s.resolveWallet(userID, tx.WalletName)
	if err != nil {
		return nil, fmt.Errorf("dompet asal '%s' tidak ditemukan: %w", tx.WalletName, err)
	}

	toWalletID, toWalletName, err := s.resolveWallet(userID, tx.ToWalletName)
	if err != nil {
		return nil, fmt.Errorf("dompet tujuan '%s' tidak ditemukan: %w", tx.ToWalletName, err)
	}

	if fromWalletID == toWalletID {
		return nil, errors.New("dompet asal dan dompet tujuan tidak boleh sama")
	}

	desc := tx.Description
	if desc == "" {
		desc = fmt.Sprintf("Transfer dari %s ke %s", fromWalletName, toWalletName)
	}

	input := TransferTransactionInput{
		FromWalletID: fromWalletID,
		ToWalletID:   toWalletID,
		Amount:       tx.Amount,
		Description:  desc,
		Date:         time.Now(),
	}

	if err := s.transactionSvc.TransferTransaction(userID, input); err != nil {
		return nil, fmt.Errorf("gagal transfer: %w", err)
	}

	return &entity.SavedTransaction{
		Description:  desc,
		Amount:       tx.Amount,
		Type:         "transfer",
		WalletName:   fromWalletName,
		ToWalletName: toWalletName,
	}, nil
}

func (s *ChatbotService) payDebtOne(userID uint, tx *entity.TransactionItemAI) (*entity.SavedTransaction, error) {
	if s.debtSvc == nil {
		return nil, errors.New("debt service not configured")
	}

	walletID, walletName, err := s.resolveWallet(userID, tx.WalletName)
	if err != nil {
		return nil, fmt.Errorf("wallet '%s' tidak ditemukan: %w", tx.WalletName, err)
	}

	debtID := tx.ID
	debtName := ""
	if debtID == 0 {
		debts, err := s.debtRepo.FindByUserID(userID, "")
		if err == nil {
			descLower := strings.ToLower(tx.Description)
			for _, d := range debts {
				if !d.IsPaid && (strings.Contains(descLower, strings.ToLower(d.Name)) || strings.Contains(strings.ToLower(d.Name), descLower)) {
					debtID = d.ID
					debtName = d.Name
					break
				}
			}
		}
	}

	if debtID == 0 {
		return nil, errors.New("ID utang tidak ditemukan untuk pembayaran")
	}

	debt, err := s.debtSvc.PayDebt(debtID, userID, PayDebtInput{
		WalletID: walletID,
		Amount:   tx.Amount,
		Note:     tx.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("gagal bayar utang: %w", err)
	}
	if debt != nil && debt.Name != "" {
		debtName = debt.Name
	}

	desc := fmt.Sprintf("Bayar Utang %s", debtName)
	if tx.Description != "" && !strings.Contains(tx.Description, debtName) {
		desc += " (" + tx.Description + ")"
	}

	return &entity.SavedTransaction{
		ID:           debtID,
		Description:  desc,
		Amount:       tx.Amount,
		Type:         "expense",
		CategoryName: "Utang",
		WalletName:   walletName,
	}, nil
}

func (s *ChatbotService) saveGoalOne(userID uint, tx *entity.TransactionItemAI) (*entity.SavedTransaction, error) {
	if s.savingGoalSvc == nil {
		return nil, errors.New("saving goal service not configured")
	}

	walletID, walletName, err := s.resolveWallet(userID, tx.WalletName)
	if err != nil {
		return nil, fmt.Errorf("wallet '%s' tidak ditemukan: %w", tx.WalletName, err)
	}

	goalID := tx.ID
	goalName := ""
	if goalID == 0 {
		goals, err := s.savingGoalRepo.FindAll(userID)
		if err == nil {
			descLower := strings.ToLower(tx.Description)
			for _, g := range goals {
				if !g.IsFinished && (strings.Contains(descLower, strings.ToLower(g.Name)) || strings.Contains(strings.ToLower(g.Name), descLower)) {
					goalID = g.ID
					goalName = g.Name
					break
				}
			}
		}
	}

	if goalID == 0 {
		return nil, errors.New("ID target tabungan tidak ditemukan")
	}

	desc := tx.Description
	if desc == "" {
		if goalName != "" {
			desc = "Setor Tabungan " + goalName
		} else {
			desc = "Setor Tabungan"
		}
	}

	contrib, err := s.savingGoalSvc.AddContribution(userID, goalID, ContributionInput{
		WalletID:    walletID,
		Amount:      tx.Amount,
		Date:        time.Now(),
		Description: desc,
	})
	if err != nil {
		return nil, fmt.Errorf("gagal setor tabungan: %w", err)
	}

	return &entity.SavedTransaction{
		ID:           contrib.ID,
		Description:  desc,
		Amount:       tx.Amount,
		Type:         "expense",
		CategoryName: "Tabungan",
		WalletName:   walletName,
	}, nil
}

func (s *ChatbotService) createWishlistOne(userID uint, tx *entity.TransactionItemAI) (*entity.SavedTransaction, error) {
	if s.wishlistSvc == nil {
		return nil, errors.New("wishlist service not configured")
	}

	categoryID, categoryName, err := s.resolveCategory(userID, tx.CategoryName, "expense")
	if err != nil {
		return nil, fmt.Errorf("kategori '%s' tidak ditemukan: %w", tx.CategoryName, err)
	}

	prio := strings.ToLower(tx.Priority)
	if prio != "low" && prio != "medium" && prio != "high" {
		prio = "medium"
	}

	name := tx.Description
	if name == "" {
		name = "Item Wishlist"
	}

	req := &StoreWishlistRequest{
		CategoryID:     categoryID,
		Name:           name,
		EstimatedPrice: tx.Amount,
		Priority:       prio,
	}

	if err := s.wishlistSvc.Create(userID, req); err != nil {
		return nil, fmt.Errorf("gagal menambahkan ke wishlist: %w", err)
	}

	return &entity.SavedTransaction{
		Description:  fmt.Sprintf("%s [%s]", name, prio),
		Amount:       tx.Amount,
		Type:         "wishlist",
		CategoryName: categoryName,
	}, nil
}

func (s *ChatbotService) createDebtOne(userID uint, tx *entity.TransactionItemAI) (*entity.SavedTransaction, error) {
	walletID, walletName, err := s.resolveWallet(userID, tx.WalletName)
	if err != nil {
		return nil, fmt.Errorf("wallet '%s' tidak ditemukan: %w", tx.WalletName, err)
	}

	debtType := "debt"
	if tx.Type == "receivable" {
		debtType = "receivable"
	}

	desc := tx.Description
	if desc == "" {
		if debtType == "debt" {
			desc = "Utang Baru"
		} else {
			desc = "Piutang Baru"
		}
	}

	input := CreateDebtInput{
		WalletID:    walletID,
		Name:        desc,
		Amount:      tx.Amount,
		Type:        debtType,
		Description: "Dicatat oleh Cuan AI",
	}

	created, err := s.debtSvc.CreateDebt(userID, input)
	if err != nil {
		return nil, fmt.Errorf("gagal mencatat utang/piutang: %w", err)
	}

	return &entity.SavedTransaction{
		ID:           created.ID,
		Description:  desc,
		Amount:       tx.Amount,
		Type:         debtType,
		CategoryName: "Utang/Piutang",
		WalletName:   walletName,
	}, nil
}

func (s *ChatbotService) createGoalOne(userID uint, tx *entity.TransactionItemAI) (*entity.SavedTransaction, error) {
	name := tx.Description
	if name == "" {
		name = "Target Tabungan"
	}

	categoryID, _, err := s.resolveCategory(userID, "Tabungan", "expense")
	if err != nil {
		// Fallback jika tidak ada kategori "Tabungan"
		categoryID, _, _ = s.resolveCategory(userID, "Lainnya", "expense")
	}

	input := CreateGoalInput{
		Name:         name,
		TargetAmount: tx.Amount,
		CategoryID:   categoryID,
	}

	created, err := s.savingGoalSvc.CreateGoal(userID, input)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat target tabungan: %w", err)
	}

	return &entity.SavedTransaction{
		ID:           created.ID,
		Description:  name,
		Amount:       tx.Amount,
		Type:         "goal",
		CategoryName: "Target Tabungan",
		WalletName:   "-",
	}, nil
}

// updateGoalOne
func (s *ChatbotService) updateGoalOne(userID uint, tx *entity.TransactionItemAI) (*entity.SavedTransaction, error) {
	if tx.ID == 0 {
		return nil, errors.New("ID target tabungan tidak valid untuk update")
	}

	existing, err := s.savingGoalRepo.FindByID(tx.ID, userID)
	if err != nil || existing == nil {
		return nil, fmt.Errorf("target tabungan tidak ditemukan: %w", err)
	}

	name := tx.Description
	if name == "" {
		name = existing.Name
	}

	amount := tx.Amount
	if amount <= 0 {
		amount = existing.TargetAmount
	}

	categoryID := existing.CategoryID
	if categoryID == 0 {
		categoryID, _, _ = s.resolveCategory(userID, "Tabungan", "expense")
	}

	input := CreateGoalInput{
		Name:         name,
		TargetAmount: amount,
		CategoryID:   categoryID,
		Deadline:     existing.Deadline,
		Icon:         existing.Icon,
	}

	created, err := s.savingGoalSvc.UpdateGoal(userID, tx.ID, input)
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui target tabungan: %w", err)
	}

	return &entity.SavedTransaction{
		ID:           created.ID,
		Description:  name,
		Amount:       amount,
		Type:         "goal",
		CategoryName: "Target Tabungan",
		WalletName:   "-",
	}, nil
}

// deleteGoalOne
func (s *ChatbotService) deleteGoalOne(userID uint, tx *entity.TransactionItemAI) (*entity.SavedTransaction, error) {
	if tx.ID == 0 {
		return nil, errors.New("ID target tabungan tidak valid untuk delete")
	}

	existing, err := s.savingGoalRepo.FindByID(tx.ID, userID)
	if err != nil || existing == nil {
		return nil, fmt.Errorf("target tabungan tidak ditemukan: %w", err)
	}

	err = s.savingGoalSvc.DeleteGoal(userID, tx.ID)
	if err != nil {
		return nil, fmt.Errorf("gagal menghapus target tabungan: %w", err)
	}

	return &entity.SavedTransaction{
		ID:           tx.ID,
		Description:  existing.Name,
		Amount:       existing.TargetAmount,
		Type:         "goal",
		CategoryName: "Target Tabungan",
		WalletName:   "-",
	}, nil
}

// updateDebtOne
func (s *ChatbotService) updateDebtOne(userID uint, tx *entity.TransactionItemAI) (*entity.SavedTransaction, error) {
	if tx.ID == 0 {
		return nil, errors.New("ID utang/piutang tidak valid untuk update")
	}

	existing, err := s.debtRepo.FindByID(tx.ID, userID)
	if err != nil || existing == nil {
		return nil, fmt.Errorf("utang/piutang tidak ditemukan: %w", err)
	}

	walletID := existing.WalletID
	walletName := ""
	if tx.WalletName != "" {
		walletID, walletName, err = s.resolveWallet(userID, tx.WalletName)
		if err != nil {
			return nil, fmt.Errorf("wallet '%s' tidak ditemukan: %w", tx.WalletName, err)
		}
	} else {
		if wallet, err := s.walletRepo.FindByID(existing.WalletID, userID); err == nil && wallet != nil {
			walletName = wallet.Name
		}
	}

	name := tx.Description
	if name == "" {
		name = existing.Name
	}

	amount := tx.Amount
	if amount <= 0 {
		amount = existing.Amount
	}

	input := UpdateDebtInput{
		WalletID:    walletID,
		Name:        name,
		Amount:      amount,
		Description: existing.Description,
		DueDate:     existing.DueDate,
	}

	created, err := s.debtSvc.UpdateDebt(tx.ID, userID, input)
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui utang/piutang: %w", err)
	}

	return &entity.SavedTransaction{
		ID:           created.ID,
		Description:  name,
		Amount:       amount,
		Type:         string(existing.Type),
		CategoryName: "Utang/Piutang",
		WalletName:   walletName,
	}, nil
}

// deleteDebtOne
func (s *ChatbotService) deleteDebtOne(userID uint, tx *entity.TransactionItemAI) (*entity.SavedTransaction, error) {
	if tx.ID == 0 {
		return nil, errors.New("ID utang/piutang tidak valid untuk delete")
	}

	existing, err := s.debtRepo.FindByID(tx.ID, userID)
	if err != nil || existing == nil {
		return nil, fmt.Errorf("utang/piutang tidak ditemukan: %w", err)
	}

	err = s.debtSvc.DeleteDebt(tx.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal menghapus utang/piutang: %w", err)
	}

	return &entity.SavedTransaction{
		ID:           tx.ID,
		Description:  existing.Name,
		Amount:       existing.Amount,
		Type:         string(existing.Type),
		CategoryName: "Utang/Piutang",
		WalletName:   "-",
	}, nil
}

// updateWishlistOne
func (s *ChatbotService) updateWishlistOne(userID uint, tx *entity.TransactionItemAI) (*entity.SavedTransaction, error) {
	if tx.ID == 0 {
		return nil, errors.New("ID wishlist tidak valid untuk update")
	}

	existing, err := s.wishlistRepo.FindByID(tx.ID, userID)
	if err != nil || existing == nil {
		return nil, fmt.Errorf("wishlist tidak ditemukan: %w", err)
	}

	categoryID := existing.CategoryID
	categoryName := ""
	if tx.CategoryName != "" {
		categoryID, categoryName, err = s.resolveCategory(userID, tx.CategoryName, "expense")
		if err != nil {
			return nil, fmt.Errorf("kategori '%s' tidak ditemukan: %w", tx.CategoryName, err)
		}
	}

	prio := tx.Priority
	if prio == "" {
		prio = string(existing.Priority)
	}
	if prio == "" {
		prio = "medium"
	}

	name := tx.Description
	if name == "" {
		name = existing.Name
	}

	amount := tx.Amount
	if amount <= 0 {
		amount = existing.EstimatedPrice
	}

	req := &StoreWishlistRequest{
		CategoryID:     categoryID,
		Name:           name,
		EstimatedPrice: amount,
		Priority:       prio,
	}

	if err := s.wishlistSvc.Update(tx.ID, userID, req); err != nil {
		return nil, fmt.Errorf("gagal memperbarui wishlist: %w", err)
	}

	return &entity.SavedTransaction{
		ID:           tx.ID,
		Description:  fmt.Sprintf("%s [%s]", name, prio),
		Amount:       amount,
		Type:         "wishlist",
		CategoryName: categoryName,
		WalletName:   "-",
	}, nil
}

// deleteWishlistOne
func (s *ChatbotService) deleteWishlistOne(userID uint, tx *entity.TransactionItemAI) (*entity.SavedTransaction, error) {
	if tx.ID == 0 {
		return nil, errors.New("ID wishlist tidak valid untuk delete")
	}

	existing, err := s.wishlistRepo.FindByID(tx.ID, userID)
	if err != nil || existing == nil {
		return nil, fmt.Errorf("wishlist tidak ditemukan: %w", err)
	}

	err = s.wishlistSvc.Delete(tx.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal menghapus wishlist: %w", err)
	}

	return &entity.SavedTransaction{
		ID:           tx.ID,
		Description:  existing.Name,
		Amount:       existing.EstimatedPrice,
		Type:         "wishlist",
		CategoryName: "Wishlist",
		WalletName:   "-",
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
