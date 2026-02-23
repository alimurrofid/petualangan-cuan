package service

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"

	"github.com/rs/zerolog/log"
)

type WishlistService interface {
	Create(userID uint, req *StoreWishlistRequest) error
	FindAll(userID uint) ([]entity.WishlistItem, error)
	FindByID(id uint, userID uint) (*entity.WishlistItem, error)
	Update(id uint, userID uint, req *StoreWishlistRequest) error
	Delete(id uint, userID uint) error
	MarkAsBought(id uint, userID uint) error
}

type wishlistService struct {
	wishlistRepo repository.WishlistRepository
}

func NewWishlistService(wishlistRepo repository.WishlistRepository) WishlistService {
	return &wishlistService{wishlistRepo}
}

type StoreWishlistRequest struct {
	CategoryID     uint    `json:"category_id" validate:"required"`
	Name           string  `json:"name" validate:"required"`
	EstimatedPrice float64 `json:"estimated_price" validate:"required"`
	Priority       string  `json:"priority" validate:"oneof=low medium high"`
}

func (s *wishlistService) Create(userID uint, req *StoreWishlistRequest) error {
	item := &entity.WishlistItem{
		UserID:         userID,
		CategoryID:     req.CategoryID,
		Name:           req.Name,
		EstimatedPrice: req.EstimatedPrice,
		Priority:       entity.WishlistPriority(req.Priority),
	}
	if req.Priority == "" {
		item.Priority = entity.WishlistPriorityMedium
	}
	err := s.wishlistRepo.Create(item)
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Msg("Failed to create wishlist item")
		return err
	}
	log.Info().Uint("user_id", userID).Uint("wishlist_id", item.ID).Msg("Wishlist item created successfully")
	return nil
}

func (s *wishlistService) FindAll(userID uint) ([]entity.WishlistItem, error) {
	return s.wishlistRepo.FindAllByUserID(userID)
}

func (s *wishlistService) FindByID(id uint, userID uint) (*entity.WishlistItem, error) {
	return s.wishlistRepo.FindByID(id, userID)
}

func (s *wishlistService) Update(id uint, userID uint, req *StoreWishlistRequest) error {
	item, err := s.wishlistRepo.FindByID(id, userID)
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Uint("wishlist_id", id).Msg("Failed to find wishlist item")
		return err
	}

	item.CategoryID = req.CategoryID
	item.Name = req.Name
	item.EstimatedPrice = req.EstimatedPrice
	if req.Priority != "" {
		item.Priority = entity.WishlistPriority(req.Priority)
	}

	err = s.wishlistRepo.Update(item)
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Uint("wishlist_id", id).Msg("Failed to update wishlist item")
		return err
	}
	log.Info().Uint("user_id", userID).Uint("wishlist_id", id).Msg("Wishlist item updated successfully")
	return nil
}

func (s *wishlistService) Delete(id uint, userID uint) error {
	err := s.wishlistRepo.Delete(id, userID)
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Uint("wishlist_id", id).Msg("Failed to delete wishlist item")
		return err
	}
	log.Info().Uint("user_id", userID).Uint("wishlist_id", id).Msg("Wishlist item deleted successfully")
	return nil
}

func (s *wishlistService) MarkAsBought(id uint, userID uint) error {
	err := s.wishlistRepo.MarkAsBought(id, userID)
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Uint("wishlist_id", id).Msg("Failed to mark wishlist item as bought")
		return err
	}
	log.Info().Uint("user_id", userID).Uint("wishlist_id", id).Msg("Wishlist item marked as bought")
	return nil
}
