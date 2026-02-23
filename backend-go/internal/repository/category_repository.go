package repository

import (
	"cuan-backend/internal/entity"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(category *entity.Category) error
	FindAll(userID uint) ([]entity.Category, error)
	FindByID(id uint, userID uint) (*entity.Category, error)
	Update(category *entity.Category) error
	Delete(id uint, userID uint) error
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db}
}

func (r *categoryRepository) Create(category *entity.Category) error {
	if err := r.db.Create(category).Error; err != nil {
		log.Error().Err(err).Uint("user_id", category.UserID).Msg("Database operation failed")
		return err
	}
	return nil
}

func (r *categoryRepository) FindAll(userID uint) ([]entity.Category, error) {
	var categories []entity.Category
	err := r.db.Where("user_id = ?", userID).Find(&categories).Error
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Msg("Database operation failed")
	}
	return categories, err
}

func (r *categoryRepository) FindByID(id uint, userID uint) (*entity.Category, error) {
	var category entity.Category
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&category).Error
	if err != nil {
		log.Error().Err(err).Uint("category_id", id).Uint("user_id", userID).Msg("Database operation failed")
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) Update(category *entity.Category) error {
	if err := r.db.Save(category).Error; err != nil {
		log.Error().Err(err).Uint("category_id", category.ID).Uint("user_id", category.UserID).Msg("Database operation failed")
		return err
	}
	return nil
}

func (r *categoryRepository) Delete(id uint, userID uint) error {
	if err := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&entity.Category{}).Error; err != nil {
		log.Error().Err(err).Uint("category_id", id).Uint("user_id", userID).Msg("Database operation failed")
		return err
	}
	return nil
}
