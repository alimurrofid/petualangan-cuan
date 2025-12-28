package mock

import (
	"cuan-backend/internal/entity"

	"github.com/stretchr/testify/mock"
)

type CategoryRepositoryMock struct {
	mock.Mock
}

func (m *CategoryRepositoryMock) Create(category *entity.Category) error {
	args := m.Called(category)
	return args.Error(0)
}

func (m *CategoryRepositoryMock) FindAll(userID uint) ([]entity.Category, error) {
	args := m.Called(userID)
	return args.Get(0).([]entity.Category), args.Error(1)
}

func (m *CategoryRepositoryMock) FindByID(id uint, userID uint) (*entity.Category, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Category), args.Error(1)
}

func (m *CategoryRepositoryMock) Update(category *entity.Category) error {
	args := m.Called(category)
	return args.Error(0)
}

func (m *CategoryRepositoryMock) Delete(id uint, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}
