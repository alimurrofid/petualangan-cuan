package entity

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type ChatMessage struct {
	ID              uint               `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          uint               `gorm:"not null;index"           json:"user_id"`
	Role            string             `gorm:"not null;type:varchar(20)" json:"role"` // "user" | "assistant"
	Content         string             `gorm:"type:text"                json:"content"`
	AudioURL        string             `gorm:"type:varchar(500)"        json:"audio_url,omitempty"`
	ImageURL        string             `gorm:"type:varchar(500)"        json:"image_url,omitempty"`
	TransactionsRaw string             `gorm:"type:text"                json:"-"`
	Transactions    []SavedTransaction `gorm:"-"                        json:"transactions,omitempty"`
	CreatedAt       time.Time          `json:"created_at"`
}

func (m *ChatMessage) AfterFind(tx *gorm.DB) (err error) {
	if m.TransactionsRaw != "" {
		_ = json.Unmarshal([]byte(m.TransactionsRaw), &m.Transactions)
	}
	return
}

func (m *ChatMessage) BeforeSave(tx *gorm.DB) (err error) {
	if len(m.Transactions) > 0 {
		b, _ := json.Marshal(m.Transactions)
		m.TransactionsRaw = string(b)
	}
	return
}
