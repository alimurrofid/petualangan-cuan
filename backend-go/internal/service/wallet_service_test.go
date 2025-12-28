package service_test

import (
	"testing"

	"cuan-backend/internal/entity"
	"cuan-backend/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockWalletRepository struct {
	mock.Mock
}

func (m *MockWalletRepository) Create(wallet *entity.Wallet) error {
	args := m.Called(wallet)
	return args.Error(0)
}

func (m *MockWalletRepository) Update(wallet *entity.Wallet) error {
	args := m.Called(wallet)
	return args.Error(0)
}

func (m *MockWalletRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockWalletRepository) FindByID(id uint) (*entity.Wallet, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Wallet), args.Error(1)
}

func (m *MockWalletRepository) FindByUserID(userID uint) ([]entity.Wallet, error) {
	args := m.Called(userID)
	return args.Get(0).([]entity.Wallet), args.Error(1)
}

func TestCreateWallet(t *testing.T) {
	mockRepo := new(MockWalletRepository)
	walletService := service.NewWalletService(mockRepo)

	input := service.CreateWalletInput{
		UserID:  1,
		Name:    "Test Wallet",
		Type:    "Bank",
		Balance: 100000,
		Icon:    "Landmark",
	}

	mockRepo.On("Create", mock.AnythingOfType("*entity.Wallet")).Return(nil)

	result, err := walletService.CreateWallet(input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, input.Name, result.Name)
	mockRepo.AssertExpectations(t)
}

func TestGetWalletByID_Success(t *testing.T) {
	mockRepo := new(MockWalletRepository)
	walletService := service.NewWalletService(mockRepo)

	wallet := &entity.Wallet{ID: 1, UserID: 1, Name: "My Wallet"}

	mockRepo.On("FindByID", uint(1)).Return(wallet, nil)

	result, err := walletService.GetWalletByID(1, 1)

	assert.NoError(t, err)
	assert.Equal(t, wallet, result)
}

func TestGetWalletByID_Unauthorized(t *testing.T) {
	mockRepo := new(MockWalletRepository)
	walletService := service.NewWalletService(mockRepo)

	wallet := &entity.Wallet{ID: 1, UserID: 2, Name: "Other Wallet"}

	mockRepo.On("FindByID", uint(1)).Return(wallet, nil)

	result, err := walletService.GetWalletByID(1, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "unauthorized access to wallet", err.Error())
}
