package entity

import (
	"time"
)

type Wallet struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null" json:"user_id"`
	User       User      `gorm:"foreignKey:UserID" json:"-"`
	Name       string    `gorm:"not null" json:"name"`
	Type       string    `gorm:"not null" json:"type"` // Bank, E-Wallet, Cash
	Balance          float64   `gorm:"not null;default:0" json:"balance"`
	AvailableBalance float64   `gorm:"-" json:"available_balance"`
	Icon             string    `json:"icon"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
