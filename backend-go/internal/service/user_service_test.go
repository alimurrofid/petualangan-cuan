package service_test

import (
	"errors"
	"testing"

	"cuan-backend/internal/entity"
	"cuan-backend/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository is a mock of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *entity.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(email string) (*entity.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(id uint) (*entity.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func TestRegister(t *testing.T) {
	mockRepo := new(MockUserRepository)
	userService := service.NewUserService(mockRepo)

	input := service.RegisterInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	mockRepo.On("Create", mock.AnythingOfType("*entity.User")).Return(nil)

	token, err := userService.Register(input)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	mockRepo.AssertExpectations(t)
}

func TestLogin(t *testing.T) {
	mockRepo := new(MockUserRepository)
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

	token, err := userService.Login(input)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	mockRepo.AssertExpectations(t)
}

func TestLogin_InvalidPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
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

	token, err := userService.Login(input)

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, "invalid email or password", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestLogin_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	userService := service.NewUserService(mockRepo)

	mockRepo.On("FindByEmail", "notfound@example.com").Return(nil, errors.New("record not found"))

	input := service.LoginInput{
		Email:    "notfound@example.com",
		Password: "password123",
	}

	token, err := userService.Login(input)

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, "invalid email or password", err.Error())
	mockRepo.AssertExpectations(t)
}
