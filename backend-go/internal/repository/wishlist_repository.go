package repository

import (
	"cuan-backend/internal/entity"

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
	return r.db.Create(item).Error
}

func (r *wishlistRepository) FindAllByUserID(userID uint) ([]entity.WishlistItem, error) {
	var items []entity.WishlistItem
	err := r.db.Preload("Category").Where("user_id = ?", userID).Find(&items).Error
	return items, err
}

func (r *wishlistRepository) FindByID(id uint, userID uint) (*entity.WishlistItem, error) {
	var item entity.WishlistItem
	err := r.db.Preload("Category").Where("id = ? AND user_id = ?", id, userID).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *wishlistRepository) Update(item *entity.WishlistItem) error {
	return r.db.Save(item).Error
}

func (r *wishlistRepository) Delete(id uint, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&entity.WishlistItem{}).Error
}

func (r *wishlistRepository) MarkAsBought(id uint, userID uint) error {
	return r.db.Model(&entity.WishlistItem{}).Where("id = ? AND user_id = ?", id, userID).Update("is_bought", true).Error
}
