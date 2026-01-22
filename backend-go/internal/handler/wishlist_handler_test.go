package handler_test

import (
	"bytes"
	"cuan-backend/internal/entity"
	"cuan-backend/internal/handler"
	"cuan-backend/internal/service"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWishlistService is a mock implementation of service.WishlistService
type MockWishlistService struct {
	mock.Mock
}

func (m *MockWishlistService) Create(userID uint, req *service.StoreWishlistRequest) error {
	args := m.Called(userID, req)
	return args.Error(0)
}

func (m *MockWishlistService) FindAll(userID uint) ([]entity.WishlistItem, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.WishlistItem), args.Error(1)
}

func (m *MockWishlistService) FindByID(id uint, userID uint) (*entity.WishlistItem, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WishlistItem), args.Error(1)
}

func (m *MockWishlistService) Update(id uint, userID uint, req *service.StoreWishlistRequest) error {
	args := m.Called(id, userID, req)
	return args.Error(0)
}

func (m *MockWishlistService) Delete(id uint, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

func (m *MockWishlistService) MarkAsBought(id uint, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

// Handler tests

func TestCreateWishlist_Handler(t *testing.T) {
	mockService := new(MockWishlistService)
	h := handler.NewWishlistHandler(mockService)

	app := fiber.New()
	app.Post("/api/wishlists", mockAuthMiddleware(1), h.Create)

	input := service.StoreWishlistRequest{
		Name:           "New Item",
		EstimatedPrice: 500,
		CategoryID:     1,
		Priority:       "high",
	}
	body, _ := json.Marshal(input)

	mockService.On("Create", uint(1), mock.MatchedBy(func(req *service.StoreWishlistRequest) bool {
		return req.Name == input.Name && req.EstimatedPrice == input.EstimatedPrice
	})).Return(nil)

	req := httptest.NewRequest("POST", "/api/wishlists", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestFindAllWishlist_Handler(t *testing.T) {
	mockService := new(MockWishlistService)
	h := handler.NewWishlistHandler(mockService)

	app := fiber.New()
	app.Get("/api/wishlists", mockAuthMiddleware(1), h.FindAll)

	expectedItems := []entity.WishlistItem{
		{UserID: 1, Name: "Item 1"},
	}

	mockService.On("FindAll", uint(1)).Return(expectedItems, nil)

	req := httptest.NewRequest("GET", "/api/wishlists", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestFindWishlistByID_Handler(t *testing.T) {
	mockService := new(MockWishlistService)
	h := handler.NewWishlistHandler(mockService)

	app := fiber.New()
	app.Get("/api/wishlists/:id", mockAuthMiddleware(1), h.FindByID)

	id := 1
	expectedItem := &entity.WishlistItem{ID: uint(id), UserID: 1, Name: "Item 1"}

	mockService.On("FindByID", uint(id), uint(1)).Return(expectedItem, nil)

	req := httptest.NewRequest("GET", "/api/wishlists/"+strconv.Itoa(id), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestUpdateWishlist_Handler(t *testing.T) {
	mockService := new(MockWishlistService)
	h := handler.NewWishlistHandler(mockService)

	app := fiber.New()
	app.Put("/api/wishlists/:id", mockAuthMiddleware(1), h.Update)

	id := 1
	input := service.StoreWishlistRequest{
		Name:           "Updated Item",
		EstimatedPrice: 600,
	}
	body, _ := json.Marshal(input)

	mockService.On("Update", uint(id), uint(1), mock.MatchedBy(func(req *service.StoreWishlistRequest) bool {
		return req.Name == "Updated Item"
	})).Return(nil)

	req := httptest.NewRequest("PUT", "/api/wishlists/"+strconv.Itoa(id), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestDeleteWishlist_Handler(t *testing.T) {
	mockService := new(MockWishlistService)
	h := handler.NewWishlistHandler(mockService)

	app := fiber.New()
	app.Delete("/api/wishlists/:id", mockAuthMiddleware(1), h.Delete)

	id := 1

	mockService.On("Delete", uint(id), uint(1)).Return(nil)

	req := httptest.NewRequest("DELETE", "/api/wishlists/"+strconv.Itoa(id), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestMarkAsBought_Handler(t *testing.T) {
	mockService := new(MockWishlistService)
	h := handler.NewWishlistHandler(mockService)

	app := fiber.New()
	app.Patch("/api/wishlists/:id/bought", mockAuthMiddleware(1), h.MarkAsBought)

	id := 1

	mockService.On("MarkAsBought", uint(id), uint(1)).Return(nil)

	req := httptest.NewRequest("PATCH", "/api/wishlists/"+strconv.Itoa(id)+"/bought", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestFindWishlistByID_Handler_NotFound(t *testing.T) {
	mockService := new(MockWishlistService)
	h := handler.NewWishlistHandler(mockService)

	app := fiber.New()
	app.Get("/api/wishlists/:id", mockAuthMiddleware(1), h.FindByID)

	id := 99

	mockService.On("FindByID", uint(id), uint(1)).Return(nil, errors.New("not found"))

	req := httptest.NewRequest("GET", "/api/wishlists/"+strconv.Itoa(id), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	mockService.AssertExpectations(t)
}
