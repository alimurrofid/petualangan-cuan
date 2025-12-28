package service_test

import (
	"errors"
	"testing"
	"time"

	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository/mock"
	"cuan-backend/internal/service"

	"github.com/stretchr/testify/assert"
)

func TestGetTransactions(t *testing.T) {
	mockRepo := new(mock.TransactionRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	// Passing nil for DB as it's not used in GetTransactions, but providing WalletRepo mock
	svc := service.NewTransactionService(mockRepo, mockWalletRepo, nil)
	userID := uint(1)

	mockData := []entity.Transaction{
		{Description: "Test Item", Amount: 10000, Date: time.Now()},
	}

	// Case 1: Success
	mockRepo.On("FindAll", userID).Return(mockData, nil).Once()
	result, err := svc.GetTransactions(userID)
	
	assert.NoError(t, err)
	assert.Equal(t, len(result), 1)
	assert.Equal(t, result[0].Description, "Test Item")
	mockRepo.AssertExpectations(t)

	// Case 2: Error
	mockRepo.On("FindAll", userID).Return([]entity.Transaction{}, errors.New("db error")).Once()
	result, err = svc.GetTransactions(userID)

	assert.Error(t, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}
