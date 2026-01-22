package service_test

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository/mock"
	"cuan-backend/internal/service"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
)

func TestCreateWishlist(t *testing.T) {
	mockRepo := new(mock.WishlistRepositoryMock)
	svc := service.NewWishlistService(mockRepo)
	userID := uint(1)

	req := &service.StoreWishlistRequest{
		Name:           "Item 1",
		EstimatedPrice: 1000,
		CategoryID:     1,
		Priority:       "high",
	}

	mockRepo.On("Create", testMock.MatchedBy(func(item *entity.WishlistItem) bool {
		return item.Name == req.Name && item.Priority == entity.WishlistPriorityHigh
	})).Return(nil)

	err := svc.Create(userID, req)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestFindAllWishlist(t *testing.T) {
	mockRepo := new(mock.WishlistRepositoryMock)
	svc := service.NewWishlistService(mockRepo)
	userID := uint(1)

	expectedItems := []entity.WishlistItem{
		{UserID: userID, Name: "Item 1"},
	}

	mockRepo.On("FindAllByUserID", userID).Return(expectedItems, nil)

	items, err := svc.FindAll(userID)

	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "Item 1", items[0].Name)
	mockRepo.AssertExpectations(t)
}

func TestFindWishlistByID(t *testing.T) {
	mockRepo := new(mock.WishlistRepositoryMock)
	svc := service.NewWishlistService(mockRepo)
	userID := uint(1)
	id := uint(1)

	expectedItem := &entity.WishlistItem{ID: id, UserID: userID, Name: "Item 1"}

	mockRepo.On("FindByID", id, userID).Return(expectedItem, nil)

	item, err := svc.FindByID(id, userID)

	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, "Item 1", item.Name)
	mockRepo.AssertExpectations(t)
}

func TestUpdateWishlist(t *testing.T) {
	mockRepo := new(mock.WishlistRepositoryMock)
	svc := service.NewWishlistService(mockRepo)
	userID := uint(1)
	id := uint(1)

	req := &service.StoreWishlistRequest{
		Name:           "Updated Item",
		EstimatedPrice: 2000,
		CategoryID:     1,
		Priority:       "medium",
	}

	existingItem := &entity.WishlistItem{ID: id, UserID: userID, Name: "Item 1", Priority: entity.WishlistPriorityLow}

	mockRepo.On("FindByID", id, userID).Return(existingItem, nil)
	mockRepo.On("Update", testMock.MatchedBy(func(item *entity.WishlistItem) bool {
		return item.Name == "Updated Item" && item.Priority == entity.WishlistPriorityMedium
	})).Return(nil)

	err := svc.Update(id, userID, req)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteWishlist(t *testing.T) {
	mockRepo := new(mock.WishlistRepositoryMock)
	svc := service.NewWishlistService(mockRepo)
	userID := uint(1)
	id := uint(1)

	mockRepo.On("Delete", id, userID).Return(nil)

	err := svc.Delete(id, userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMarkAsBought(t *testing.T) {
	mockRepo := new(mock.WishlistRepositoryMock)
	svc := service.NewWishlistService(mockRepo)
	userID := uint(1)
	id := uint(1)

	mockRepo.On("MarkAsBought", id, userID).Return(nil)

	err := svc.MarkAsBought(id, userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateWishlist_NotFound(t *testing.T) {
	mockRepo := new(mock.WishlistRepositoryMock)
	svc := service.NewWishlistService(mockRepo)
	userID := uint(1)
	id := uint(1)
	req := &service.StoreWishlistRequest{}

	mockRepo.On("FindByID", id, userID).Return(nil, errors.New("not found"))

	err := svc.Update(id, userID, req)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}
