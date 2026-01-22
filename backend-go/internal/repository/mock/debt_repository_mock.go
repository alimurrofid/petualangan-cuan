package mock

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type DebtRepositoryMock struct {
	mock.Mock
}

func (m *DebtRepositoryMock) WithTx(tx *gorm.DB) repository.DebtRepository {
	args := m.Called(tx)
	if args.Get(0) == nil {
		return m
	}
	return args.Get(0).(repository.DebtRepository)
}

func (m *DebtRepositoryMock) Create(debt *entity.Debt) error {
	args := m.Called(debt)
	return args.Error(0)
}

func (m *DebtRepositoryMock) FindByUserID(userID uint, debtType string) ([]entity.Debt, error) {
	args := m.Called(userID, debtType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Debt), args.Error(1)
}

func (m *DebtRepositoryMock) FindByID(id uint, userID uint) (*entity.Debt, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Debt), args.Error(1)
}

func (m *DebtRepositoryMock) Update(debt *entity.Debt) error {
	args := m.Called(debt)
	return args.Error(0)
}

func (m *DebtRepositoryMock) Delete(id uint, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

func (m *DebtRepositoryMock) GetTotalPayments(userID uint, startDate, endDate string) (float64, error) {
	args := m.Called(userID, startDate, endDate)
	return args.Get(0).(float64), args.Error(1)
}
