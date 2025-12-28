package mock

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type TransactionRepositoryMock struct {
	mock.Mock
}

func (m *TransactionRepositoryMock) Create(transaction *entity.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *TransactionRepositoryMock) FindAll(userID uint) ([]entity.Transaction, error) {
	args := m.Called(userID)
	return args.Get(0).([]entity.Transaction), args.Error(1)
}

func (m *TransactionRepositoryMock) FindByID(id uint, userID uint) (*entity.Transaction, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Transaction), args.Error(1)
}

func (m *TransactionRepositoryMock) Delete(id uint, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

func (m *TransactionRepositoryMock) WithTx(tx *gorm.DB) repository.TransactionRepository {
	args := m.Called(tx)
	return args.Get(0).(repository.TransactionRepository)
}
