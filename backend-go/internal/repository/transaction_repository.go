package repository

import (
	"cuan-backend/internal/entity"

	"gorm.io/gorm"
)

type TransactionRepository interface {
	GetAll() ([]entity.Transaction, error)
	Create(transaction entity.Transaction) error
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db}
}

func (r *transactionRepository) GetAll() ([]entity.Transaction, error) {
	var transactions []entity.Transaction
	err := r.db.Order("created_at desc").Find(&transactions).Error
	return transactions, err
}

func (r *transactionRepository) Create(transaction entity.Transaction) error {
	return r.db.Create(&transaction).Error
}
