package entity

import (
	"time"
)

type DebtType string

const (
	DebtTypePayable    DebtType = "debt"       // Utang (kita berutang)
	DebtTypeReceivable DebtType = "receivable" // Piutang (orang lain berutang ke kita)
)

type Debt struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	UserID      uint      `gorm:"not null" json:"user_id"`
	User        User      `gorm:"foreignKey:UserID" json:"-"`
	WalletID    uint      `gorm:"not null" json:"wallet_id"`
	Wallet      Wallet    `gorm:"foreignKey:WalletID" json:"wallet"`
	Name        string    `gorm:"not null" json:"name"`
	Amount      float64   `gorm:"not null" json:"amount"`
	Remaining   float64   `gorm:"not null" json:"remaining"`
	Type        DebtType  `gorm:"type:varchar(20);not null" json:"type"`
	Description string    `json:"description"`
	DueDate     *time.Time    `json:"due_date"`
	Payments    []DebtPayment `json:"payments" gorm:"foreignKey:DebtID"`
	IsPaid      bool          `gorm:"default:false" json:"is_paid"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
