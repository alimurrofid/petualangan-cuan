package entity

import "time"

type WishlistPriority string

const (
	WishlistPriorityLow    WishlistPriority = "low"
	WishlistPriorityMedium WishlistPriority = "medium"
	WishlistPriorityHigh   WishlistPriority = "high"
)

type WishlistItem struct {
	ID             uint             `gorm:"primaryKey" json:"id"`
	UserID         uint             `gorm:"not null" json:"user_id"`
	User           User             `gorm:"foreignKey:UserID" json:"-"`
	CategoryID     uint             `gorm:"not null" json:"category_id"`
	Category       Category         `gorm:"foreignKey:CategoryID" json:"category"`
	Name           string           `gorm:"size:255;not null" json:"name"`
	EstimatedPrice float64          `gorm:"not null" json:"estimated_price"`
	IsBought       bool             `gorm:"default:false" json:"is_bought"`
	Priority       WishlistPriority `gorm:"type:varchar(20);default:'low'" json:"priority"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}
