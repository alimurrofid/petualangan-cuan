package entity

import (
	"time"
)

type DebtPayment struct {
	ID            uint        `gorm:"primarykey" json:"id"`
	DebtID        uint        `gorm:"not null" json:"debt_id"`
	Debt          Debt        `gorm:"foreignKey:DebtID" json:"-"`
	TransactionID uint        `gorm:"not null" json:"transaction_id"`
	Transaction   Transaction `gorm:"foreignKey:TransactionID" json:"transaction"`
	WalletID      uint        `gorm:"not null" json:"wallet_id"`
	Wallet        Wallet      `gorm:"foreignKey:WalletID" json:"wallet"` // Wallet used for THIS payment
	Amount        float64     `gorm:"not null" json:"amount"`
	Date          time.Time   `json:"date"`
	Note          string      `json:"note"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}
