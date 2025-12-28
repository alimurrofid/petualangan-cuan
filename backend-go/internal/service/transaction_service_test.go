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
	svc := service.NewTransactionService(mockRepo)

	mockData := []entity.Transaction{
		{Item: "Test Item", Amount: 10000, Date: time.Now()},
	}

	// Case 1: Success
	mockRepo.On("GetAll").Return(mockData, nil).Once()
	result, err := svc.GetTransactions()
	
	assert.NoError(t, err)
	assert.Equal(t, len(result), 1)
	assert.Equal(t, result[0].Item, "Test Item")
	mockRepo.AssertExpectations(t)

	// Case 2: Error
	mockRepo.On("GetAll").Return([]entity.Transaction{}, errors.New("db error")).Once()
	result, err = svc.GetTransactions()

	assert.Error(t, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}
