package entity

import "time"

type ChatMessage struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"not null;index"           json:"user_id"`
	Role      string    `gorm:"not null;type:varchar(20)" json:"role"` // "user" | "assistant"
	Content   string    `gorm:"type:text"                json:"content"`
	AudioURL  string    `gorm:"type:varchar(500)"        json:"audio_url,omitempty"`
	ImageURL  string    `gorm:"type:varchar(500)"        json:"image_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
