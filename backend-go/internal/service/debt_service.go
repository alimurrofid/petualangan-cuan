package service

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"
	"errors"
	"time"

	"gorm.io/gorm"
)

type UpdateDebtInput struct {
	WalletID    uint      `json:"wallet_id" binding:"required"`
	Name        string    `json:"name" binding:"required"`
	Amount      float64   `json:"amount" binding:"required,gt=0"`
	Description string    `json:"description"`
	DueDate     *time.Time `json:"due_date"`
}

func (s *debtService) UpdateDebt(id uint, userID uint, input UpdateDebtInput) (*entity.Debt, error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return nil, tx.Error
	}

	debt, err := s.debtRepo.WithTx(tx).FindByID(id, userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	paidAmount := debt.Amount - debt.Remaining
	if input.Amount < paidAmount {
		tx.Rollback()
		return nil, errors.New("new amount cannot be less than already paid amount")
	}

	oldWallet, err := s.walletRepo.WithTx(tx).FindByID(debt.WalletID, userID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("old wallet not found")
	}

	if debt.Type == entity.DebtTypePayable {
		oldWallet.Balance -= debt.Amount
	} else {
		oldWallet.Balance += debt.Amount
	}

	if err := s.walletRepo.WithTx(tx).Update(oldWallet); err != nil {
		tx.Rollback()
		return nil, err
	}

	newWallet, err := s.walletRepo.WithTx(tx).FindByID(input.WalletID, userID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("new wallet not found")
	}

	if debt.Type == entity.DebtTypePayable {
		newWallet.Balance += input.Amount
	} else {
		newWallet.Balance -= input.Amount
	}

	if err := s.walletRepo.WithTx(tx).Update(newWallet); err != nil {
		tx.Rollback()
		return nil, err
	}

	debt.Name = input.Name
	debt.Description = input.Description
	debt.DueDate = input.DueDate
	debt.WalletID = input.WalletID
	debt.Amount = input.Amount
	debt.Remaining = input.Amount - paidAmount

    if debt.Remaining <= 0 {
        debt.Remaining = 0
        debt.IsPaid = true
    } else {
        debt.IsPaid = false
    }

	if err := s.debtRepo.WithTx(tx).Update(debt); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return debt, nil
}

type CreateDebtInput struct {
	WalletID    uint      `json:"wallet_id" binding:"required"`
	Name        string    `json:"name" binding:"required"`
	Amount      float64   `json:"amount" binding:"required,gt=0"`
	Type        string    `json:"type" binding:"required,oneof=debt receivable"`
	Description string    `json:"description"`
	DueDate     *time.Time `json:"due_date"`
}

type PayDebtInput struct {
	WalletID uint    `json:"wallet_id" binding:"required"`
	Amount   float64 `json:"amount" binding:"required,gt=0"`
    Note     string  `json:"note"`
}

type DebtService interface {
	CreateDebt(userID uint, input CreateDebtInput) (*entity.Debt, error)
	GetDebts(userID uint, debtType string) ([]entity.Debt, error)
	GetDebt(id uint, userID uint) (*entity.Debt, error)
	PayDebt(id uint, userID uint, input PayDebtInput) (*entity.Debt, error)
	UpdateDebt(id uint, userID uint, input UpdateDebtInput) (*entity.Debt, error)
	DeleteDebt(id uint, userID uint) error
	DeletePayment(id uint, userID uint) error
}

type debtService struct {
	debtRepo        repository.DebtRepository
	transactionRepo repository.TransactionRepository
	walletRepo      repository.WalletRepository
	db              *gorm.DB
}

func NewDebtService(debtRepo repository.DebtRepository, transactionRepo repository.TransactionRepository, walletRepo repository.WalletRepository, db *gorm.DB) DebtService {
	return &debtService{
		debtRepo:        debtRepo,
		transactionRepo: transactionRepo,
		walletRepo:      walletRepo,
		db:              db,
	}
}

func (s *debtService) CreateDebt(userID uint, input CreateDebtInput) (*entity.Debt, error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return nil, tx.Error
	}

	debt := &entity.Debt{
		UserID:      userID,
		WalletID:    input.WalletID,
		Name:        input.Name,
		Amount:      input.Amount,
		Remaining:   input.Amount,
		Type:        entity.DebtType(input.Type),
		Description: input.Description,
		DueDate:     input.DueDate,
		IsPaid:      false,
	}

	if err := s.debtRepo.WithTx(tx).Create(debt); err != nil {
		tx.Rollback()
		return nil, err
	}

	wallet, err := s.walletRepo.WithTx(tx).FindByID(input.WalletID, userID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("wallet not found")
	}

	var transactionType string
	var categoryName string
	if input.Type == string(entity.DebtTypePayable) {
		transactionType = "income"
		categoryName = "Utang"
		wallet.Balance += input.Amount
	} else {
		transactionType = "expense"
		categoryName = "Piutang"
		wallet.Balance -= input.Amount
	}

	var cat entity.Category
	err = tx.Where(entity.Category{UserID: userID, Name: categoryName, Type: transactionType}).
		Attrs(entity.Category{Icon: "Em_MoneyBag", BudgetLimit: 0}).
		FirstOrCreate(&cat).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := s.walletRepo.WithTx(tx).Update(wallet); err != nil {
		tx.Rollback()
		return nil, err
	}

	transaction := &entity.Transaction{
		UserID:      userID,
		WalletID:    input.WalletID,
		CategoryID:  cat.ID,
		Amount:      input.Amount,
		Type:        transactionType,
		Description: "Debt/Receivable: " + input.Name + " - " + input.Description,
		Date:        time.Now(),
	}

	if err := s.transactionRepo.WithTx(tx).Create(transaction); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return debt, nil
}

func (s *debtService) GetDebts(userID uint, debtType string) ([]entity.Debt, error) {
	return s.debtRepo.FindByUserID(userID, debtType)
}

func (s *debtService) GetDebt(id uint, userID uint) (*entity.Debt, error) {
	return s.debtRepo.FindByID(id, userID)
}

func (s *debtService) PayDebt(id uint, userID uint, input PayDebtInput) (*entity.Debt, error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return nil, tx.Error
	}

	debt, err := s.debtRepo.WithTx(tx).FindByID(id, userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if debt.IsPaid {
		tx.Rollback()
		return nil, errors.New("debt is already fully paid")
	}

	if input.Amount > debt.Remaining {
		tx.Rollback()
		return nil, errors.New("payment amount exceeds remaining debt")
	}

	debt.Remaining -= input.Amount
	if debt.Remaining <= 0 {
		debt.Remaining = 0
		debt.IsPaid = true
	}

	if err := s.debtRepo.WithTx(tx).Update(debt); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 3. Handle Wallet & Transaction
	wallet, err := s.walletRepo.WithTx(tx).FindByID(input.WalletID, userID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("wallet not found")
	}

	var transactionType string
	var categoryName string

	if debt.Type == entity.DebtTypePayable {
		transactionType = "expense"
		categoryName = "Bayar Utang"
		wallet.Balance -= input.Amount
	} else {
		transactionType = "income"
		categoryName = "Terima Piutang"
		wallet.Balance += input.Amount
	}

	var cat entity.Category
	err = tx.Where(entity.Category{UserID: userID, Name: categoryName, Type: transactionType}).
		Attrs(entity.Category{Icon: "Em_MoneyBag", BudgetLimit: 0}).
		FirstOrCreate(&cat).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := s.walletRepo.WithTx(tx).Update(wallet); err != nil {
		tx.Rollback()
		return nil, err
	}

    description := "Payment for: " + debt.Name
    if input.Note != "" {
        description += " - " + input.Note
    } else {
        description += " (" + debt.Description + ")"
    }

	transaction := &entity.Transaction{
		UserID:      userID,
		WalletID:    input.WalletID,
		CategoryID:  cat.ID,
		Amount:      input.Amount,
		Type:        transactionType,
		Description: description,
		Date:        time.Now(),
	}

	if err := s.transactionRepo.WithTx(tx).Create(transaction); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 4. Create DebtPayment Record through Association or direct create
	// We need TransactionID for this.
	debtPayment := &entity.DebtPayment{
		DebtID:        debt.ID,
		TransactionID: transaction.ID,
		WalletID:      input.WalletID,
		Amount:        input.Amount,
		Date:          time.Now(),
		Note:          input.Note,
	}
    
    if debtPayment.Note == "" {
        debtPayment.Note = "Pembayaran Cicilan"
    }

	// We can't easily use a repository for this new entity unless we add it to the interface or use DB directly.
	// Using simple approach: s.db.Create (bound to tx)
	if err := tx.Create(debtPayment).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return debt, nil
}

func (s *debtService) DeleteDebt(id uint, userID uint) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return tx.Error
	}

	debt, err := s.debtRepo.WithTx(tx).FindByID(id, userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	wallet, err := s.walletRepo.WithTx(tx).FindByID(debt.WalletID, userID)
	if err != nil {
		tx.Rollback()
		return errors.New("wallet linked to debt not found")
	}

	if debt.Type == entity.DebtTypePayable {
		wallet.Balance -= debt.Remaining
	} else {
		wallet.Balance += debt.Remaining
	}

	if err := s.walletRepo.WithTx(tx).Update(wallet); err != nil {
		tx.Rollback()
		return err
	}

    // Delete associated DebtPayments first to avoid FK constraint violation
    if err := tx.Where("debt_id = ?", id).Delete(&entity.DebtPayment{}).Error; err != nil {
        tx.Rollback()
        return err
    }
    
	if err := s.debtRepo.WithTx(tx).Delete(id, userID); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *debtService) DeletePayment(id uint, userID uint) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return tx.Error
	}

	// 1. Find DebtPayment
	var payment entity.DebtPayment
	if err := tx.Preload("Debt").Preload("Transaction").First(&payment, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Verify User via Debt (or Transaction) ownership
	if payment.Debt.UserID != userID {
		tx.Rollback()
		return errors.New("unauthorized")
	}

	// 2. Revert Wallet Balance
	wallet, err := s.walletRepo.WithTx(tx).FindByID(payment.WalletID, userID)
	if err != nil {
		tx.Rollback()
		return errors.New("wallet not found")
	}

	// Logic: Inverting the payment effect.
	// Payment for Debt(Payable) was Expense (Balance - Amount). Revert: Balance + Amount.
	// Payment for Receivable was Income (Balance + Amount). Revert: Balance - Amount.
	// Check Transaction Type or Debt Type.
	if payment.Debt.Type == entity.DebtTypePayable {
		wallet.Balance += payment.Amount
	} else {
		wallet.Balance -= payment.Amount
	}

	if err := s.walletRepo.WithTx(tx).Update(wallet); err != nil {
		tx.Rollback()
		return err
	}

	// 3. Revert Debt Remaining
	payment.Debt.Remaining += payment.Amount

	// Check if it was paid, now might not be.
	if payment.Debt.Remaining > 0 {
		payment.Debt.IsPaid = false
	}

	if err := s.debtRepo.WithTx(tx).Update(&payment.Debt); err != nil {
		tx.Rollback()
		return err
	}

	// 4. Delete Records - MUST DELETE Payment before Transaction due to FK
    if err := tx.Delete(&payment).Error; err != nil {
        tx.Rollback()
        return err
    }

	if err := s.transactionRepo.WithTx(tx).Delete(payment.TransactionID, userID); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
