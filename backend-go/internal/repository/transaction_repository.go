package repository

import (
	"cuan-backend/internal/entity"

	"gorm.io/gorm"
)

type TransactionRepository interface {
	Create(transaction *entity.Transaction) error
	FindAll(userID uint) ([]entity.Transaction, error)
	FindByID(id uint, userID uint) (*entity.Transaction, error)
	Delete(id uint, userID uint) error
	FindSummaryByDateRange(userID uint, startDate, endDate string) ([]entity.TransactionSummary, error)
	// Used for transactions, we need access to DB transaction object
	WithTx(tx *gorm.DB) TransactionRepository
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db}
}

func (r *transactionRepository) WithTx(tx *gorm.DB) TransactionRepository {
	return &transactionRepository{db: tx}
}

func (r *transactionRepository) Create(transaction *entity.Transaction) error {
	return r.db.Create(transaction).Error
}

func (r *transactionRepository) FindAll(userID uint) ([]entity.Transaction, error) {
	var transactions []entity.Transaction
	// Preload Wallet and Category to get names and icons
	err := r.db.Where("user_id = ?", userID).
		Preload("Wallet").
		Preload("Category").
		Order("date desc, created_at desc").
		Find(&transactions).Error
	return transactions, err
}

func (r *transactionRepository) FindByID(id uint, userID uint) (*entity.Transaction, error) {
	var transaction entity.Transaction
	err := r.db.Where("id = ? AND user_id = ?", id, userID).
		Preload("Wallet").
		Preload("Category").
		First(&transaction).Error
	return &transaction, err
}

func (r *transactionRepository) Delete(id uint, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&entity.Transaction{}).Error
}

func (r *transactionRepository) FindSummaryByDateRange(userID uint, startDate, endDate string) ([]entity.TransactionSummary, error) {
	var results []entity.TransactionSummary
	
	// Postgres specific: DATE(date)
	err := r.db.Model(&entity.Transaction{}).
		Select("TO_CHAR(date, 'YYYY-MM-DD') as date, SUM(CASE WHEN type = 'income' OR type = 'transfer_in' THEN amount ELSE 0 END) as income, SUM(CASE WHEN type = 'expense' OR type = 'transfer_out' THEN amount ELSE 0 END) as expense").
		Where("user_id = ? AND date >= ? AND date <= ?", userID, startDate, endDate).
		Group("TO_CHAR(date, 'YYYY-MM-DD')").
		Scan(&results).Error

	return results, err
}
