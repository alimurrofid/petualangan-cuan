package service

import (
	"bytes"
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"
	"errors"
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"

	"gorm.io/gorm"
)

type TransactionService interface {
	CreateTransaction(userID uint, input CreateTransactionInput) (*entity.Transaction, error)
	GetTransactions(userID uint, params entity.TransactionFilterParams) ([]entity.Transaction, int64, error)
	GetTransaction(id uint, userID uint) (*entity.Transaction, error)
	UpdateTransaction(id uint, userID uint, input CreateTransactionInput) (*entity.Transaction, error)
	DeleteTransaction(id uint, userID uint) error
	TransferTransaction(userID uint, input TransferTransactionInput) error
	GetCalendarData(userID uint, startDate, endDate string, walletID *uint, categoryID *uint, search string) ([]entity.TransactionSummary, error)
	GetReport(userID uint, startDate, endDate string, walletID *uint, filterType *string) ([]entity.CategoryBreakdown, error)
	ExportTransactions(userID uint, params entity.TransactionFilterParams) (*bytes.Buffer, error)
	ExportReport(userID uint, startDate, endDate string, walletID *uint, filterType *string) (*bytes.Buffer, error)
}

type transactionService struct {
	repo       repository.TransactionRepository
	walletRepo repository.WalletRepository
	db         *gorm.DB // Store DB connection to start transactions
}

func NewTransactionService(repo repository.TransactionRepository, walletRepo repository.WalletRepository, db *gorm.DB) TransactionService {
	return &transactionService{
		repo:       repo,
		walletRepo: walletRepo,
		db:         db,
	}
}

type CreateTransactionInput struct {
	WalletID    uint      `json:"wallet_id" binding:"required"`
	CategoryID  uint      `json:"category_id" binding:"required"`
	Amount      float64   `json:"amount" binding:"required"`
	Type        string    `json:"type" binding:"required"` // income, expense
	Description string    `json:"description"`
	Attachment  string    `json:"attachment"`
	Date        time.Time `json:"date" binding:"required"`
}

type TransferTransactionInput struct {
	FromWalletID uint      `json:"from_wallet_id" binding:"required"`
	ToWalletID   uint      `json:"to_wallet_id" binding:"required"`
	Amount       float64   `json:"amount" binding:"required"`
	Description  string    `json:"description"`
	Date         time.Time `json:"date" binding:"required"`
}

func (s *transactionService) CreateTransaction(userID uint, input CreateTransactionInput) (*entity.Transaction, error) {
	// Start DB Transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return nil, tx.Error
	}

	// 1. Check Wallet Ownership
	wallet, err := s.walletRepo.FindByID(input.WalletID, userID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("wallet not found")
	}

	// 2. Create Transaction Record
	transaction := &entity.Transaction{
		UserID:      userID,
		WalletID:    input.WalletID,
		CategoryID:  input.CategoryID,
		Amount:      input.Amount,
		Type:        input.Type,
		Description: input.Description,
		Attachment:  input.Attachment,
		Date:        input.Date,
	}

	// Use repository bound to this transaction
	if err := s.repo.WithTx(tx).Create(transaction); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 3. Update Wallet Balance
	// 3. Update Wallet Balance if NOT saving_allocation
	if input.Type != "saving_allocation" {
		switch input.Type {
		case "income":
			wallet.Balance += input.Amount
		case "expense":
			if wallet.Balance < input.Amount {
				// Optional: Allow negative balance or return error?
			}
			wallet.Balance -= input.Amount
		}
	}

	// Use wallet repository linked to this transaction? 
	// To do that properly we need walletRepo.WithTx too, or just update using tx here.
	// For simplicity, let's assume walletRepo.Update uses its own DB or we pass tx manually.
	// Since WalletRepo implementation is simple `db.Save`, we can temporarily hack it or add WithTx.
	// A cleaner way is to just do:
	if err := tx.Save(wallet).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Fetch full object with preloads to return
	return s.repo.FindByID(transaction.ID, userID)
}

func (s *transactionService) GetTransactions(userID uint, params entity.TransactionFilterParams) ([]entity.Transaction, int64, error) {
	return s.repo.FindAll(userID, params)
}

func (s *transactionService) GetTransaction(id uint, userID uint) (*entity.Transaction, error) {
	return s.repo.FindByID(id, userID)
}

func (s *transactionService) UpdateTransaction(id uint, userID uint, input CreateTransactionInput) (*entity.Transaction, error) {
	// Start DB Transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return nil, tx.Error
	}

	// 1. Fetch Existing Transaction
	t, err := s.repo.FindByID(id, userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// 2. Revert Old Balance
	oldWallet, err := s.walletRepo.WithTx(tx).FindByID(t.WalletID, userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	switch t.Type {
	case "income", "transfer_in":
		oldWallet.Balance -= t.Amount
	case "expense", "transfer_out":
		oldWallet.Balance += t.Amount
	case "saving_allocation":
		// Do nothing to physical balance
	}

	if err := s.walletRepo.WithTx(tx).Update(oldWallet); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 3. Update Transaction Fields
	t.WalletID = input.WalletID
	t.CategoryID = input.CategoryID
	t.Amount = input.Amount
	t.Type = input.Type
	t.Description = input.Description
	// Only update attachment if provided (or handle deletion appropriately in handler)
	// For now, if input has it, update it.
	if input.Attachment != "" {
		t.Attachment = input.Attachment
	}
	t.Date = input.Date

	if err := s.repo.WithTx(tx).Update(t); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 4. Apply New Balance
	// Fetch new wallet (might be same as old)
	newWallet, err := s.walletRepo.WithTx(tx).FindByID(input.WalletID, userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	switch input.Type {
	case "income", "transfer_in":
		newWallet.Balance += input.Amount
	case "expense", "transfer_out":
		newWallet.Balance -= input.Amount
	case "saving_allocation":
		// Do nothing
	}

	if err := s.walletRepo.WithTx(tx).Update(newWallet); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 5. Cascade Update to Related Transaction (if exists)
	if t.RelatedTransactionID != nil {
		relatedID := *t.RelatedTransactionID
		
		var relatedTx entity.Transaction
		if err := tx.Where("id = ?", relatedID).First(&relatedTx).Error; err == nil {
			// Revert old balance effect of related tx
			// Use Transactional Wallet Repo!!
			relatedWallet, err := s.walletRepo.WithTx(tx).FindByID(relatedTx.WalletID, userID)
			if err == nil {
				// Revert OLD amount
				// Revert OLD amount
				switch relatedTx.Type {
				case "transfer_in":
					relatedWallet.Balance -= relatedTx.Amount // Revert add
				case "transfer_out":
					relatedWallet.Balance += relatedTx.Amount // Revert sub
				}
				
				// Apply NEW amount (from input.Amount)
				// Apply NEW amount (from input.Amount)
				switch relatedTx.Type {
				case "transfer_in":
					relatedWallet.Balance += input.Amount
				case "transfer_out":
					relatedWallet.Balance -= input.Amount
				}
				
				if err := s.walletRepo.WithTx(tx).Update(relatedWallet); err != nil {
					tx.Rollback()
					return nil, err
				}
				
				// Update the related transaction record
				relatedTx.Amount = input.Amount
				relatedTx.Date = input.Date
				relatedTx.Description = input.Description 
				
				if err := tx.Save(&relatedTx).Error; err != nil {
					tx.Rollback()
					return nil, err
				}
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return s.repo.FindByID(t.ID, userID)
}

func (s *transactionService) DeleteTransaction(id uint, userID uint) error {
	// Start DB Transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Get Transaction to know amount/wallet
	t, err := s.repo.FindByID(id, userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 2. Get Wallet
	w, err := s.walletRepo.FindByID(t.WalletID, userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 3. Revert Balance
	// 3. Revert Balance
	switch t.Type {
	case "income":
		w.Balance -= t.Amount
	case "expense":
		w.Balance += t.Amount
	case "transfer_in":
		w.Balance -= t.Amount
	case "transfer_out":
		w.Balance += t.Amount
	case "saving_allocation":
		// Do nothing
	}

	if err := tx.Save(w).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 4. Delete Transaction
	if err := tx.Delete(t).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 5. Cascade Delete Related Transaction
	if t.RelatedTransactionID != nil {
		relatedID := *t.RelatedTransactionID
		// Get related
		var relatedTx entity.Transaction
		if err := tx.Where("id = ?", relatedID).First(&relatedTx).Error; err == nil {
			// Revert balance for related
			relatedWallet, err := s.walletRepo.FindByID(relatedTx.WalletID, userID)
			if err == nil {
				switch relatedTx.Type {
				case "transfer_in":
					relatedWallet.Balance -= relatedTx.Amount
				case "transfer_out":
					relatedWallet.Balance += relatedTx.Amount
				}
				if err := tx.Save(relatedWallet).Error; err != nil {
					tx.Rollback()
					return err
				}
				
				// Delete related
				if err := tx.Delete(&relatedTx).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		}
	}

	return tx.Commit().Error
}

func (s *transactionService) TransferTransaction(userID uint, input TransferTransactionInput) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return tx.Error
	}

	// 1. Get Wallets
	fromWallet, err := s.walletRepo.FindByID(input.FromWalletID, userID)
	if err != nil {
		tx.Rollback()
		return errors.New("source wallet not found")
	}

	toWallet, err := s.walletRepo.FindByID(input.ToWalletID, userID)
	if err != nil {
		tx.Rollback()
		return errors.New("destination wallet not found")
	}

	// 2. Resolve Transfer Category
	transferCatID, err := s.getCategoryForTransfer(userID)
	if err != nil {
		// If fails to find/create, we can't proceed cleanly
		tx.Rollback()
		return err
	}

	// 3. Create Outgoing Transaction
	expenseTx := &entity.Transaction{
		UserID:      userID,
		WalletID:    input.FromWalletID,
		CategoryID:  transferCatID,
		Amount:      input.Amount,
		Type:        "transfer_out",
		Description: input.Description,
		Date:        input.Date,
	}

	if err := s.repo.WithTx(tx).Create(expenseTx); err != nil {
		tx.Rollback()
		return err
	}

	fromWallet.Balance -= input.Amount
	if err := tx.Save(fromWallet).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 4. Create Incoming Transaction
	incomeTx := &entity.Transaction{
		UserID:      userID,
		WalletID:    input.ToWalletID,
		CategoryID:  transferCatID,
		Amount:      input.Amount,
		Type:        "transfer_in",
		Description: input.Description,
		Date:        input.Date,
	}

	if err := s.repo.WithTx(tx).Create(incomeTx); err != nil {
		tx.Rollback()
		return err
	}
	
	// Link Transactions
	expenseTx.RelatedTransactionID = &incomeTx.ID
	incomeTx.RelatedTransactionID = &expenseTx.ID
	
	if err := tx.Save(expenseTx).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Save(incomeTx).Error; err != nil {
		tx.Rollback()
		return err
	}

	toWallet.Balance += input.Amount
	if err := tx.Save(toWallet).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// Helper to find or create appropriate category for transfer
func (s *transactionService) getCategoryForTransfer(userID uint) (uint, error) {

	// But let's use the proper struct
	var cat entity.Category

	// Use FirstOrCreate to handle race condition atomicity (requires unique constraint usually, but GORM handles basic check)
	// Ideally we should have a unique index on (user_id, type) where type='transfer' or name='Transfer'
	// Assuming application level unique check is insufficient, but FirstOrCreate does:
	// SELECT * FROM ... LIMIT 1; if not found -> INSERT
	// Inside a transaction, if we set isolation level, it might work, or we rely on DB constraint.
	// For now, using FirstOrCreate is the requested solution.
	
	err := s.db.Where(entity.Category{UserID: userID, Type: "transfer"}).
		Attrs(entity.Category{Name: "Transfer", Icon: "Em_Exchange", BudgetLimit: 0}).
		FirstOrCreate(&cat).Error

	if err != nil {
		return 0, err
	}
	
	return cat.ID, nil
}

func (s *transactionService) GetCalendarData(userID uint, startDate, endDate string, walletID *uint, categoryID *uint, search string) ([]entity.TransactionSummary, error) {
	return s.repo.FindSummaryByDateRange(userID, startDate, endDate, walletID, categoryID, search)
}

func (s *transactionService) GetReport(userID uint, startDate, endDate string, walletID *uint, filterType *string) ([]entity.CategoryBreakdown, error) {
	return s.repo.GetCategoryBreakdown(userID, startDate, endDate, walletID, filterType)
}

func (s *transactionService) ExportTransactions(userID uint, params entity.TransactionFilterParams) (*bytes.Buffer, error) {
	// Force limit to 0 to get all transactions
	params.Limit = 0
	transactions, _, err := s.repo.FindAll(userID, params)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// Create Sheet
	sheetName := "Transactions"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1") // Remove default sheet

	// Header
	headers := []string{"No", "Date", "Description", "Category", "Wallet", "Type", "Amount"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// Style Header
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
	})
	f.SetCellStyle(sheetName, "A1", "G1", style)

	// Rows
	for i, t := range transactions {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), i+1)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), t.Date.Format("2006-01-02"))
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), t.Description)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), t.Category.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), t.Wallet.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), t.Type)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), t.Amount)
	}
	
	f.SetColWidth(sheetName, "B", "B", 12)
	f.SetColWidth(sheetName, "C", "C", 30)
	f.SetColWidth(sheetName, "D", "E", 15)
	f.SetColWidth(sheetName, "F", "F", 10)
	f.SetColWidth(sheetName, "G", "G", 15)

	return f.WriteToBuffer()
}

func (s *transactionService) ExportReport(userID uint, startDate, endDate string, walletID *uint, filterType *string) (*bytes.Buffer, error) {
	data, err := s.GetReport(userID, startDate, endDate, walletID, filterType)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	sheetName := "Report"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	// Header
	headers := []string{"No", "Category", "Type", "Total Amount", "Budget Limit", "Is Over Budget"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// Style
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
	})
	f.SetCellStyle(sheetName, "A1", "F1", style)

	// Rows
	for i, item := range data {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), i+1)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.CategoryName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), item.Type)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), item.TotalAmount)
		
		if item.Type == "expense" && item.BudgetLimit > 0 {
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), item.BudgetLimit)
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), "-")
		}
		
		if item.IsOverBudget {
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), "Yes")
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), "No")
		}
	}

	f.SetColWidth(sheetName, "B", "C", 20)
	f.SetColWidth(sheetName, "D", "E", 20)
	f.SetColWidth(sheetName, "F", "F", 15)

	return f.WriteToBuffer()
}
