package service

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"

	"github.com/rs/zerolog/log"
)

type CategoryService interface {
	CreateCategory(userID uint, input CreateCategoryInput) (*entity.Category, error)
	GetCategories(userID uint) ([]entity.Category, error)
	GetCategory(id uint, userID uint) (*entity.Category, error)
	UpdateCategory(id uint, userID uint, input UpdateCategoryInput) (*entity.Category, error)
	DeleteCategory(id uint, userID uint) error
}

type categoryService struct {
	repo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryService{repo}
}

type CreateCategoryInput struct {
	Name        string  `json:"name" binding:"required"`
	Type        string  `json:"type" binding:"required"`
	Icon        string  `json:"icon"`
	BudgetLimit float64 `json:"budget_limit"`
}

type UpdateCategoryInput struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Icon        string  `json:"icon"`
	BudgetLimit float64 `json:"budget_limit"`
}

func (s *categoryService) CreateCategory(userID uint, input CreateCategoryInput) (*entity.Category, error) {
	category := &entity.Category{
		UserID:      userID,
		Name:        input.Name,
		Type:        input.Type,
		Icon:        input.Icon,
		BudgetLimit: input.BudgetLimit,
	}
	err := s.repo.Create(category)
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Str("category_name", category.Name).Msg("Failed to create category")
		return nil, err
	}
	log.Info().Uint("user_id", userID).Str("category_name", category.Name).Msg("Category created successfully")
	return category, nil
}

func (s *categoryService) GetCategories(userID uint) ([]entity.Category, error) {
	return s.repo.FindAll(userID)
}

func (s *categoryService) GetCategory(id uint, userID uint) (*entity.Category, error) {
	return s.repo.FindByID(id, userID)
}

func (s *categoryService) UpdateCategory(id uint, userID uint, input UpdateCategoryInput) (*entity.Category, error) {
	category, err := s.repo.FindByID(id, userID)
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Uint("category_id", id).Msg("Failed to find category for update")
		return nil, err
	}

	if input.Name != "" {
		category.Name = input.Name
	}
	if input.Type != "" {
		category.Type = input.Type
	}

	category.Icon = input.Icon
	category.BudgetLimit = input.BudgetLimit

	err = s.repo.Update(category)
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Uint("category_id", id).Msg("Failed to update category")
		return nil, err
	}
	log.Info().Uint("user_id", userID).Uint("category_id", id).Msg("Category updated successfully")
	return category, nil
}

func (s *categoryService) DeleteCategory(id uint, userID uint) error {
	err := s.repo.Delete(id, userID)
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Uint("category_id", id).Msg("Failed to delete category")
		return err
	}
	log.Info().Uint("user_id", userID).Uint("category_id", id).Msg("Category deleted successfully")
	return nil
}
