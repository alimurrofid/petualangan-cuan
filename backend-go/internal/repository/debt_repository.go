package repository

import (
	"cuan-backend/internal/entity"

	"gorm.io/gorm"
)

type DebtRepository interface {
	Create(debt *entity.Debt) error
	Update(debt *entity.Debt) error
	Delete(id uint, userID uint) error
	FindByID(id uint, userID uint) (*entity.Debt, error)
	FindByUserID(userID uint, debtType string) ([]entity.Debt, error)
	GetTotalPayments(userID uint, startDate, endDate string) (float64, error)
	WithTx(tx *gorm.DB) DebtRepository
}

type debtRepository struct {
	db *gorm.DB
}

func NewDebtRepository(db *gorm.DB) DebtRepository {
	return &debtRepository{db}
}

func (r *debtRepository) WithTx(tx *gorm.DB) DebtRepository {
	return &debtRepository{db: tx}
}

func (r *debtRepository) Create(debt *entity.Debt) error {
	return r.db.Create(debt).Error
}

func (r *debtRepository) Update(debt *entity.Debt) error {
	return r.db.Save(debt).Error
}

func (r *debtRepository) Delete(id uint, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&entity.Debt{}).Error
}

func (r *debtRepository) FindByID(id uint, userID uint) (*entity.Debt, error) {
	var debt entity.Debt
	err := r.db.Preload("Wallet").Preload("Payments").Preload("Payments.Wallet").Where("id = ? AND user_id = ?", id, userID).First(&debt).Error
	if err != nil {
		return nil, err
	}
	return &debt, nil
}

func (r *debtRepository) FindByUserID(userID uint, debtType string) ([]entity.Debt, error) {
	var debts []entity.Debt
	query := r.db.Preload("Wallet").Preload("Payments").Preload("Payments.Wallet").Where("user_id = ?", userID)
	if debtType != "" {
		query = query.Where("type = ?", debtType)
	}
	err := query.Order("created_at desc").Find(&debts).Error
	return debts, err
}

func (r *debtRepository) GetTotalPayments(userID uint, startDate, endDate string) (float64, error) {
	var total float64
	err := r.db.Model(&entity.DebtPayment{}).
		Joins("JOIN debts ON debts.id = debt_payments.debt_id").
		Where("debts.user_id = ? AND debts.type = ? AND debt_payments.date BETWEEN ? AND ?", userID, entity.DebtTypePayable, startDate, endDate).
		Select("COALESCE(SUM(debt_payments.amount), 0)").
		Scan(&total).Error
	return total, err
}
