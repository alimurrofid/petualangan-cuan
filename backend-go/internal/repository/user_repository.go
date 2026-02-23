package repository

import (
	"cuan-backend/internal/entity"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *entity.User) error
	FindByEmail(email string) (*entity.User, error)
	FindByID(id uint) (*entity.User, error)
	FindByPhone(phone string) (*entity.User, error)
	Update(user *entity.User) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db}
}

func (r *userRepository) Update(user *entity.User) error {
	if err := r.db.Save(user).Error; err != nil {
		log.Error().Err(err).Uint("user_id", user.ID).Msg("Database operation failed")
		return err
	}
	return nil
}

func (r *userRepository) Create(user *entity.User) error {
	if err := r.db.Create(user).Error; err != nil {
		log.Error().Err(err).Str("email", user.Email).Msg("Database operation failed")
		return err
	}
	return nil
}

func (r *userRepository) FindByEmail(email string) (*entity.User, error) {
	var user entity.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		log.Error().Err(err).Str("email", email).Msg("Database operation failed")
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(id uint) (*entity.User, error) {
	var user entity.User
	err := r.db.First(&user, id).Error
	if err != nil {
		log.Error().Err(err).Uint("user_id", id).Msg("Database operation failed")
		return nil, err
	}
	return &user, nil
}
func (r *userRepository) FindByPhone(phone string) (*entity.User, error) {
	var user entity.User
	err := r.db.Where("phone = ?", phone).First(&user).Error
	if err != nil {
		log.Error().Err(err).Str("phone", phone).Msg("Database operation failed")
		return nil, err
	}
	return &user, nil
}
