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

	// Contributions
	AddContribution(contribution *entity.SavingContribution) error
	GetActiveContributions(walletID uint) (float64, error) // For calculating available balance
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
			return db.Order("date desc") // Order contributions by date
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

// GetActiveContributions returns the total amount allocated to saving goals from a specific wallet.
// Only counts goals that are NOT yet achieved (or maybe we always count them? 
// Requirement: "Active Contributions" usually implies funds are still "locked" in the goal.
// If an achieved goal is "cashed out", then we might need to handle that.
// For now, let's assume ALL contributions to ANY goal (achieved or not) are deducted from Available Balance
// because the money is physically there but logically "spent" on the goal.
// UNLESS the goal is deleted or funds are withdrawn (functionality for later).
func (r *savingGoalRepository) GetActiveContributions(walletID uint) (float64, error) {
	var total float64
	// Sum amount of all contributions where wallet_id = X
	// In a more complex system, we might filter by Goal status.
	// For "Petualangan Cuan", let's assume all contributions count.
	err := r.db.Model(&entity.SavingContribution{}).
		Where("wallet_id = ?", walletID).
		Select("COALESCE(SUM(amount), 0)").
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
