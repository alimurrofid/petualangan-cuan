package service

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"
	"errors"
	"time"

	"gorm.io/gorm"
)

type SavingGoalService interface {
	CreateGoal(userID uint, input CreateGoalInput) (*entity.SavingGoal, error)
	GetGoals(userID uint) ([]entity.SavingGoal, error)
	AddContribution(userID uint, goalID uint, input ContributionInput) (*entity.SavingContribution, error)
	// Additional methods like Update, Delete, Withdraw can be added later
}

type savingGoalService struct {
	repo               repository.SavingGoalRepository
	walletRepo         repository.WalletRepository
	transactionService TransactionService // Re-use transaction creation logic
	db                 *gorm.DB
}

func NewSavingGoalService(repo repository.SavingGoalRepository, walletRepo repository.WalletRepository, transactionService TransactionService, db *gorm.DB) SavingGoalService {
	return &savingGoalService{
		repo:               repo,
		walletRepo:         walletRepo,
		transactionService: transactionService,
		db:                 db,
	}
}

type CreateGoalInput struct {
	Name         string    `json:"name" binding:"required"`
	TargetAmount float64   `json:"target_amount" binding:"required"`
	CategoryID   uint      `json:"category_id" binding:"required"`
	Deadline     *time.Time `json:"deadline"`
	Icon         string    `json:"icon"`
}

type ContributionInput struct {
	WalletID uint      `json:"wallet_id" binding:"required"`
	Amount   float64   `json:"amount" binding:"required"`
	Date     time.Time `json:"date" binding:"required"`
	Description string `json:"description"`
}

func (s *savingGoalService) CreateGoal(userID uint, input CreateGoalInput) (*entity.SavingGoal, error) {
	goal := &entity.SavingGoal{
		UserID:       userID,
		Name:         input.Name,
		TargetAmount: input.TargetAmount,
		CategoryID:   input.CategoryID,
		Deadline:     input.Deadline,
		Icon:         input.Icon,
	}
	
	if err := s.repo.Create(goal); err != nil {
		return nil, err
	}
	
	return goal, nil
}

func (s *savingGoalService) GetGoals(userID uint) ([]entity.SavingGoal, error) {
	return s.repo.FindAll(userID)
}

func (s *savingGoalService) AddContribution(userID uint, goalID uint, input ContributionInput) (*entity.SavingContribution, error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Validate Wallet & Goal coverage
	goal, err := s.repo.FindByID(goalID, userID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("goal not found")
	}

	// 2. Create "Virtual" Transaction (Saving Allocation)
	// We need a special category for saving allocation usually, or we just leave it null/default?
	// It's better to have a system category "Savings" or similar, or just manage it here.
	// For now, let's create a transaction with type 'saving_allocation'.
	
	// Create Transaction via Service is tricky because it commits its own transaction if we call the public method.
	// We should probably call `repo.Create` directly or exposing a WithTx method.
	// Alternatively, we use `transactionService.CreateTransaction` but that might trigger physical balance update if we don't modify it first.
	// Plan: Update TransactionService to handle `saving_allocation` correctly (NO physical update).
	
	// But `TransactionService` does not accept an external TX transaction.
	// So we might need to rely on `TransactionService` handling its own atomic operation 
	// OR we replicate the logic here.
	// Given `TransactionService` is complex, let's try to use it, BUT we need to wrap everything in one DB transaction.
	// The current `TransactionService.CreateTransaction` creates its OWN transaction `s.db.Begin()`.
	// This makes nesting hard without savepoints.
	// Simplified approach for this iteration: Implement the logic "inline" here using `tx`.
	
	// 2a. Create Transaction Record
	// Use goal.CategoryID if set, otherwise fallback to "Tabungan"
	var categoryID uint
	if goal.CategoryID != 0 {
		categoryID = goal.CategoryID
	} else {
		// Fallback: Find "Tabungan" category
		var cat entity.Category
		if err := tx.Where("user_id = ? AND type = ?", userID, "expense").Where("name = ?", "Tabungan").FirstOrCreate(&cat, entity.Category{
			UserID: userID, Name: "Tabungan", Type: "expense", Icon: "PiggyBank",
		}).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		categoryID = cat.ID
	}

	// Use provided description or fallback
	desc := input.Description
	if desc == "" {
		desc = "Alokasi ke " + goal.Name
	}

	transaction := &entity.Transaction{
		UserID:      userID,
		WalletID:    input.WalletID,
		CategoryID:  categoryID,
		Amount:      input.Amount,
		Type:        "saving_allocation", // NEW TYPE
		Date:        input.Date,
		Description: desc,
	}

	// Create Transaction
	if err := tx.Create(transaction).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 3. Create Contribution Record
	contribution := &entity.SavingContribution{
		GoalID:        goal.ID,
		WalletID:      input.WalletID,
		TransactionID: transaction.ID,
		Amount:        input.Amount,
		Date:          input.Date,
	}

	if err := tx.Create(contribution).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 4. Update Goal Progress
	goal.CurrentAmount += input.Amount
	if goal.CurrentAmount >= goal.TargetAmount {
		goal.IsAchieved = true
	}
	
	if err := tx.Save(goal).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return contribution, nil
}
