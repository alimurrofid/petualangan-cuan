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

func (m *TransactionRepositoryMock) FindAll(userID uint, params entity.TransactionFilterParams) ([]entity.Transaction, int64, error) {
	args := m.Called(userID, params)
	return args.Get(0).([]entity.Transaction), args.Get(1).(int64), args.Error(2)
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

func (m *TransactionRepositoryMock) FindSummaryByDateRange(userID uint, startDate, endDate string, walletID *uint, categoryID *uint, search string) ([]entity.TransactionSummary, error) {
	args := m.Called(userID, startDate, endDate, walletID, categoryID, search)
	return args.Get(0).([]entity.TransactionSummary), args.Error(1)
}

func (m *TransactionRepositoryMock) GetCategoryBreakdown(userID uint, startDate, endDate string, walletID *uint, filterType *string) ([]entity.CategoryBreakdown, error) {
	args := m.Called(userID, startDate, endDate, walletID, filterType)
	return args.Get(0).([]entity.CategoryBreakdown), args.Error(1)
}

func (m *TransactionRepositoryMock) GetMonthlyTrend(userID uint, startDate, endDate string) ([]entity.MonthlyTrend, error) {
	args := m.Called(userID, startDate, endDate)
	return args.Get(0).([]entity.MonthlyTrend), args.Error(1)
}

func (m *TransactionRepositoryMock) GetRecentTransactions(userID uint, limit int) ([]entity.Transaction, error) {
	args := m.Called(userID, limit)
	return args.Get(0).([]entity.Transaction), args.Error(1)
}

func (m *TransactionRepositoryMock) WithTx(tx *gorm.DB) repository.TransactionRepository {
	args := m.Called(tx)
	return args.Get(0).(repository.TransactionRepository)
}
