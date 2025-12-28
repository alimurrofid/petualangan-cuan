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
