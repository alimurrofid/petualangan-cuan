package service_test

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository/mock"
	"cuan-backend/internal/service"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateGoal(t *testing.T) {
	mockRepo := new(mock.SavingGoalRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	
	svc := service.NewSavingGoalService(mockRepo, mockWalletRepo, nil, db)
	userID := uint(1)

	input := service.CreateGoalInput{
		Name:         "Buy Car",
		TargetAmount: 100000,
		CategoryID:   1,
		Icon:         "Car",
	}

	mockRepo.On("Create", testMock.AnythingOfType("*entity.SavingGoal")).Return(nil)

	goal, err := svc.CreateGoal(userID, input)

	assert.NoError(t, err)
	assert.NotNil(t, goal)
	assert.Equal(t, input.Name, goal.Name)
	assert.Equal(t, userID, goal.UserID)
	mockRepo.AssertExpectations(t)
}

func TestGetGoals(t *testing.T) {
	mockRepo := new(mock.SavingGoalRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	
	svc := service.NewSavingGoalService(mockRepo, mockWalletRepo, nil, db)
	userID := uint(1)

	mockGoals := []entity.SavingGoal{
		{UserID: userID, Name: "Goal 1", TargetAmount: 5000},
	}

	mockRepo.On("FindAll", userID).Return(mockGoals, nil)

	goals, err := svc.GetGoals(userID)

	assert.NoError(t, err)
	assert.Len(t, goals, 1)
	assert.Equal(t, "Goal 1", goals[0].Name)
	mockRepo.AssertExpectations(t)
}

func TestUpdateGoal(t *testing.T) {
	mockRepo := new(mock.SavingGoalRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	
	svc := service.NewSavingGoalService(mockRepo, mockWalletRepo, nil, db)
	userID := uint(1)
	goalID := uint(1)

	existingGoal := &entity.SavingGoal{
		ID: goalID, UserID: userID, Name: "Old Name", TargetAmount: 1000,
	}

	input := service.CreateGoalInput{
		Name:         "New Name",
		TargetAmount: 2000,
		CategoryID:   1,
	}

	mockRepo.On("FindByID", goalID, userID).Return(existingGoal, nil)
	mockRepo.On("Update", testMock.MatchedBy(func(g *entity.SavingGoal) bool {
		return g.Name == "New Name" && g.TargetAmount == 2000
	})).Return(nil)

	updatedGoal, err := svc.UpdateGoal(userID, goalID, input)

	assert.NoError(t, err)
	assert.Equal(t, "New Name", updatedGoal.Name)
	mockRepo.AssertExpectations(t)
}

func TestDeleteGoal(t *testing.T) {
	mockRepo := new(mock.SavingGoalRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	
	svc := service.NewSavingGoalService(mockRepo, mockWalletRepo, nil, db)
	userID := uint(1)
	goalID := uint(1)

	existingGoal := &entity.SavingGoal{ID: goalID, UserID: userID}

	mockRepo.On("FindByID", goalID, userID).Return(existingGoal, nil)
	mockRepo.On("DeleteContributions", goalID).Return(nil)
	mockRepo.On("Delete", existingGoal).Return(nil)

	err := svc.DeleteGoal(userID, goalID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteGoal_NotFound(t *testing.T) {
	mockRepo := new(mock.SavingGoalRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	
	svc := service.NewSavingGoalService(mockRepo, mockWalletRepo, nil, db)
	userID := uint(1)
	goalID := uint(1)

	mockRepo.On("FindByID", goalID, userID).Return(nil, errors.New("not found"))

	err := svc.DeleteGoal(userID, goalID)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAddContribution(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	db.AutoMigrate(&entity.SavingGoal{}, &entity.Transaction{}, &entity.SavingContribution{}, &entity.Category{}, &entity.Wallet{})
	
	userID := uint(1)
	goalID := uint(1)
	walletID := uint(1)
	
	db.Create(&entity.SavingGoal{ID: goalID, UserID: userID, Name: "Goal", TargetAmount: 1000, CurrentAmount: 0})
	db.Create(&entity.Wallet{ID: walletID, UserID: userID, Balance: 5000})
	
	mockRepo := new(mock.SavingGoalRepositoryMock)
	
	mockWalletRepo := new(mock.WalletRepositoryMock)
	
	svc := service.NewSavingGoalService(mockRepo, mockWalletRepo, nil, db)
	
	input := service.ContributionInput{
		WalletID: walletID,
		Amount:   500,
		Date:     time.Now(),
	}
	
	mockGoal := &entity.SavingGoal{
		ID: goalID, UserID: userID, Name: "My Goal", TargetAmount: 1000, CurrentAmount: 0,
	}
	
	mockRepo.On("FindByID", goalID, userID).Return(mockGoal, nil)
    
    mockWalletRepo.On("WithTx", testMock.Anything).Return(mockWalletRepo)
    mockWalletRepo.On("FindByID", walletID, userID).Return(&entity.Wallet{ID: walletID, UserID: userID, Balance: 5000}, nil)
	
	contribution, err := svc.AddContribution(userID, goalID, input)
	
	assert.NoError(t, err)
	assert.NotNil(t, contribution)
	assert.Equal(t, 500.0, contribution.Amount)
	
	var updatedGoal entity.SavingGoal
	db.First(&updatedGoal, goalID)
	
	assert.Equal(t, 500.0, updatedGoal.CurrentAmount)
}

func TestFinishGoal_Success(t *testing.T) {
	mockRepo := new(mock.SavingGoalRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})

	svc := service.NewSavingGoalService(mockRepo, mockWalletRepo, nil, db)
	userID := uint(1)
	goalID := uint(1)

	existingGoal := &entity.SavingGoal{
		ID: goalID, UserID: userID, IsAchieved: true, IsFinished: false,
	}

	mockRepo.On("FindByID", goalID, userID).Return(existingGoal, nil)
	mockRepo.On("Update", testMock.MatchedBy(func(g *entity.SavingGoal) bool {
		return g.ID == goalID && g.IsFinished == true
	})).Return(nil)

	err := svc.FinishGoal(userID, goalID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestFinishGoal_Error_NotFound(t *testing.T) {
	mockRepo := new(mock.SavingGoalRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})

	svc := service.NewSavingGoalService(mockRepo, mockWalletRepo, nil, db)
	userID := uint(1)
	goalID := uint(1)

	mockRepo.On("FindByID", goalID, userID).Return(nil, errors.New("not found"))

	err := svc.FinishGoal(userID, goalID)

	assert.Error(t, err)
	assert.Equal(t, "goal not found", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestFinishGoal_Error_NotAchieved(t *testing.T) {
	mockRepo := new(mock.SavingGoalRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})

	svc := service.NewSavingGoalService(mockRepo, mockWalletRepo, nil, db)
	userID := uint(1)
	goalID := uint(1)

	existingGoal := &entity.SavingGoal{
		ID: goalID, UserID: userID, IsAchieved: false, IsFinished: false,
	}

	mockRepo.On("FindByID", goalID, userID).Return(existingGoal, nil)

	err := svc.FinishGoal(userID, goalID)

	assert.Error(t, err)
	assert.Equal(t, "goal is not achieved yet", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestFinishGoal_Error_AlreadyFinished(t *testing.T) {
	mockRepo := new(mock.SavingGoalRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})

	svc := service.NewSavingGoalService(mockRepo, mockWalletRepo, nil, db)
	userID := uint(1)
	goalID := uint(1)

	existingGoal := &entity.SavingGoal{
		ID: goalID, UserID: userID, IsAchieved: true, IsFinished: true,
	}

	mockRepo.On("FindByID", goalID, userID).Return(existingGoal, nil)

	err := svc.FinishGoal(userID, goalID)

	assert.Error(t, err)
	assert.Equal(t, "goal is already finished", err.Error())
	mockRepo.AssertExpectations(t)
}
