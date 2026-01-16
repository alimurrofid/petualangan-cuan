package mock

import (
	"cuan-backend/internal/entity"

	"github.com/stretchr/testify/mock"
)

type SavingGoalRepositoryMock struct {
	mock.Mock
}

func (m *SavingGoalRepositoryMock) Create(goal *entity.SavingGoal) error {
	args := m.Called(goal)
	return args.Error(0)
}

func (m *SavingGoalRepositoryMock) FindAll(userID uint) ([]entity.SavingGoal, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.SavingGoal), args.Error(1)
}

func (m *SavingGoalRepositoryMock) FindByID(id uint, userID uint) (*entity.SavingGoal, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.SavingGoal), args.Error(1)
}

func (m *SavingGoalRepositoryMock) Update(goal *entity.SavingGoal) error {
	args := m.Called(goal)
	return args.Error(0)
}

func (m *SavingGoalRepositoryMock) Delete(goal *entity.SavingGoal) error {
	args := m.Called(goal)
	return args.Error(0)
}

func (m *SavingGoalRepositoryMock) AddContribution(contribution *entity.SavingContribution) error {
	args := m.Called(contribution)
	return args.Error(0)
}

func (m *SavingGoalRepositoryMock) GetActiveContributions(walletID uint) (float64, error) {
	args := m.Called(walletID)
	return args.Get(0).(float64), args.Error(1)
}
