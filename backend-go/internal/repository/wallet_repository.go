package repository

import (
	"cuan-backend/internal/entity"

	"gorm.io/gorm"
)

type WalletRepository interface {
	Create(wallet *entity.Wallet) error
	Update(wallet *entity.Wallet) error
	Delete(id uint, userID uint) error
	FindByID(id uint, userID uint) (*entity.Wallet, error)
	FindByUserID(userID uint) ([]entity.Wallet, error)
}

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) WalletRepository {
	return &walletRepository{db}
}

func (r *walletRepository) Create(wallet *entity.Wallet) error {
	return r.db.Create(wallet).Error
}

func (r *walletRepository) Update(wallet *entity.Wallet) error {
	return r.db.Save(wallet).Error
}

func (r *walletRepository) Delete(id uint, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&entity.Wallet{}).Error
}

func (r *walletRepository) FindByID(id uint, userID uint) (*entity.Wallet, error) {
	var wallet entity.Wallet
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) FindByUserID(userID uint) ([]entity.Wallet, error) {
	var wallets []entity.Wallet
	err := r.db.Where("user_id = ?", userID).Find(&wallets).Error
	return wallets, err
}
