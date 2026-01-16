package service_test

import (
	"errors"
	"testing"

	"cuan-backend/internal/entity"
	"cuan-backend/internal/service"

	"cuan-backend/internal/repository/mock"

	"github.com/stretchr/testify/assert"
	testifyMock "github.com/stretchr/testify/mock"
)

func TestCreateWallet(t *testing.T) {
	mockRepo := new(mock.WalletRepositoryMock)
	mockSavingRepo := new(mock.SavingGoalRepositoryMock)
	walletService := service.NewWalletService(mockRepo, mockSavingRepo)

	input := service.CreateWalletInput{
		UserID:  1,
		Name:    "Test Wallet",
		Type:    "Bank",
		Balance: 100000,
		Icon:    "Landmark",
	}

	mockRepo.On("Create", testifyMock.AnythingOfType("*entity.Wallet")).Return(nil)

	result, err := walletService.CreateWallet(input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, input.Name, result.Name)
	mockRepo.AssertExpectations(t)
}

func TestGetWalletByID_Success(t *testing.T) {
	mockRepo := new(mock.WalletRepositoryMock)
	mockSavingRepo := new(mock.SavingGoalRepositoryMock)
	walletService := service.NewWalletService(mockRepo, mockSavingRepo)

	wallet := &entity.Wallet{ID: 1, UserID: 1, Name: "My Wallet"}

	// Expect FindByID called with id=1, userID=1
	mockRepo.On("FindByID", uint(1), uint(1)).Return(wallet, nil)
	// Expect GetActiveContributions to be called
	mockSavingRepo.On("GetActiveContributions", uint(1)).Return(0.0, nil)

	result, err := walletService.GetWalletByID(1, 1)

	assert.NoError(t, err)
	assert.Equal(t, wallet, result)
}

func TestGetWalletByID_NotFound(t *testing.T) {
	mockRepo := new(mock.WalletRepositoryMock)
	mockSavingRepo := new(mock.SavingGoalRepositoryMock)
	walletService := service.NewWalletService(mockRepo, mockSavingRepo)

	// Simulate not found (e.g. belongs to other user)
	mockRepo.On("FindByID", uint(1), uint(1)).Return(nil, errors.New("record not found"))

	result, err := walletService.GetWalletByID(1, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "record not found", err.Error())
}
