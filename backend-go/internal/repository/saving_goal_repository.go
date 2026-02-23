package repository

import (
	"cuan-backend/internal/entity"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type SavingGoalRepository interface {
	Create(goal *entity.SavingGoal) error
	FindAll(userID uint) ([]entity.SavingGoal, error)
	FindByID(id uint, userID uint) (*entity.SavingGoal, error)
	Update(goal *entity.SavingGoal) error
	Delete(goal *entity.SavingGoal) error

	AddContribution(contribution *entity.SavingContribution) error
	GetActiveContributions(walletID uint) (float64, error)
	DeleteContributions(goalID uint) error
	FindContributionByID(id uint) (*entity.SavingContribution, error)
	DeleteContribution(contribution *entity.SavingContribution) error
}

type savingGoalRepository struct {
	db *gorm.DB
}

func NewSavingGoalRepository(db *gorm.DB) SavingGoalRepository {
	return &savingGoalRepository{db: db}
}

func (r *savingGoalRepository) Create(goal *entity.SavingGoal) error {
	if err := r.db.Create(goal).Error; err != nil {
		log.Error().Err(err).Uint("user_id", goal.UserID).Msg("Database operation failed")
		return err
	}
	return nil
}

func (r *savingGoalRepository) FindAll(userID uint) ([]entity.SavingGoal, error) {
	var goals []entity.SavingGoal
	err := r.db.Where("user_id = ?", userID).
		Preload("Contributions", func(db *gorm.DB) *gorm.DB {
			return db.Order("date desc")
		}).
		Preload("Contributions.Wallet").
		Preload("Contributions.Transaction").
		Find(&goals).Error
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Msg("Database operation failed")
	}
	return goals, err
}

func (r *savingGoalRepository) FindByID(id uint, userID uint) (*entity.SavingGoal, error) {
	var goal entity.SavingGoal
	err := r.db.Where("id = ? AND user_id = ?", id, userID).
		Preload("Contributions", func(db *gorm.DB) *gorm.DB {
			return db.Order("date desc")
		}).
		Preload("Contributions.Wallet").
		Preload("Contributions.Transaction").
		First(&goal).Error
	if err != nil {
		log.Error().Err(err).Uint("saving_goal_id", id).Uint("user_id", userID).Msg("Database operation failed")
		return nil, err
	}
	return &goal, nil
}

func (r *savingGoalRepository) Update(goal *entity.SavingGoal) error {
	if err := r.db.Save(goal).Error; err != nil {
		log.Error().Err(err).Uint("saving_goal_id", goal.ID).Uint("user_id", goal.UserID).Msg("Database operation failed")
		return err
	}
	return nil
}

func (r *savingGoalRepository) Delete(goal *entity.SavingGoal) error {
	if err := r.db.Delete(goal).Error; err != nil {
		log.Error().Err(err).Uint("saving_goal_id", goal.ID).Uint("user_id", goal.UserID).Msg("Database operation failed")
		return err
	}
	return nil
}

func (r *savingGoalRepository) AddContribution(contribution *entity.SavingContribution) error {
	if err := r.db.Create(contribution).Error; err != nil {
		log.Error().Err(err).Uint("goal_id", contribution.GoalID).Msg("Database operation failed")
		return err
	}
	return nil
}

func (r *savingGoalRepository) GetActiveContributions(walletID uint) (float64, error) {
	var total float64
	err := r.db.Table("saving_contributions").
		Joins("JOIN saving_goals ON saving_goals.id = saving_contributions.goal_id").
		Where("saving_contributions.wallet_id = ? AND saving_goals.is_finished = ?", walletID, false).
		Select("COALESCE(SUM(saving_contributions.amount), 0)").
		Scan(&total).Error
	if err != nil {
		log.Error().Err(err).Uint("wallet_id", walletID).Msg("Database operation failed")
	}
	return total, err
}

func (r *savingGoalRepository) DeleteContributions(goalID uint) error {
	if err := r.db.Where("goal_id = ?", goalID).Delete(&entity.SavingContribution{}).Error; err != nil {
		log.Error().Err(err).Uint("goal_id", goalID).Msg("Database operation failed")
		return err
	}
	return nil
}

func (r *savingGoalRepository) FindContributionByID(id uint) (*entity.SavingContribution, error) {
	var contribution entity.SavingContribution
	err := r.db.Preload("Transaction").Preload("SavingGoal").First(&contribution, id).Error
	if err != nil {
		log.Error().Err(err).Uint("contribution_id", id).Msg("Database operation failed")
		return nil, err
	}
	return &contribution, nil
}

func (r *savingGoalRepository) DeleteContribution(contribution *entity.SavingContribution) error {
	if err := r.db.Delete(contribution).Error; err != nil {
		log.Error().Err(err).Uint("contribution_id", contribution.ID).Msg("Database operation failed")
		return err
	}
	return nil
}
