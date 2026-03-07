package entity

import (
	"time"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"type:varchar(100);not null" json:"name"`
	Email        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	GoogleID     *string   `gorm:"type:varchar(100);uniqueIndex" json:"google_id,omitempty"`
	Phone        *string   `gorm:"type:varchar(20);uniqueIndex" json:"phone,omitempty"`
	RefreshToken string    `gorm:"type:text" json:"-"`
	Password     string    `gorm:"type:varchar(255)" json:"-"`
	Payday       *int      `gorm:"default:1" json:"payday"` // Tanggal gajian, default hari ke-1
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
