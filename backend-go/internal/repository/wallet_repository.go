package repository

import (
	"cuan-backend/internal/entity"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type WalletRepository interface {
	Create(wallet *entity.Wallet) error
	Update(wallet *entity.Wallet) error
	Delete(id uint, userID uint) error
	FindByID(id uint, userID uint) (*entity.Wallet, error)
	FindByUserID(userID uint) ([]entity.Wallet, error)
	WithTx(tx *gorm.DB) WalletRepository
}

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) WalletRepository {
	return &walletRepository{db}
}

func (r *walletRepository) WithTx(tx *gorm.DB) WalletRepository {
	return &walletRepository{db: tx}
}

func (r *walletRepository) Create(wallet *entity.Wallet) error {
	if err := r.db.Create(wallet).Error; err != nil {
		log.Error().Err(err).Uint("user_id", wallet.UserID).Msg("Database operation failed")
		return err
	}
	return nil
}

func (r *walletRepository) Update(wallet *entity.Wallet) error {
	if err := r.db.Save(wallet).Error; err != nil {
		log.Error().Err(err).Uint("wallet_id", wallet.ID).Uint("user_id", wallet.UserID).Msg("Database operation failed")
		return err
	}
	return nil
}

func (r *walletRepository) Delete(id uint, userID uint) error {
	if err := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&entity.Wallet{}).Error; err != nil {
		log.Error().Err(err).Uint("wallet_id", id).Uint("user_id", userID).Msg("Database operation failed")
		return err
	}
	return nil
}

func (r *walletRepository) FindByID(id uint, userID uint) (*entity.Wallet, error) {
	var wallet entity.Wallet
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&wallet).Error
	if err != nil {
		log.Error().Err(err).Uint("wallet_id", id).Uint("user_id", userID).Msg("Database operation failed")
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) FindByUserID(userID uint) ([]entity.Wallet, error) {
	var wallets []entity.Wallet
	err := r.db.Where("user_id = ?", userID).Find(&wallets).Error
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Msg("Database operation failed")
	}
	return wallets, err
}
