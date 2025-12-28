package mock

import (
	"cuan-backend/internal/entity"

	"github.com/stretchr/testify/mock"
)

type WalletRepositoryMock struct {
	mock.Mock
}

func (m *WalletRepositoryMock) Create(wallet *entity.Wallet) error {
	args := m.Called(wallet)
	return args.Error(0)
}

func (m *WalletRepositoryMock) Update(wallet *entity.Wallet) error {
	args := m.Called(wallet)
	return args.Error(0)
}

func (m *WalletRepositoryMock) Delete(id uint, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

func (m *WalletRepositoryMock) FindByID(id uint, userID uint) (*entity.Wallet, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Wallet), args.Error(1)
}

func (m *WalletRepositoryMock) FindByUserID(userID uint) ([]entity.Wallet, error) {
	args := m.Called(userID)
	return args.Get(0).([]entity.Wallet), args.Error(1)
}
