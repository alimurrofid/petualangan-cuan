package service

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"
)

type CreateWalletInput struct {
	UserID  uint    `json:"user_id"`
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	Balance float64 `json:"balance"`
	Icon    string  `json:"icon"`
}

type UpdateWalletInput struct {
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	Balance float64 `json:"balance"`
	Icon    string  `json:"icon"`
}

type WalletService interface {
	CreateWallet(input CreateWalletInput) (*entity.Wallet, error)
	GetWalletByID(id uint, userID uint) (*entity.Wallet, error)
	GetUserWallets(userID uint) ([]entity.Wallet, error)
	UpdateWallet(id uint, userID uint, input UpdateWalletInput) (*entity.Wallet, error)
	DeleteWallet(id uint, userID uint) error
}

type walletService struct {
	walletRepository repository.WalletRepository
}

func NewWalletService(walletRepository repository.WalletRepository) WalletService {
	return &walletService{walletRepository}
}

func (s *walletService) CreateWallet(input CreateWalletInput) (*entity.Wallet, error) {
	wallet := &entity.Wallet{
		UserID:  input.UserID,
		Name:    input.Name,
		Type:    input.Type,
		Balance: input.Balance,
		Icon:    input.Icon,
	}

	err := s.walletRepository.Create(wallet)
	if err != nil {
		return nil, err
	}

	return wallet, nil
}

func (s *walletService) GetWalletByID(id uint, userID uint) (*entity.Wallet, error) {
	return s.walletRepository.FindByID(id, userID)
}

func (s *walletService) GetUserWallets(userID uint) ([]entity.Wallet, error) {
	return s.walletRepository.FindByUserID(userID)
}

func (s *walletService) UpdateWallet(id uint, userID uint, input UpdateWalletInput) (*entity.Wallet, error) {
	wallet, err := s.walletRepository.FindByID(id, userID)
	if err != nil {
		return nil, err
	}

	wallet.Name = input.Name
	wallet.Type = input.Type
	wallet.Balance = input.Balance
	wallet.Icon = input.Icon

	err = s.walletRepository.Update(wallet)
	if err != nil {
		return nil, err
	}

	return wallet, nil
}

func (s *walletService) DeleteWallet(id uint, userID uint) error {
	return s.walletRepository.Delete(id, userID)
}
