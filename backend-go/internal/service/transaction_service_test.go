package service_test

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository/mock"
	"cuan-backend/internal/service"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGetTransactions(t *testing.T) {
	mockRepo := new(mock.TransactionRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	svc := service.NewTransactionService(mockRepo, mockWalletRepo, nil)
	userID := uint(1)

	mockData := []entity.Transaction{
		{Description: "Test Item", Amount: 10000, Date: time.Now()},
	}

	params := entity.TransactionFilterParams{
		Page: 1, Limit: 10, Search: "Item",
	}

	mockRepo.On("FindAll", userID, params).Return(mockData, int64(1), nil).Once()
	result, total, err := svc.GetTransactions(userID, params)
	
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, result[0].Description, "Test Item")
	mockRepo.AssertExpectations(t)

	mockRepo.On("FindAll", userID, params).Return([]entity.Transaction{}, int64(0), errors.New("db error")).Once()
	result, _, err = svc.GetTransactions(userID, params)

	assert.Error(t, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}

func TestRaceConditionCategoryTransfer(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)
	
	err = db.AutoMigrate(&entity.Category{}, &entity.Transaction{}, &entity.Wallet{}, &entity.User{})
	assert.NoError(t, err)

	userID := uint(1)
	db.Create(&entity.User{ID: userID, Email: "test@test.com"})
	db.Create(&entity.Wallet{ID: 1, UserID: userID, Balance: 100000})
	db.Create(&entity.Wallet{ID: 2, UserID: userID, Balance: 0})

	mockRepo := new(mock.TransactionRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	mockWalletRepo.On("FindByID", uint(1), userID).Return(&entity.Wallet{ID: 1, UserID: userID, Balance: 100000}, nil)
	mockWalletRepo.On("FindByID", uint(2), userID).Return(&entity.Wallet{ID: 2, UserID: userID, Balance: 0}, nil)
	
	mockRepo.On("WithTx", testMock.Anything).Return(mockRepo)
	mockRepo.On("Create", testMock.Anything).Return(nil)

	svc := service.NewTransactionService(mockRepo, mockWalletRepo, db)

	var wg sync.WaitGroup
	concurrency := 10

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			input := service.TransferTransactionInput{
				FromWalletID: 1,
				ToWalletID:   2,
				Amount:       100,
				Date:         time.Now(),
			}
			_ = svc.TransferTransaction(userID, input)
		}()
	}
	wg.Wait()
	var count int64
	db.Model(&entity.Category{}).Where("user_id = ? AND type = ?", userID, "transfer").Count(&count)
	assert.Equal(t, int64(1), count, "Should only have 1 Transfer category")
}
