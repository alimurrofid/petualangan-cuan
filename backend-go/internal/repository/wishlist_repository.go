package repository

import (
	"cuan-backend/internal/entity"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type WishlistRepository interface {
	Create(item *entity.WishlistItem) error
	FindAllByUserID(userID uint) ([]entity.WishlistItem, error)
	FindByID(id uint, userID uint) (*entity.WishlistItem, error)
	Update(item *entity.WishlistItem) error
	Delete(id uint, userID uint) error
	MarkAsBought(id uint, userID uint) error
}

type wishlistRepository struct {
	db *gorm.DB
}

func NewWishlistRepository(db *gorm.DB) WishlistRepository {
	return &wishlistRepository{db}
}

func (r *wishlistRepository) Create(item *entity.WishlistItem) error {
	if err := r.db.Create(item).Error; err != nil {
		log.Error().Err(err).Uint("user_id", item.UserID).Msg("Database operation failed")
		return err
	}
	return nil
}

func (r *wishlistRepository) FindAllByUserID(userID uint) ([]entity.WishlistItem, error) {
	var items []entity.WishlistItem
	err := r.db.Preload("Category").Where("user_id = ?", userID).Find(&items).Error
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Msg("Database operation failed")
	}
	return items, err
}

func (r *wishlistRepository) FindByID(id uint, userID uint) (*entity.WishlistItem, error) {
	var item entity.WishlistItem
	err := r.db.Preload("Category").Where("id = ? AND user_id = ?", id, userID).First(&item).Error
	if err != nil {
		log.Error().Err(err).Uint("wishlist_id", id).Uint("user_id", userID).Msg("Database operation failed")
		return nil, err
	}
	return &item, nil
}

func (r *wishlistRepository) Update(item *entity.WishlistItem) error {
	if err := r.db.Save(item).Error; err != nil {
		log.Error().Err(err).Uint("wishlist_id", item.ID).Uint("user_id", item.UserID).Msg("Database operation failed")
		return err
	}
	return nil
}

func (r *wishlistRepository) Delete(id uint, userID uint) error {
	if err := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&entity.WishlistItem{}).Error; err != nil {
		log.Error().Err(err).Uint("wishlist_id", id).Uint("user_id", userID).Msg("Database operation failed")
		return err
	}
	return nil
}

func (r *wishlistRepository) MarkAsBought(id uint, userID uint) error {
	if err := r.db.Model(&entity.WishlistItem{}).Where("id = ? AND user_id = ?", id, userID).Update("is_bought", true).Error; err != nil {
		log.Error().Err(err).Uint("wishlist_id", id).Uint("user_id", userID).Msg("Database operation failed")
		return err
	}
	return nil
}
