package repository

import (
	"cuan-backend/internal/entity"

	"gorm.io/gorm"
)

type ChatRepository interface {
	Save(msg *entity.ChatMessage) error
	FindByUserID(userID uint, limit int) ([]entity.ChatMessage, error)
	DeleteByUserID(userID uint) error
}

type chatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) Save(msg *entity.ChatMessage) error {
	return r.db.Create(msg).Error
}

func (r *chatRepository) FindByUserID(userID uint, limit int) ([]entity.ChatMessage, error) {
	var messages []entity.ChatMessage
	err := r.db.
		Where("user_id = ?", userID).
		Order("created_at ASC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

func (r *chatRepository) DeleteByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&entity.ChatMessage{}).Error
}
