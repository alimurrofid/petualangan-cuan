package entity

import (
	"time"
)

type Transaction struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	UserID      uint           `gorm:"not null" json:"user_id"`
	User        User           `gorm:"foreignKey:UserID" json:"-"`
	RelatedTransactionID *uint `json:"related_transaction_id"`
	WalletID    uint           `gorm:"not null" json:"wallet_id"`
	Wallet      Wallet         `gorm:"foreignKey:WalletID" json:"wallet"`
	CategoryID  uint           `gorm:"not null" json:"category_id"`
	Category    Category       `gorm:"foreignKey:CategoryID" json:"category"`
	Amount      float64        `gorm:"not null" json:"amount"`
	Type        string         `gorm:"not null" json:"type"`
	Description string         `json:"description"`
	Date        time.Time      `gorm:"not null" json:"date"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type TransactionSummary struct {
	Date    string  `json:"date"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
}

type TransactionFilterParams struct {
	Page       int
	Limit      int
	StartDate  string
	EndDate    string
	WalletID   uint
	CategoryID uint
	Search     string
	Type       string
}
