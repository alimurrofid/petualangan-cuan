package entity

import "time"

type SavingGoal struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"not null" json:"user_id"`
	User          User      `gorm:"foreignKey:UserID" json:"-"`
	Name          string    `gorm:"not null" json:"name"`
	TargetAmount  float64   `gorm:"not null" json:"target_amount"`
	CurrentAmount float64   `gorm:"not null;default:0" json:"current_amount"`
	CategoryID    uint      `gorm:"default:null" json:"category_id"`
	Category      Category  `gorm:"foreignKey:CategoryID" json:"category"`
	Deadline      *time.Time `json:"deadline"` // Optional deadline
	IsAchieved    bool      `gorm:"default:false" json:"is_achieved"`
	Icon          string    `json:"icon"` // PiggyBank or Target icon
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type SavingContribution struct {
	ID            uint        `gorm:"primaryKey" json:"id"`
	GoalID        uint        `gorm:"not null" json:"goal_id"`
	SavingGoal    SavingGoal  `gorm:"foreignKey:GoalID" json:"saving_goal"`
	WalletID      uint        `gorm:"not null" json:"wallet_id"`
	Wallet        Wallet      `gorm:"foreignKey:WalletID" json:"wallet"`
	TransactionID uint        `gorm:"not null;unique" json:"transaction_id"` // One-to-one with payment tx
	Transaction   Transaction `gorm:"foreignKey:TransactionID" json:"transaction"`
	Amount        float64     `gorm:"not null" json:"amount"`
	Date          time.Time   `gorm:"not null" json:"date"`
	CreatedAt     time.Time   `json:"created_at"`
}
