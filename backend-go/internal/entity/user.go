package entity

import (
	"time"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	Email     string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	GoogleID  string         `gorm:"type:varchar(100);uniqueIndex" json:"google_id,omitempty"`
	Password  string         `gorm:"type:varchar(255)" json:"-"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`

}
