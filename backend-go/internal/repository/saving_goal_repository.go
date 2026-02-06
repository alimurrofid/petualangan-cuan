package repository

import (
	"cuan-backend/internal/entity"

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
	return r.db.Create(goal).Error
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
		return nil, err
	}
	return &goal, nil
}

func (r *savingGoalRepository) Update(goal *entity.SavingGoal) error {
	return r.db.Save(goal).Error
}

func (r *savingGoalRepository) Delete(goal *entity.SavingGoal) error {
	return r.db.Delete(goal).Error
}

func (r *savingGoalRepository) AddContribution(contribution *entity.SavingContribution) error {
	return r.db.Create(contribution).Error
}

func (r *savingGoalRepository) GetActiveContributions(walletID uint) (float64, error) {
	var total float64
	err := r.db.Table("saving_contributions").
		Joins("JOIN saving_goals ON saving_goals.id = saving_contributions.goal_id").
		Where("saving_contributions.wallet_id = ? AND saving_goals.is_finished = ?", walletID, false).
		Select("COALESCE(SUM(saving_contributions.amount), 0)").
		Scan(&total).Error
	return total, err
}

func (r *savingGoalRepository) DeleteContributions(goalID uint) error {
	return r.db.Where("goal_id = ?", goalID).Delete(&entity.SavingContribution{}).Error
}

func (r *savingGoalRepository) FindContributionByID(id uint) (*entity.SavingContribution, error) {
	var contribution entity.SavingContribution
	err := r.db.Preload("Transaction").Preload("SavingGoal").First(&contribution, id).Error
	if err != nil {
		return nil, err
	}
	return &contribution, nil
}

func (r *savingGoalRepository) DeleteContribution(contribution *entity.SavingContribution) error {
	return r.db.Delete(contribution).Error
}
