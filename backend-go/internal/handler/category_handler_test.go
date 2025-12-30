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
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCategoryService struct {
	mock.Mock
}

func (m *MockCategoryService) CreateCategory(userID uint, input service.CreateCategoryInput) (*entity.Category, error) {
	args := m.Called(userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Category), args.Error(1)
}

func (m *MockCategoryService) GetCategories(userID uint) ([]entity.Category, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Category), args.Error(1)
}

func (m *MockCategoryService) GetCategory(id uint, userID uint) (*entity.Category, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Category), args.Error(1)
}

func (m *MockCategoryService) UpdateCategory(id uint, userID uint, input service.UpdateCategoryInput) (*entity.Category, error) {
	args := m.Called(id, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Category), args.Error(1)
}

func (m *MockCategoryService) DeleteCategory(id uint, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

func TestCreateCategory(t *testing.T) {
	mockService := new(MockCategoryService)
	categoryHandler := handler.NewCategoryHandler(mockService)

	app := fiber.New()
	app.Post("/api/categories", mockAuthMiddleware(1), categoryHandler.CreateCategory)

	input := service.CreateCategoryInput{
		Name: "Food",
		Type: "expense",
		Icon: "burger",
	}
	body, _ := json.Marshal(input)

	mockCategory := &entity.Category{ID: 1, UserID: 1, Name: "Food", Type: "expense"}
	mockService.On("CreateCategory", uint(1), input).Return(mockCategory, nil)

	req := httptest.NewRequest("POST", "/api/categories", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetCategories(t *testing.T) {
	mockService := new(MockCategoryService)
	categoryHandler := handler.NewCategoryHandler(mockService)

	app := fiber.New()
	app.Get("/api/categories", mockAuthMiddleware(1), categoryHandler.GetCategories)

	mockCategories := []entity.Category{
		{ID: 1, UserID: 1, Name: "Food", Type: "expense"},
		{ID: 2, UserID: 1, Name: "Salary", Type: "income"},
	}
	mockService.On("GetCategories", uint(1)).Return(mockCategories, nil)

	req := httptest.NewRequest("GET", "/api/categories", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetCategory(t *testing.T) {
	mockService := new(MockCategoryService)
	categoryHandler := handler.NewCategoryHandler(mockService)

	app := fiber.New()
	app.Get("/api/categories/:id", mockAuthMiddleware(1), categoryHandler.GetCategory)

	mockCategory := &entity.Category{ID: 1, UserID: 1, Name: "Food", Type: "expense"}
	mockService.On("GetCategory", uint(1), uint(1)).Return(mockCategory, nil)

	req := httptest.NewRequest("GET", "/api/categories/1", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetCategory_NotFound(t *testing.T) {
	mockService := new(MockCategoryService)
	categoryHandler := handler.NewCategoryHandler(mockService)

	app := fiber.New()
	app.Get("/api/categories/:id", mockAuthMiddleware(1), categoryHandler.GetCategory)

	mockService.On("GetCategory", uint(999), uint(1)).Return(nil, errors.New("category not found"))

	req := httptest.NewRequest("GET", "/api/categories/999", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestUpdateCategory(t *testing.T) {
	mockService := new(MockCategoryService)
	categoryHandler := handler.NewCategoryHandler(mockService)

	app := fiber.New()
	app.Put("/api/categories/:id", mockAuthMiddleware(1), categoryHandler.UpdateCategory)

	input := service.UpdateCategoryInput{
		Name: "Food & Drinks",
	}
	body, _ := json.Marshal(input)

	mockCategory := &entity.Category{ID: 1, UserID: 1, Name: "Food & Drinks", Type: "expense"}
	mockService.On("UpdateCategory", uint(1), uint(1), input).Return(mockCategory, nil)

	req := httptest.NewRequest("PUT", "/api/categories/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestDeleteCategory(t *testing.T) {
	mockService := new(MockCategoryService)
	categoryHandler := handler.NewCategoryHandler(mockService)

	app := fiber.New()
	app.Delete("/api/categories/:id", mockAuthMiddleware(1), categoryHandler.DeleteCategory)

	mockService.On("DeleteCategory", uint(1), uint(1)).Return(nil)

	req := httptest.NewRequest("DELETE", "/api/categories/1", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}
