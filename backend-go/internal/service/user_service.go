package service

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"
	"cuan-backend/pkg/middleware"
	"errors"

	"os"
	"time"

	"github.com/rs/zerolog/log"

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

func (s *userService) GetProfile(id uint) (*entity.User, error) {
	user, err := s.userRepository.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

type UpdateProfileInput struct {
	Name   string  `json:"name"`
	Email  string  `json:"email"`
	Phone  *string `json:"phone"`
	Payday *int    `json:"payday"` // Tanggal gajian (1-28), nil = tidak diubah
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
	log.Info().Str("email", input.Email).Msg("Starting user registration")
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

	log.Info().Uint("user_id", user.ID).Str("email", user.Email).Msg("Completed user registration")
	return user, accessToken, refreshToken, nil
}

func (s *userService) Login(input LoginInput) (*entity.User, string, string, error) {
	log.Info().Str("email", input.Email).Msg("Starting user login")
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
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !parsedToken.Valid {
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
	user.Email = input.Email
	// Phone pointer: nil = tidak diubah, non-nil = update (termasuk string kosong untuk clear)
	if input.Phone != nil {
		if *input.Phone == "" {
			user.Phone = nil
		} else {
			user.Phone = input.Phone
		}
	}
	// Payday: nil = tidak diubah; nilai di luar 1-31 = reset ke 1
	if input.Payday != nil {
		p := *input.Payday
		if p < 1 || p > 31 {
			p = 1
		}
		user.Payday = &p
	}

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
	user, err := s.userRepository.FindByEmail(email)
	if err != nil {
		user = &entity.User{
			Name:     name,
			Email:    email,
			GoogleID: &googleID,
		}
		err = s.userRepository.Create(user)
		if err != nil {
			return nil, "", "", err
		}
	} else {
		if user.GoogleID == nil || *user.GoogleID == "" {
			user.GoogleID = &googleID
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

	var userID uint
	if idFloat, ok := claims["user_id"].(float64); ok {
		userID = uint(idFloat)
	} else {
		return "", "", errors.New("invalid user id in token")
	}

	user, err := s.userRepository.FindByID(userID)
	if err != nil {
		return "", "", errors.New("user not found")
	}
	if user.RefreshToken != refreshToken {
		return "", "", errors.New("invalid refresh token (reuse detected or logged out)")
	}

	accessDur, refreshDur := getTokenDurations()
	newAccessToken, err := middleware.GenerateToken(userID, accessDur)
	if err != nil {
		return "", "", err
	}
	newRefreshToken, err := middleware.GenerateToken(userID, refreshDur)
	if err != nil {
		return "", "", err
	}

	user.RefreshToken = newRefreshToken
	if err := s.userRepository.Update(user); err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

