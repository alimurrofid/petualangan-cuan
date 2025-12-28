package mock

import (
	"cuan-backend/internal/entity"

	"github.com/stretchr/testify/mock"
)

type TransactionRepositoryMock struct {
	mock.Mock
}

func (m *TransactionRepositoryMock) GetAll() ([]entity.Transaction, error) {
	args := m.Called()
	return args.Get(0).([]entity.Transaction), args.Error(1)
}

func (m *TransactionRepositoryMock) Create(transaction entity.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}
