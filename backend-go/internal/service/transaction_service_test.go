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

// TestGetTransactions covers SuccessGetTransactions and FilterLogicVerification (via mock expectations)
func TestGetTransactions(t *testing.T) {
	mockRepo := new(mock.TransactionRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	svc := service.NewTransactionService(mockRepo, mockWalletRepo, nil)
	userID := uint(1)

	mockData := []entity.Transaction{
		{Description: "Test Item", Amount: 10000, Date: time.Now()},
	}

	// Case 1: Success with Filters
	params := entity.TransactionFilterParams{
		Page: 1, Limit: 10, Search: "Item",
	}
	// Verify that repo is called with correct params
	mockRepo.On("FindAll", userID, params).Return(mockData, int64(1), nil).Once()
	result, total, err := svc.GetTransactions(userID, params)
	
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, result[0].Description, "Test Item")
	mockRepo.AssertExpectations(t)

	// Case 2: Error
	mockRepo.On("FindAll", userID, params).Return([]entity.Transaction{}, int64(0), errors.New("db error")).Once()
	result, _, err = svc.GetTransactions(userID, params)

	assert.Error(t, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}

// TestRaceConditionCategoryTransfer tests that concurrent requests don't create duplicates.
// Since we use FirstOrCreate in the service which relies on DB, we need a real DB or a very smart mock.
// Mocks are hard for concurrency on the same method call if not designed for it.
// However, the service uses `s.db.Where(...).FirstOrCreate(...)`.
// We can use an in-memory SQLite DB for this test to truly verify GORM behavior, 
// as mocking GORM's chaining for FirstOrCreate is painful and error-prone.
func TestRaceConditionCategoryTransfer(t *testing.T) {
	// Setup in-memory DB
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)
	
	// Migrate necessary tables
	err = db.AutoMigrate(&entity.Category{}, &entity.Transaction{}, &entity.Wallet{}, &entity.User{})
	assert.NoError(t, err)

	// Create User and Wallets
	userID := uint(1)
	db.Create(&entity.User{ID: userID, Email: "test@test.com"})
	db.Create(&entity.Wallet{ID: 1, UserID: userID, Balance: 100000})
	db.Create(&entity.Wallet{ID: 2, UserID: userID, Balance: 0})

	// Setup Service with real DB
	// We need mocks for Repos, but the transfer logic uses `s.db` directly for category creation
	mockRepo := new(mock.TransactionRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	// We need to implement FindByID for wallet in mock because TransferTransaction calls it
	mockWalletRepo.On("FindByID", uint(1), userID).Return(&entity.Wallet{ID: 1, UserID: userID, Balance: 100000}, nil)
	mockWalletRepo.On("FindByID", uint(2), userID).Return(&entity.Wallet{ID: 2, UserID: userID, Balance: 0}, nil)
	
	// Create is called on Repo
	mockRepo.On("WithTx", testMock.Anything).Return(mockRepo) // Chain
	mockRepo.On("Create", testMock.Anything).Return(nil)

	svc := service.NewTransactionService(mockRepo, mockWalletRepo, db)

	// Run concurrent transfers
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
			// We ignore error here as we focus on category creation
			_ = svc.TransferTransaction(userID, input)
		}()
	}
	wg.Wait()

	// Verify only 1 "Transfer" category exists
	var count int64
	db.Model(&entity.Category{}).Where("user_id = ? AND type = ?", userID, "transfer").Count(&count)
	assert.Equal(t, int64(1), count, "Should only have 1 Transfer category")
}
