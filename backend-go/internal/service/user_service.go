package service

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"
	"cuan-backend/pkg/middleware"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserService interface {
	Register(input RegisterInput) (*entity.User, string, error)
	Login(input LoginInput) (*entity.User, string, error)
	Logout(token string) error
	UpdateProfile(id uint, input UpdateProfileInput) (*entity.User, error)
	ChangePassword(id uint, input ChangePasswordInput) error
	LoginOrRegisterGoogle(email string, name string, googleID string) (*entity.User, string, error)
	GetProfile(id uint) (*entity.User, error)
}

// ... (existing code) ...

func (s *userService) GetProfile(id uint) (*entity.User, error) {
	user, err := s.userRepository.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

type UpdateProfileInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type ChangePasswordInput struct {
	NewPassword string `json:"new_password"`
}

type userService struct {
	userRepository repository.UserRepository
}

func NewUserService(userRepository repository.UserRepository) UserService {
	return &userService{userRepository}
}

func (s *userService) Register(input RegisterInput) (*entity.User, string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	user := &entity.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPassword),
	}

	err = s.userRepository.Create(user)
	if err != nil {
		return nil, "", err
	}

	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *userService) Login(input LoginInput) (*entity.User, string, error) {
	user, err := s.userRepository.FindByEmail(input.Email)
	if err != nil {
		return nil, "", errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return nil, "", errors.New("invalid email or password")
	}

	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *userService) Logout(token string) error {
	// For stateless JWT, meaningful logout requires a blacklist/Redis.
	// For this task, we will just return nil to satisfy the interface 
	// and assume the client acts by discarding the token.
	return nil
}

func (s *userService) UpdateProfile(id uint, input UpdateProfileInput) (*entity.User, error) {
	user, err := s.userRepository.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	user.Name = input.Name
	user.Email = input.Email // Minimal validation for now

	err = s.userRepository.Update(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) ChangePassword(id uint, input ChangePasswordInput) error {
	user, err := s.userRepository.FindByID(id)
	if err != nil {
		return errors.New("user not found")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	return s.userRepository.Update(user)
}

func (s *userService) LoginOrRegisterGoogle(email string, name string, googleID string) (*entity.User, string, error) {
	// Check if user exists by email
	user, err := s.userRepository.FindByEmail(email)
	if err != nil {
		// User not found, create new user
		user = &entity.User{
			Name:     name,
			Email:    email,
			GoogleID: googleID,
		}
		err = s.userRepository.Create(user)
		if err != nil {
			return nil, "", err
		}
	} else {
		// User found, update GoogleID if not set
		if user.GoogleID == "" {
			user.GoogleID = googleID
			err = s.userRepository.Update(user)
			if err != nil {
				return nil, "", err
			}
		}
	}

	// Generate Token
	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}
