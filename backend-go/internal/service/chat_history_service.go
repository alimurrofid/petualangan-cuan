package service

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"
)

const DefaultChatHistoryLimit = 100

type ChatHistoryService interface {
	SaveMessage(userID uint, role, content, audioURL, imageURL string) error
	GetHistory(userID uint, limit int) ([]entity.ChatMessage, error)
	ClearHistory(userID uint) error
}

type chatHistoryService struct {
	repo repository.ChatRepository
}

func NewChatHistoryService(repo repository.ChatRepository) ChatHistoryService {
	return &chatHistoryService{repo: repo}
}

func (s *chatHistoryService) SaveMessage(userID uint, role, content, audioURL, imageURL string) error {
	msg := &entity.ChatMessage{
		UserID:   userID,
		Role:     role,
		Content:  content,
		AudioURL: audioURL,
		ImageURL: imageURL,
	}
	return s.repo.Save(msg)
}

func (s *chatHistoryService) GetHistory(userID uint, limit int) ([]entity.ChatMessage, error) {
	if limit <= 0 {
		limit = DefaultChatHistoryLimit
	}
	return s.repo.FindByUserID(userID, limit)
}

func (s *chatHistoryService) ClearHistory(userID uint) error {
	return s.repo.DeleteByUserID(userID)
}
