package entity

import (
	"time"

	"gorm.io/gorm"
)

type Transaction struct {
	gorm.Model
	Item       string    `json:"item"`
	Category   string    `json:"category"`
	Amount     int64     `json:"amount"`
	Type       string    `json:"type"`
	RawChat    string    `json:"raw_chat"`
	Date       time.Time `json:"date"`
}
