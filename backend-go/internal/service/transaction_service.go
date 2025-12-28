package service

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"
)

type TransactionService interface {
	GetTransactions() ([]entity.Transaction, error)
}

type transactionService struct {
	repo repository.TransactionRepository
}

func NewTransactionService(repo repository.TransactionRepository) TransactionService {
	return &transactionService{repo}
}

func (s *transactionService) GetTransactions() ([]entity.Transaction, error) {
	return s.repo.GetAll()
}
