package service_test

import (
	"errors"
	"os"
	"testing"

	"cuan-backend/internal/entity"
	"cuan-backend/internal/service"

	"cuan-backend/internal/repository/mock"

	"github.com/stretchr/testify/assert"
	testifyMock "github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	os.Setenv("JWT_SECRET", "test_secret_key")
}

func TestRegister(t *testing.T) {
	mockRepo := new(mock.UserRepositoryMock)
	userService := service.NewUserService(mockRepo)

	input := service.RegisterInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	mockRepo.On("Create", testifyMock.AnythingOfType("*entity.User")).Return(nil)

	user, token, err := userService.Register(input)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotNil(t, user)
	assert.Equal(t, input.Name, user.Name)
	mockRepo.AssertExpectations(t)
}

func TestLogin(t *testing.T) {
	mockRepo := new(mock.UserRepositoryMock)
	userService := service.NewUserService(mockRepo)

	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &entity.User{
		ID:       1,
		Name:     "Test User",
		Email:    "test@example.com",
		Password: string(hashedPassword),
	}

	mockRepo.On("FindByEmail", "test@example.com").Return(user, nil)

	input := service.LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	}

	userRes, token, err := userService.Login(input)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotNil(t, userRes)
	assert.Equal(t, user.Email, userRes.Email)
	mockRepo.AssertExpectations(t)
}

func TestLogin_InvalidPassword(t *testing.T) {
	mockRepo := new(mock.UserRepositoryMock)
	userService := service.NewUserService(mockRepo)

	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &entity.User{
		ID:       1,
		Name:     "Test User",
		Email:    "test@example.com",
		Password: string(hashedPassword),
	}

	mockRepo.On("FindByEmail", "test@example.com").Return(user, nil)

	input := service.LoginInput{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	userRes, token, err := userService.Login(input)

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Nil(t, userRes)
	assert.Equal(t, "invalid email or password", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestLogin_UserNotFound(t *testing.T) {
	mockRepo := new(mock.UserRepositoryMock)
	userService := service.NewUserService(mockRepo)

	mockRepo.On("FindByEmail", "notfound@example.com").Return(nil, errors.New("record not found"))

	input := service.LoginInput{
		Email:    "notfound@example.com",
		Password: "password123",
	}

	userRes, token, err := userService.Login(input)

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Nil(t, userRes)
	assert.Equal(t, "invalid email or password", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestChangePassword(t *testing.T) {
	mockRepo := new(mock.UserRepositoryMock)
	userService := service.NewUserService(mockRepo)

	// Setup initial user with a password
	initialPassword := "oldpassword"
	hashedInitialPassword, _ := bcrypt.GenerateFromPassword([]byte(initialPassword), bcrypt.DefaultCost)
	
	user := &entity.User{
		ID:       1,
		Name:     "Test User",
		Email:    "test@example.com",
		Password: string(hashedInitialPassword),
	}

	// Mocking FindByID
	mockRepo.On("FindByID", uint(1)).Return(user, nil)

	// Mocking Update
	mockRepo.On("Update", testifyMock.AnythingOfType("*entity.User")).Return(nil).Run(func(args testifyMock.Arguments) {
		updatedUser := args.Get(0).(*entity.User)
		// Verify password has changed and is hashed
		err := bcrypt.CompareHashAndPassword([]byte(updatedUser.Password), []byte("newpassword"))
		assert.NoError(t, err)
	})

	input := service.ChangePasswordInput{
		NewPassword: "newpassword",
	}

	err := userService.ChangePassword(1, input)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
