package service_test

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository/mock"
	"cuan-backend/internal/service"
	"testing"

	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateDebt(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&entity.Debt{}, &entity.Transaction{}, &entity.Wallet{}, &entity.Category{}, &entity.User{})

	mockRepo := new(mock.DebtRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	mockTxRepo := new(mock.TransactionRepositoryMock)

	svc := service.NewDebtService(mockRepo, mockTxRepo, mockWalletRepo, db)
	userID := uint(1)
	walletID := uint(1)

	db.Create(&entity.Wallet{ID: walletID, UserID: userID, Balance: 1000})

	input := service.CreateDebtInput{
		WalletID: walletID,
		Name:     "Test Debt",
		Amount:   500,
		Type:     "debt",
	}

	// Expectations
	// CreateDebt Calls: 
	// 1. debtRepo.WithTx(tx).Create(...)
	// 2. walletRepo.WithTx(tx).FindByID(...)
	// 3. walletRepo.WithTx(tx).Update(...) (if no error)
	// 4. transactionRepo.WithTx(tx).Create(...)

	mockRepo.On("WithTx", testMock.Anything).Return(mockRepo).Once()
	mockWalletRepo.On("WithTx", testMock.Anything).Return(mockWalletRepo).Twice()
	mockTxRepo.On("WithTx", testMock.Anything).Return(mockTxRepo).Once()

	mockRepo.On("Create", testMock.MatchedBy(func(d *entity.Debt) bool {
		return d.Name == "Test Debt" && d.Amount == 500
	})).Return(nil)

	mockWalletRepo.On("FindByID", walletID, userID).Return(&entity.Wallet{ID: walletID, UserID: userID, Balance: 1000}, nil)
	mockWalletRepo.On("Update", testMock.Anything).Return(nil)

	mockTxRepo.On("Create", testMock.Anything).Return(nil)

	debt, err := svc.CreateDebt(userID, input)

	assert.NoError(t, err)
	if assert.NotNil(t, debt) {
		assert.Equal(t, "Test Debt", debt.Name)
	}

	mockRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
	mockTxRepo.AssertExpectations(t)
}

func TestGetDebts(t *testing.T) {
	mockRepo := new(mock.DebtRepositoryMock)
	svc := service.NewDebtService(mockRepo, nil, nil, nil)
	userID := uint(1)

	mockRepo.On("FindByUserID", userID, "debt").Return([]entity.Debt{}, nil)

	_, err := svc.GetDebts(userID, "debt")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateDebt(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	
	mockRepo := new(mock.DebtRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	
	svc := service.NewDebtService(mockRepo, nil, mockWalletRepo, db)
	userID := uint(1)
	debtID := uint(1)
	walletID := uint(1)

	existingDebt := &entity.Debt{
		ID: debtID, UserID: userID, WalletID: walletID, Amount: 1000, Remaining: 500, Type: "debt",
	}
	
	input := service.UpdateDebtInput{
		WalletID: walletID,
		Name:     "Updated Debt",
		Amount:   2000,
	}

	// UpdateDebt Calls:
	// 1. debtRepo.WithTx.FindByID
    // 2. walletRepo.WithTx.FindByID
    // 3. walletRepo.WithTx.Update
    // 4. walletRepo.WithTx.FindByID
    // 5. walletRepo.WithTx.Update
    // 6. debtRepo.WithTx.Update

	mockRepo.On("WithTx", testMock.Anything).Return(mockRepo).Times(2)
	mockWalletRepo.On("WithTx", testMock.Anything).Return(mockWalletRepo).Times(4)

	mockRepo.On("FindByID", debtID, userID).Return(existingDebt, nil)
	
	// Wallet calls need to be carefully ordered or matched.
	// 1. Find Old Wallet
	mockWalletRepo.On("FindByID", walletID, userID).Return(&entity.Wallet{ID: walletID, Balance: 1000}, nil).Times(2)
	
	// 2. Update Old and New Wallet
	mockWalletRepo.On("Update", testMock.Anything).Return(nil).Times(2)

	mockRepo.On("Update", testMock.MatchedBy(func(d *entity.Debt) bool {
		return d.Amount == 2000 
	})).Return(nil)

	updatedDebt, err := svc.UpdateDebt(debtID, userID, input)
	
	assert.NoError(t, err)
	assert.Equal(t, 2000.0, updatedDebt.Amount)
	mockRepo.AssertExpectations(t)
}

func TestDeleteDebt(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&entity.DebtPayment{})
	
	mockRepo := new(mock.DebtRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)

	svc := service.NewDebtService(mockRepo, nil, mockWalletRepo, db)
	userID := uint(1)
	debtID := uint(1)
	walletID := uint(1)
	
	existingDebt := &entity.Debt{
		ID: debtID, UserID: userID, WalletID: walletID, Amount: 1000, Remaining: 500, Type: "debt",
	}

    // Calls:
    // 1. debtRepo.WithTx.FindByID
    // 2. walletRepo.WithTx.FindByID
    // 3. walletRepo.WithTx.Update
    // 4. debtRepo.WithTx.Delete

	mockRepo.On("WithTx", testMock.Anything).Return(mockRepo).Twice()
	mockWalletRepo.On("WithTx", testMock.Anything).Return(mockWalletRepo).Twice()

	mockRepo.On("FindByID", debtID, userID).Return(existingDebt, nil)
	mockWalletRepo.On("FindByID", walletID, userID).Return(&entity.Wallet{ID: walletID, Balance: 1000}, nil)
	
	mockWalletRepo.On("Update", testMock.Anything).Return(nil)
	
	mockRepo.On("Delete", debtID, userID).Return(nil)

	err := svc.DeleteDebt(debtID, userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestPayDebt(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&entity.Category{}, &entity.Transaction{}, &entity.DebtPayment{})

	mockRepo := new(mock.DebtRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	mockTxRepo := new(mock.TransactionRepositoryMock)

	svc := service.NewDebtService(mockRepo, mockTxRepo, mockWalletRepo, db)
	userID := uint(1)
	debtID := uint(1)
	walletID := uint(1)

	existingDebt := &entity.Debt{
		ID: debtID, UserID: userID, WalletID: walletID, Amount: 1000, Remaining: 1000, Type: "debt", IsPaid: false,
	}

	input := service.PayDebtInput{
		WalletID: walletID,
		Amount:   500,
		Note:     "Partial Payment",
	}

    // Calls:
    // 1. debtRepo.WithTx.FindByID
    // 2. debtRepo.WithTx.Update
    // 3. walletRepo.WithTx.FindByID
    // 4. walletRepo.WithTx.Update
    // 5. txRepo.WithTx.Create

	mockRepo.On("WithTx", testMock.Anything).Return(mockRepo).Twice()
	mockWalletRepo.On("WithTx", testMock.Anything).Return(mockWalletRepo).Twice()
	mockTxRepo.On("WithTx", testMock.Anything).Return(mockTxRepo).Once()

	mockRepo.On("FindByID", debtID, userID).Return(existingDebt, nil)
	
	mockRepo.On("Update", testMock.MatchedBy(func(d *entity.Debt) bool {
		return d.Remaining == 500
	})).Return(nil)

	mockWalletRepo.On("FindByID", walletID, userID).Return(&entity.Wallet{ID: walletID, Balance: 1000}, nil)
	mockWalletRepo.On("Update", testMock.Anything).Return(nil)

	mockTxRepo.On("Create", testMock.Anything).Return(nil)

	debt, err := svc.PayDebt(debtID, userID, input)

	assert.NoError(t, err)
	assert.Equal(t, 500.0, debt.Remaining)
	mockRepo.AssertExpectations(t)
}
