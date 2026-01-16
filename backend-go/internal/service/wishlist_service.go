package service

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"
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
	return s.wishlistRepo.Create(item)
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
		return err
	}

	item.CategoryID = req.CategoryID
	item.Name = req.Name
	item.EstimatedPrice = req.EstimatedPrice
	if req.Priority != "" {
		item.Priority = entity.WishlistPriority(req.Priority)
	}

	return s.wishlistRepo.Update(item)
}

func (s *wishlistService) Delete(id uint, userID uint) error {
	return s.wishlistRepo.Delete(id, userID)
}

func (s *wishlistService) MarkAsBought(id uint, userID uint) error {
	return s.wishlistRepo.MarkAsBought(id, userID)
}
