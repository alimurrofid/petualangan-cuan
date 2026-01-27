package service

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"
	"cuan-backend/pkg/middleware"
	"errors"

	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
	Register(input RegisterInput) (*entity.User, string, string, error)
	Login(input LoginInput) (*entity.User, string, string, error)
	Logout(token string) error
	UpdateProfile(id uint, input UpdateProfileInput) (*entity.User, error)
	ChangePassword(id uint, input ChangePasswordInput) error
	LoginOrRegisterGoogle(email string, name string, googleID string) (*entity.User, string, string, error)
	GetProfile(id uint) (*entity.User, error)
	RefreshToken(refreshToken string) (string, string, error)
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

// Helper to get durations
func getTokenDurations() (time.Duration, time.Duration) {
	accessStr := os.Getenv("JWT_ACCESS_EXPIRY")
	if accessStr == "" {
		accessStr = "15m"
	}
	refreshStr := os.Getenv("JWT_REFRESH_EXPIRY")
	if refreshStr == "" {
		refreshStr = "72h"
	}

	accessDur, _ := time.ParseDuration(accessStr)
	refreshDur, _ := time.ParseDuration(refreshStr)
	
	return accessDur, refreshDur
}

func (s *userService) Register(input RegisterInput) (*entity.User, string, string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", err
	}

	user := &entity.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPassword),
	}

	err = s.userRepository.Create(user)
	if err != nil {
		return nil, "", "", err
	}

	accessDur, refreshDur := getTokenDurations()
	accessToken, err := middleware.GenerateToken(user.ID, accessDur)
	if err != nil {
		return nil, "", "", err
	}
	refreshToken, err := middleware.GenerateToken(user.ID, refreshDur)
	if err != nil {
		return nil, "", "", err
	}

	user.RefreshToken = refreshToken
	if err := s.userRepository.Update(user); err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *userService) Login(input LoginInput) (*entity.User, string, string, error) {
	user, err := s.userRepository.FindByEmail(input.Email)
	if err != nil {
		return nil, "", "", errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return nil, "", "", errors.New("invalid email or password")
	}

	accessDur, refreshDur := getTokenDurations()
	accessToken, err := middleware.GenerateToken(user.ID, accessDur)
	if err != nil {
		return nil, "", "", err
	}
	refreshToken, err := middleware.GenerateToken(user.ID, refreshDur)
	if err != nil {
		return nil, "", "", err
	}

	user.RefreshToken = refreshToken
	if err := s.userRepository.Update(user); err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *userService) Logout(token string) error {
	// Parse token to get userID
	// Since we don't have straightforward Parse in middleware exposed universally yet (except ad-hoc in RefreshToken),
	// we'll reuse the logic or if it fails we just ignore (best effort).
	// Actually, strict logout means cleaning DB.
	
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !parsedToken.Valid {
		// If token invalid, maybe already expired. Safe to ignore or return error?
		// For UX, return nil.
		return nil
	}
	
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil
	}

	var userID uint
	if idFloat, ok := claims["user_id"].(float64); ok {
		userID = uint(idFloat)
	} else {
		return nil
	}

	user, err := s.userRepository.FindByID(userID)
	if err != nil {
		return nil
	}

	user.RefreshToken = ""
	return s.userRepository.Update(user)
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

func (s *userService) LoginOrRegisterGoogle(email string, name string, googleID string) (*entity.User, string, string, error) {
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
			return nil, "", "", err
		}
	} else {
		// User found, update GoogleID if not set
		if user.GoogleID == "" {
			user.GoogleID = googleID
			err = s.userRepository.Update(user)
			if err != nil {
				return nil, "", "", err
			}
		}
	}

	accessDur, refreshDur := getTokenDurations()
	accessToken, err := middleware.GenerateToken(user.ID, accessDur)
	if err != nil {
		return nil, "", "", err
	}
	refreshToken, err := middleware.GenerateToken(user.ID, refreshDur)
	if err != nil {
		return nil, "", "", err
	}

	user.RefreshToken = refreshToken
	if err := s.userRepository.Update(user); err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *userService) RefreshToken(refreshToken string) (string, string, error) {
	// Parse and validate the token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return "", "", errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", errors.New("invalid token claims")
	}

	// Extract User ID
	var userID uint
	if idFloat, ok := claims["user_id"].(float64); ok {
		userID = uint(idFloat)
	} else {
		return "", "", errors.New("invalid user id in token")
	}

	// Verify user exists (optional but recommended)
	user, err := s.userRepository.FindByID(userID)
	if err != nil {
		return "", "", errors.New("user not found")
	}

	// Stateful Check: input token must match DB token
	if user.RefreshToken != refreshToken {
		return "", "", errors.New("invalid refresh token (reuse detected or logged out)")
	}

	// Generate new pair
	accessDur, refreshDur := getTokenDurations()
	newAccessToken, err := middleware.GenerateToken(userID, accessDur)
	if err != nil {
		return "", "", err
	}
	newRefreshToken, err := middleware.GenerateToken(userID, refreshDur)
	if err != nil {
		return "", "", err
	}

	// Update DB with new refresh token (Rotation)
	user.RefreshToken = newRefreshToken
	if err := s.userRepository.Update(user); err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

