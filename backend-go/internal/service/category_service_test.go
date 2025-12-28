package service_test

import (
	"errors"
	"testing"

	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository/mock"
	"cuan-backend/internal/service"

	"github.com/stretchr/testify/assert"
	testifyMock "github.com/stretchr/testify/mock"
)

func TestCreateCategory(t *testing.T) {
	mockRepo := new(mock.CategoryRepositoryMock)
	svc := service.NewCategoryService(mockRepo)

	userID := uint(1)
	input := service.CreateCategoryInput{
		Name:        "Food",
		Type:        "expense",
		Icon:        "Pizza",
		BudgetLimit: 500000,
	}

	mockRepo.On("Create", testifyMock.AnythingOfType("*entity.Category")).Return(nil)

	result, err := svc.CreateCategory(userID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, input.Name, result.Name)
	mockRepo.AssertExpectations(t)
}

func TestGetCategories(t *testing.T) {
	mockRepo := new(mock.CategoryRepositoryMock)
	svc := service.NewCategoryService(mockRepo)

	userID := uint(1)
	mockData := []entity.Category{
		{Name: "Food", Type: "expense"},
		{Name: "Salary", Type: "income"},
	}

	mockRepo.On("FindAll", userID).Return(mockData, nil)

	result, err := svc.GetCategories(userID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Food", result[0].Name)
	mockRepo.AssertExpectations(t)
}

func TestGetCategory(t *testing.T) {
	mockRepo := new(mock.CategoryRepositoryMock)
	svc := service.NewCategoryService(mockRepo)

	userID := uint(1)
	catID := uint(1)
	mockCat := &entity.Category{Name: "Food", Type: "expense"}

	mockRepo.On("FindByID", catID, userID).Return(mockCat, nil)

	result, err := svc.GetCategory(catID, userID)

	assert.NoError(t, err)
	assert.Equal(t, "Food", result.Name)
	mockRepo.AssertExpectations(t)
}

func TestGetCategory_NotFound(t *testing.T) {
	mockRepo := new(mock.CategoryRepositoryMock)
	svc := service.NewCategoryService(mockRepo)

	userID := uint(1)
	catID := uint(99)

	mockRepo.On("FindByID", catID, userID).Return(nil, errors.New("not found"))

	result, err := svc.GetCategory(catID, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}
