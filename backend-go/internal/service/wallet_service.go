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
	savingGoalRepo   repository.SavingGoalRepository
}

func NewWalletService(walletRepository repository.WalletRepository, savingGoalRepo repository.SavingGoalRepository) WalletService {
	return &walletService{walletRepository, savingGoalRepo}
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
	
	wallet.AvailableBalance = wallet.Balance
	return wallet, nil
}

func (s *walletService) GetWalletByID(id uint, userID uint) (*entity.Wallet, error) {
	wallet, err := s.walletRepository.FindByID(id, userID)
	if err != nil {
		return nil, err
	}
	
	activeContrib, err := s.savingGoalRepo.GetActiveContributions(wallet.ID)
	if err == nil {
		wallet.AvailableBalance = wallet.Balance - activeContrib
	} else {
		wallet.AvailableBalance = wallet.Balance
	}
	
	return wallet, nil
}

func (s *walletService) GetUserWallets(userID uint) ([]entity.Wallet, error) {
	wallets, err := s.walletRepository.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	
	for i := range wallets {
		activeContrib, err := s.savingGoalRepo.GetActiveContributions(wallets[i].ID)
		if err == nil {
			wallets[i].AvailableBalance = wallets[i].Balance - activeContrib
		} else {
			wallets[i].AvailableBalance = wallets[i].Balance
		}
	}
	
	return wallets, nil
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
	
	activeContrib, err := s.savingGoalRepo.GetActiveContributions(wallet.ID)
	if err == nil {
		wallet.AvailableBalance = wallet.Balance - activeContrib
	} else {
		wallet.AvailableBalance = wallet.Balance
	}

	return wallet, nil
}

func (s *walletService) DeleteWallet(id uint, userID uint) error {
	return s.walletRepository.Delete(id, userID)
}
