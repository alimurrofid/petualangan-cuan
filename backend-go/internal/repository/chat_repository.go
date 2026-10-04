package repository

import (
	"cuan-backend/internal/entity"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type ChatRepository interface {
	Save(msg *entity.ChatMessage) error
	FindByUserID(userID uint, limit int) ([]entity.ChatMessage, error)
	FindByID(id uint, userID uint) (*entity.ChatMessage, error)
	FindNextAssistantMessage(userID uint, userMsgID uint) (*entity.ChatMessage, error)
	Update(msg *entity.ChatMessage) error
	DeleteByID(id uint, userID uint) error
	DeleteByUserID(userID uint) error
}

type chatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) Save(msg *entity.ChatMessage) error {
	if err := r.db.Create(msg).Error; err != nil {
		log.Error().Err(err).Uint("user_id", msg.UserID).Msg("Database operation failed")
		return err
	}
	return nil
}

func (r *chatRepository) FindByUserID(userID uint, limit int) ([]entity.ChatMessage, error) {
	var messages []entity.ChatMessage
	err := r.db.
		Where("user_id = ?", userID).
		Order("created_at ASC").
		Limit(limit).
		Find(&messages).Error
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Msg("Database operation failed")
	}
	return messages, err
}

func (r *chatRepository) FindByID(id uint, userID uint) (*entity.ChatMessage, error) {
	var msg entity.ChatMessage
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&msg).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *chatRepository) FindNextAssistantMessage(userID uint, userMsgID uint) (*entity.ChatMessage, error) {
	var msg entity.ChatMessage
	err := r.db.Where("user_id = ? AND role = ? AND id > ?", userID, "assistant", userMsgID).
		Order("id ASC").
		First(&msg).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *chatRepository) Update(msg *entity.ChatMessage) error {
	if err := r.db.Save(msg).Error; err != nil {
		log.Error().Err(err).Uint("user_id", msg.UserID).Uint("id", msg.ID).Msg("Database update failed")
		return err
	}
	return nil
}

func (r *chatRepository) DeleteByID(id uint, userID uint) error {
	if err := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&entity.ChatMessage{}).Error; err != nil {
		log.Error().Err(err).Uint("user_id", userID).Uint("id", id).Msg("Database delete failed")
		return err
	}
	return nil
}

func (r *chatRepository) DeleteByUserID(userID uint) error {
	if err := r.db.Where("user_id = ?", userID).Delete(&entity.ChatMessage{}).Error; err != nil {
		log.Error().Err(err).Uint("user_id", userID).Msg("Database operation failed")
		return err
	}
	return nil
}
