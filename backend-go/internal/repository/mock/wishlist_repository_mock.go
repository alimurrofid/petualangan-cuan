package mock

import (
	"cuan-backend/internal/entity"

	"github.com/stretchr/testify/mock"
)

type WishlistRepositoryMock struct {
	mock.Mock
}

func (m *WishlistRepositoryMock) Create(item *entity.WishlistItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *WishlistRepositoryMock) FindAllByUserID(userID uint) ([]entity.WishlistItem, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.WishlistItem), args.Error(1)
}

func (m *WishlistRepositoryMock) FindByID(id, userID uint) (*entity.WishlistItem, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WishlistItem), args.Error(1)
}

func (m *WishlistRepositoryMock) Update(item *entity.WishlistItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *WishlistRepositoryMock) Delete(id, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

func (m *WishlistRepositoryMock) MarkAsBought(id, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}
