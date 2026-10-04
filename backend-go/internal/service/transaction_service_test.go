package service_test

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"
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

func setupTransactionTestDB(t *testing.T) (*gorm.DB, service.TransactionService, repository.WalletRepository) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&entity.Category{}, &entity.Transaction{}, &entity.Wallet{}, &entity.User{})
	assert.NoError(t, err)

	txRepo := repository.NewTransactionRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	svc := service.NewTransactionService(txRepo, walletRepo, db)

	return db, svc, walletRepo
}

func TestUpdateTransaction_ChangeWalletAndCategory(t *testing.T) {
	db, svc, walletRepo := setupTransactionTestDB(t)

	userID := uint(1)
	db.Create(&entity.User{ID: userID, Email: "wallet_cat_test@test.com"})
	db.Create(&entity.Wallet{ID: 10, UserID: userID, Name: "Wallet A", Balance: 500000})
	db.Create(&entity.Wallet{ID: 20, UserID: userID, Name: "Wallet B", Balance: 500000})
	db.Create(&entity.Category{ID: 100, UserID: userID, Name: "Food", Type: "expense"})
	db.Create(&entity.Category{ID: 200, UserID: userID, Name: "Transport", Type: "expense"})

	// 1. Create transaction in Wallet 10, Category 100, Amount 50,000 (expense)
	created, err := svc.CreateTransaction(userID, service.CreateTransactionInput{
		WalletID:    10,
		CategoryID:  100,
		Amount:      50000,
		Type:        "expense",
		Description: "Lunch",
		Date:        time.Now(),
	})
	assert.NoError(t, err)
	assert.Equal(t, uint(10), created.WalletID)
	assert.Equal(t, uint(100), created.CategoryID)

	// Check Wallet 10 balance deducted
	w1, err := walletRepo.FindByID(10, userID)
	assert.NoError(t, err)
	assert.Equal(t, float64(450000), w1.Balance)

	// 2. Update transaction to Wallet 20, Category 200, Amount 60,000 (expense)
	updated, err := svc.UpdateTransaction(created.ID, userID, service.CreateTransactionInput{
		WalletID:    20,
		CategoryID:  200,
		Amount:      60000,
		Type:        "expense",
		Description: "Taxi to office",
		Date:        time.Now(),
	})
	assert.NoError(t, err)

	// Verify transaction fields updated
	assert.Equal(t, uint(20), updated.WalletID, "WalletID must be updated to 20")
	assert.Equal(t, uint(200), updated.CategoryID, "CategoryID must be updated to 200")
	assert.Equal(t, float64(60000), updated.Amount)
	assert.Equal(t, "Wallet B", updated.Wallet.Name)
	assert.Equal(t, "Transport", updated.Category.Name)

	// Verify old wallet (10) balance restored back to 500,000
	w1, _ = walletRepo.FindByID(10, userID)
	assert.Equal(t, float64(500000), w1.Balance, "Old wallet balance must be restored")

	// Verify new wallet (20) balance deducted to 500,000 - 60,000 = 440,000
	w2, _ := walletRepo.FindByID(20, userID)
	assert.Equal(t, float64(440000), w2.Balance, "New wallet balance must be deducted")
}

func TestUpdateTransaction_ChangeOnlyCategory(t *testing.T) {
	db, svc, _ := setupTransactionTestDB(t)

	userID := uint(2)
	db.Create(&entity.User{ID: userID, Email: "cat_only_test@test.com"})
	db.Create(&entity.Wallet{ID: 30, UserID: userID, Name: "Cash", Balance: 200000})
	db.Create(&entity.Category{ID: 301, UserID: userID, Name: "Groceries", Type: "expense"})
	db.Create(&entity.Category{ID: 302, UserID: userID, Name: "Entertainment", Type: "expense"})

	created, err := svc.CreateTransaction(userID, service.CreateTransactionInput{
		WalletID:    30,
		CategoryID:  301,
		Amount:      30000,
		Type:        "expense",
		Description: "Snacks",
		Date:        time.Now(),
	})
	assert.NoError(t, err)

	// Update only Category from 301 to 302
	updated, err := svc.UpdateTransaction(created.ID, userID, service.CreateTransactionInput{
		WalletID:    30,
		CategoryID:  302,
		Amount:      30000,
		Type:        "expense",
		Description: "Snacks & Movie",
		Date:        time.Now(),
	})
	assert.NoError(t, err)

	assert.Equal(t, uint(302), updated.CategoryID, "CategoryID must be updated to 302")
	assert.Equal(t, "Entertainment", updated.Category.Name)
	assert.Equal(t, uint(30), updated.WalletID)
}

func TestUpdateTransaction_ChangeOnlyWallet(t *testing.T) {
	db, svc, walletRepo := setupTransactionTestDB(t)

	userID := uint(3)
	db.Create(&entity.User{ID: userID, Email: "wallet_only_test@test.com"})
	db.Create(&entity.Wallet{ID: 40, UserID: userID, Name: "Bank BCA", Balance: 1000000})
	db.Create(&entity.Wallet{ID: 50, UserID: userID, Name: "GoPay", Balance: 500000})
	db.Create(&entity.Category{ID: 401, UserID: userID, Name: "Salary", Type: "income"})

	// Create income 200,000 to Wallet 40
	created, err := svc.CreateTransaction(userID, service.CreateTransactionInput{
		WalletID:    40,
		CategoryID:  401,
		Amount:      200000,
		Type:        "income",
		Description: "Freelance",
		Date:        time.Now(),
	})
	assert.NoError(t, err)

	w40, _ := walletRepo.FindByID(40, userID)
	assert.Equal(t, float64(1200000), w40.Balance)

	// Move income transaction from Wallet 40 to Wallet 50
	updated, err := svc.UpdateTransaction(created.ID, userID, service.CreateTransactionInput{
		WalletID:    50,
		CategoryID:  401,
		Amount:      200000,
		Type:        "income",
		Description: "Freelance",
		Date:        time.Now(),
	})
	assert.NoError(t, err)

	assert.Equal(t, uint(50), updated.WalletID, "WalletID must be updated to 50")
	assert.Equal(t, "GoPay", updated.Wallet.Name)

	// Check Wallet 40 income reverted (1,200,000 - 200,000 = 1,000,000)
	w40, _ = walletRepo.FindByID(40, userID)
	assert.Equal(t, float64(1000000), w40.Balance)

	// Check Wallet 50 income applied (500,000 + 200,000 = 700,000)
	w50, _ := walletRepo.FindByID(50, userID)
	assert.Equal(t, float64(700000), w50.Balance)
}

