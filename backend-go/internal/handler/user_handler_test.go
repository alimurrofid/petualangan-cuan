package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"cuan-backend/internal/handler"
	"cuan-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock of UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(input service.RegisterInput) (string, error) {
	args := m.Called(input)
	return args.String(0), args.Error(1)
}

func (m *MockUserService) Login(input service.LoginInput) (string, error) {
	args := m.Called(input)
	return args.String(0), args.Error(1)
}

func (m *MockUserService) Logout(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func TestRegisterHandler(t *testing.T) {
	mockService := new(MockUserService)
	userHandler := handler.NewUserHandler(mockService)

	app := fiber.New()
	app.Post("/auth/register", userHandler.Register)

	input := service.RegisterInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(input)

	mockService.On("Register", input).Return("mock_token", nil)

	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestLoginHandler(t *testing.T) {
	mockService := new(MockUserService)
	userHandler := handler.NewUserHandler(mockService)

	app := fiber.New()
	app.Post("/auth/login", userHandler.Login)

	input := service.LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(input)

	mockService.On("Login", input).Return("mock_token", nil)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	mockService := new(MockUserService)
	userHandler := handler.NewUserHandler(mockService)

	app := fiber.New()
	app.Post("/auth/login", userHandler.Login)

	input := service.LoginInput{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	body, _ := json.Marshal(input)

	mockService.On("Login", input).Return("", errors.New("invalid credentials"))

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	mockService.AssertExpectations(t)
}
