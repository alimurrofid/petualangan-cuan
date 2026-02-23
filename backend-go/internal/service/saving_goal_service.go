package service

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type SavingGoalService interface {
	CreateGoal(userID uint, input CreateGoalInput) (*entity.SavingGoal, error)
	GetGoals(userID uint) ([]entity.SavingGoal, error)
	AddContribution(userID uint, goalID uint, input ContributionInput) (*entity.SavingContribution, error)
	UpdateGoal(userID uint, goalID uint, input CreateGoalInput) (*entity.SavingGoal, error)
	DeleteGoal(userID uint, goalID uint) error
	DeleteContribution(userID uint, contributionID uint) error
	FinishGoal(userID uint, goalID uint) error
}

type savingGoalService struct {
	repo               repository.SavingGoalRepository
	walletRepo         repository.WalletRepository
	transactionService TransactionService
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
		log.Error().Err(err).Uint("user_id", userID).Msg("Failed to create saving goal")
		return nil, err
	}
	
	log.Info().Uint("user_id", userID).Uint("goal_id", goal.ID).Str("name", goal.Name).Msg("Saving goal created successfully")
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

	goal, err := s.repo.FindByID(goalID, userID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("goal not found")
	}

	wallet, err := s.walletRepo.WithTx(tx).FindByID(input.WalletID, userID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("wallet not found")
	}

	if wallet.Balance < input.Amount {
		tx.Rollback()
		return nil, errors.New("insufficient wallet balance")
	}

	var categoryID uint
	if goal.CategoryID != 0 {
		categoryID = goal.CategoryID
	} else {
		var cat entity.Category
		if err := tx.Where("user_id = ? AND type = ?", userID, "expense").Where("name = ?", "Tabungan").FirstOrCreate(&cat, entity.Category{
			UserID: userID, Name: "Tabungan", Type: "expense", Icon: "PiggyBank",
		}).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		categoryID = cat.ID
	}

	desc := input.Description
	if desc == "" {
		desc = "Alokasi ke " + goal.Name
	}

	transaction := &entity.Transaction{
		UserID:      userID,
		WalletID:    input.WalletID,
		CategoryID:  categoryID,
		Amount:      input.Amount,
		Type:        "saving_allocation",
		Date:        input.Date,
		Description: desc,
	}

	if err := tx.Create(transaction).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
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

	goal.CurrentAmount += input.Amount
	if goal.CurrentAmount >= goal.TargetAmount {
		goal.IsAchieved = true
	}
	
	if err := tx.Save(goal).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		log.Error().Err(err).Uint("user_id", userID).Uint("goal_id", goalID).Msg("Failed to commit db transaction for AddContribution")
		return nil, err
	}

	log.Info().Uint("user_id", userID).Uint("goal_id", goalID).Uint("contribution_id", contribution.ID).Msg("Saving goal contribution added successfully")
	return contribution, nil
}

func (s *savingGoalService) UpdateGoal(userID uint, goalID uint, input CreateGoalInput) (*entity.SavingGoal, error) {
	goal, err := s.repo.FindByID(goalID, userID)
	if err != nil {
		return nil, errors.New("goal not found")
	}

	goal.Name = input.Name
	goal.TargetAmount = input.TargetAmount
	goal.CategoryID = input.CategoryID
	goal.Deadline = input.Deadline
	goal.Icon = input.Icon

	if err := s.repo.Update(goal); err != nil {
		log.Error().Err(err).Uint("user_id", userID).Uint("goal_id", goalID).Msg("Failed to update saving goal")
		return nil, err
	}

	log.Info().Uint("user_id", userID).Uint("goal_id", goalID).Msg("Saving goal updated successfully")
	return goal, nil
}

func (s *savingGoalService) DeleteGoal(userID uint, goalID uint) error {
	goal, err := s.repo.FindByID(goalID, userID)
	if err != nil {
		return errors.New("goal not found")
	}

	if err := s.repo.DeleteContributions(goalID); err != nil {
		return err
	}

	err = s.repo.Delete(goal)
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Uint("goal_id", goalID).Msg("Failed to delete saving goal")
	} else {
		log.Info().Uint("user_id", userID).Uint("goal_id", goalID).Msg("Saving goal deleted successfully")
	}
	return err
}

func (s *savingGoalService) DeleteContribution(userID uint, contributionID uint) error {
	contribution, err := s.repo.FindContributionByID(contributionID)
	if err != nil {
		return errors.New("contribution not found")
	}

	if contribution.SavingGoal.UserID != userID {
		return errors.New("unauthorized")
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	contribution.SavingGoal.CurrentAmount -= contribution.Amount
	if contribution.SavingGoal.CurrentAmount < contribution.SavingGoal.TargetAmount {
		contribution.SavingGoal.IsAchieved = false
	}

	if err := tx.Save(&contribution.SavingGoal).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	if err := tx.Delete(&entity.SavingContribution{}, contributionID).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(&entity.Transaction{}, contribution.TransactionID).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		log.Error().Err(err).Uint("user_id", userID).Uint("contribution_id", contributionID).Msg("Failed to commit db transaction for DeleteContribution")
		return err
	}

	log.Info().Uint("user_id", userID).Uint("contribution_id", contributionID).Msg("Saving goal contribution deleted successfully")
	return nil
}

func (s *savingGoalService) FinishGoal(userID uint, goalID uint) error {
	goal, err := s.repo.FindByID(goalID, userID)
	if err != nil {
		return errors.New("goal not found")
	}

	if !goal.IsAchieved {
		return errors.New("goal is not achieved yet")
	}

	if goal.IsFinished {
		return errors.New("goal is already finished")
	}

	goal.IsFinished = true
	err = s.repo.Update(goal)
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Uint("goal_id", goalID).Msg("Failed to mark saving goal as finished")
	} else {
		log.Info().Uint("user_id", userID).Uint("goal_id", goalID).Msg("Saving goal marked as finished successfully")
	}
	return err
}
