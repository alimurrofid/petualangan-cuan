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
	GetReport(userID uint, startDate, endDate string, walletIDs []uint, filterType *string) ([]entity.CategoryBreakdown, error)
	ExportTransactions(userID uint, params entity.TransactionFilterParams) (*bytes.Buffer, error)
	ExportReport(userID uint, startDate, endDate string, walletIDs []uint, filterType *string) (*bytes.Buffer, error)
}

type transactionService struct {
	repo       repository.TransactionRepository
	walletRepo repository.WalletRepository
	db         *gorm.DB
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
	Type        string    `json:"type" binding:"required"`
	Description string    `json:"description"`
	Attachment  string    `json:"attachment"`
	Date        time.Time `json:"date" binding:"required"`
}

type TransferTransactionInput struct {
	FromWalletID uint      `json:"from_wallet_id" binding:"required"`
	ToWalletID   uint      `json:"to_wallet_id" binding:"required"`
	Amount       float64   `json:"amount" binding:"required"`
	TransferFee  float64   `json:"transfer_fee"`
	Description  string    `json:"description"`
	Date         time.Time `json:"date" binding:"required"`
}

func (s *transactionService) CreateTransaction(userID uint, input CreateTransactionInput) (*entity.Transaction, error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return nil, tx.Error
	}

	wallet, err := s.walletRepo.FindByID(input.WalletID, userID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("wallet not found")
	}

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

	if err := s.repo.WithTx(tx).Create(transaction); err != nil {
		tx.Rollback()
		return nil, err
	}

	if input.Type != "saving_allocation" {
		switch input.Type {
		case "income":
			wallet.Balance += input.Amount
		case "expense":
			if wallet.Balance < input.Amount {
				tx.Rollback()
				return nil, errors.New("insufficient wallet balance")
			}
			wallet.Balance -= input.Amount
		}
	}

	if err := tx.Save(wallet).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return s.repo.FindByID(transaction.ID, userID)
}

func (s *transactionService) GetTransactions(userID uint, params entity.TransactionFilterParams) ([]entity.Transaction, int64, error) {
	return s.repo.FindAll(userID, params)
}

func (s *transactionService) GetTransaction(id uint, userID uint) (*entity.Transaction, error) {
	return s.repo.FindByID(id, userID)
}

func (s *transactionService) UpdateTransaction(id uint, userID uint, input CreateTransactionInput) (*entity.Transaction, error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return nil, tx.Error
	}

	t, err := s.repo.FindByID(id, userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	oldWallet, err := s.walletRepo.WithTx(tx).FindByID(t.WalletID, userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	switch t.Type {
	case "income", "transfer_in":
		if oldWallet.Balance < t.Amount {
			tx.Rollback()
			return nil, errors.New("insufficient wallet balance to revert income")
		}
		oldWallet.Balance -= t.Amount
	case "expense", "transfer_out":
		oldWallet.Balance += t.Amount
	case "saving_allocation":
	}

	if err := s.walletRepo.WithTx(tx).Update(oldWallet); err != nil {
		tx.Rollback()
		return nil, err
	}

	t.WalletID = input.WalletID
	t.CategoryID = input.CategoryID
	t.Amount = input.Amount
	t.Type = input.Type
	t.Description = input.Description
	if input.Attachment != "" {
		t.Attachment = input.Attachment
	}
	t.Date = input.Date

	if err := s.repo.WithTx(tx).Update(t); err != nil {
		tx.Rollback()
		return nil, err
	}

	newWallet, err := s.walletRepo.WithTx(tx).FindByID(input.WalletID, userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	switch input.Type {
	case "income", "transfer_in":
		newWallet.Balance += input.Amount
	case "expense", "transfer_out":
		if newWallet.Balance < input.Amount {
			tx.Rollback()
			return nil, errors.New("insufficient wallet balance")
		}
		newWallet.Balance -= input.Amount
	case "saving_allocation":
	}

	if err := s.walletRepo.WithTx(tx).Update(newWallet); err != nil {
		tx.Rollback()
		return nil, err
	}

	if t.RelatedTransactionID != nil {
		relatedID := *t.RelatedTransactionID
		
		var relatedTx entity.Transaction
		if err := tx.Where("id = ?", relatedID).First(&relatedTx).Error; err == nil {
			relatedWallet, err := s.walletRepo.WithTx(tx).FindByID(relatedTx.WalletID, userID)
			if err == nil {
				switch relatedTx.Type {
				case "transfer_in":
					relatedWallet.Balance -= relatedTx.Amount
				case "transfer_out":
					relatedWallet.Balance += relatedTx.Amount
				}
				
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
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	t, err := s.repo.FindByID(id, userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	w, err := s.walletRepo.FindByID(t.WalletID, userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	switch t.Type {
	case "income":
		if w.Balance < t.Amount {
			tx.Rollback()
			return errors.New("insufficient wallet balance to revert income")
		}
		w.Balance -= t.Amount
	case "expense":
		w.Balance += t.Amount
	case "transfer_in":
		if w.Balance < t.Amount {
			tx.Rollback()
			return errors.New("insufficient wallet balance to revert transfer")
		}
		w.Balance -= t.Amount
	case "transfer_out":
		w.Balance += t.Amount
	case "saving_allocation":
	}

	if err := tx.Save(w).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(t).Error; err != nil {
		tx.Rollback()
		return err
	}

	if t.RelatedTransactionID != nil {
		relatedID := *t.RelatedTransactionID
		var relatedTx entity.Transaction
		if err := tx.Where("id = ?", relatedID).First(&relatedTx).Error; err == nil {
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

	fromWallet, err := s.walletRepo.FindByID(input.FromWalletID, userID)
	if err != nil {
		tx.Rollback()
		return errors.New("source wallet not found")
	}

	totalRequired := input.Amount + input.TransferFee
	if fromWallet.Balance < totalRequired {
		tx.Rollback()
		return errors.New("insufficient balance for transfer amount + fee")
	}

	toWallet, err := s.walletRepo.FindByID(input.ToWalletID, userID)
	if err != nil {
		tx.Rollback()
		return errors.New("destination wallet not found")
	}

	transferCatID, err := s.getCategoryForTransfer(userID)
	if err != nil {
		tx.Rollback()
		return err
	}

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
	
	if input.TransferFee > 0 {
		feeCatID, err := s.getCategoryForTransferFee(userID)
		if err != nil {
			tx.Rollback()
			return err
		}

		feeTx := &entity.Transaction{
			UserID:      userID,
			WalletID:    input.FromWalletID,
			CategoryID:  feeCatID,
			Amount:      input.TransferFee,
			Type:        "expense",
			Description: "Biaya Admin Transfer: " + input.Description,
			Date:        input.Date,
		}

		if err := s.repo.WithTx(tx).Create(feeTx); err != nil {
			tx.Rollback()
			return err
		}

		fromWallet.Balance -= input.TransferFee
	}

	if err := tx.Save(fromWallet).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Save(toWallet).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *transactionService) getCategoryForTransfer(userID uint) (uint, error) {

	var cat entity.Category
	
	err := s.db.Where(entity.Category{UserID: userID, Type: "transfer"}).
		Attrs(entity.Category{Name: "Transfer", Icon: "Em_Exchange", BudgetLimit: 0}).
		FirstOrCreate(&cat).Error

	if err != nil {
		return 0, err
	}
	
	return cat.ID, nil
}

func (s *transactionService) getCategoryForTransferFee(userID uint) (uint, error) {
	var cat entity.Category
	err := s.db.Where(entity.Category{UserID: userID, Type: "expense", Name: "Biaya Admin"}).
		Attrs(entity.Category{Icon: "Em_MoneyWing", BudgetLimit: 0}).
		FirstOrCreate(&cat).Error

	if err != nil {
		return 0, err
	}
	
	return cat.ID, nil
}

func (s *transactionService) GetCalendarData(userID uint, startDate, endDate string, walletID *uint, categoryID *uint, search string) ([]entity.TransactionSummary, error) {
	return s.repo.FindSummaryByDateRange(userID, startDate, endDate, walletID, categoryID, search)
}

func (s *transactionService) GetReport(userID uint, startDate, endDate string, walletIDs []uint, filterType *string) ([]entity.CategoryBreakdown, error) {
	return s.repo.GetCategoryBreakdown(userID, startDate, endDate, walletIDs, filterType)
}

func (s *transactionService) ExportTransactions(userID uint, params entity.TransactionFilterParams) (*bytes.Buffer, error) {
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

	sheetName := "Transactions"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	headers := []string{"No", "Date", "Description", "Category", "Wallet", "Type", "Amount"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
	})
	f.SetCellStyle(sheetName, "A1", "G1", style)

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

func (s *transactionService) ExportReport(userID uint, startDate, endDate string, walletIDs []uint, filterType *string) (*bytes.Buffer, error) {
	data, err := s.GetReport(userID, startDate, endDate, walletIDs, filterType)
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

	headers := []string{"No", "Category", "Type", "Total Amount", "Budget Limit", "Is Over Budget"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
	})
	f.SetCellStyle(sheetName, "A1", "F1", style)

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
