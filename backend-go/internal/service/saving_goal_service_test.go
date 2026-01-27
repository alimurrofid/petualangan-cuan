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
	
	// Create in-memory DB for service struct requirement
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
	// AddContribution is complex because it uses a transaction and creates categories if not exist.
	// For full coverage we might need the SQLite in-memory approach used in TransactionService test,
	// because `s.db.Begin()` creates a real GORM transaction which is hard to mock perfectly with just unit mocks
	// unless we abstract the DB layer entirely (which isn't done here).
	
	// Setup in-memory DB
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	db.AutoMigrate(&entity.SavingGoal{}, &entity.Transaction{}, &entity.SavingContribution{}, &entity.Category{}, &entity.Wallet{})
	
	userID := uint(1)
	goalID := uint(1)
	walletID := uint(1)
	
	// Seed Data
	db.Create(&entity.SavingGoal{ID: goalID, UserID: userID, Name: "Goal", TargetAmount: 1000, CurrentAmount: 0})
	db.Create(&entity.Wallet{ID: walletID, UserID: userID, Balance: 5000})

	// For this test we use the real DB for the service to handle the transaction logic,
	// BUT CreateGoalInput logic mixes Repos and DB.
	// The Service uses `s.repo.FindByID` effectively.
	// If we provide the MOCK repo, `s.repo` calls go to mock.
	// But `s.db.Begin` goes to real DB.
	// `Transaction` creation uses `tx.Create` (Real DB).
	// `SavingContribution` uses `tx.Create` (Real DB).
	// `Goal` update uses `tx.Save` (Real DB).
	
	// Issue: `s.repo.FindByID` is called. If we mock it, it returns an object.
	// Then `tx.Create` tries to save things.
	
	// To test `AddContribution` properly with this mixed architecture (Service using Repository + Raw GORM TX),
	// we should probably use a functional test with SQLite for everything, OR minimal mocks.
	
	// Let's use Real Repository implementation over SQLite for this specific test? 
	// Or we just Mock `FindByID` to return a struct, and let GORM handle the rest?
	// The problem is `FindByID` returns `*entity.SavingGoal`. If that struct is not attached to the DB session, `tx.Save(goal)` might try to insert it or update it based on ID.
	
	// Actually, `AddContribution` calls `s.repo.FindByID(goalID, userID)`.
	// Then it does `tx.Save(goal)`.
	// If `goal` comes from Mock, it's just a struct. GORM `tx.Save` will try to update it.
	
	mockRepo := new(mock.SavingGoalRepositoryMock)
	// We need to configure the mock to return a goal that GORM can "Save".
	// Since it's just an ID based update effectively, it should work fine with `tx.Save`.
	
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
    
    // Fix: Add missing WalletRepo expectations
    mockWalletRepo.On("WithTx", testMock.Anything).Return(mockWalletRepo)
    mockWalletRepo.On("FindByID", walletID, userID).Return(&entity.Wallet{ID: walletID, UserID: userID, Balance: 5000}, nil)
	
	// Execute
	contribution, err := svc.AddContribution(userID, goalID, input)
	
	if err != nil {
		// If it failed, it might be due to `tx.Create(transaction)` failing because FKs (Category) don't exist in SQLite.
		// The service creates "Tabungan" category if not found.
		// So it should work if SQLite is migrated.
	}
	
	assert.NoError(t, err)
	assert.NotNil(t, contribution)
	assert.Equal(t, 500.0, contribution.Amount)
	
	// Verify goal was updated in DB (since we used Real DB for Save)
	var updatedGoal entity.SavingGoal
	db.First(&updatedGoal, goalID)
	// Wait.. we mocked `FindByID`, but `tx.Save` wrote to the DB.
	// But did we insert the initial goal into the DB? No, we only Seeded it above.
	// Wait, if `mockRepo` returns a goal that exists in DB (because we seeded it), `tx.Save` will update it.
	// If `mockRepo` returns a goal that DOES NOT exist in DB, `tx.Save` might try to INSERT it.
	// Since we seeded `goalID=1` in the DB, and Mock returns `ID=1`, `tx.Save` should update the existing record.
	
	assert.Equal(t, 500.0, updatedGoal.CurrentAmount)
}

func TestFinishGoal_Success(t *testing.T) {
	mockRepo := new(mock.SavingGoalRepositoryMock)
	mockWalletRepo := new(mock.WalletRepositoryMock)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})

	svc := service.NewSavingGoalService(mockRepo, mockWalletRepo, nil, db)
	userID := uint(1)
	goalID := uint(1)

	// Goal is achieved and not finished
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

	// Goal is NOT achieved
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

	// Goal IS already finished
	existingGoal := &entity.SavingGoal{
		ID: goalID, UserID: userID, IsAchieved: true, IsFinished: true,
	}

	mockRepo.On("FindByID", goalID, userID).Return(existingGoal, nil)

	err := svc.FinishGoal(userID, goalID)

	assert.Error(t, err)
	assert.Equal(t, "goal is already finished", err.Error())
	mockRepo.AssertExpectations(t)
}
