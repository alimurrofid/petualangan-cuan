package service

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"
)

const DefaultChatHistoryLimit = 100

type ChatHistoryService interface {
	SaveMessage(userID uint, role, content, audioURL, imageURL string, transactions []entity.SavedTransaction) error
	SaveMessageWithReply(userID uint, role, content, audioURL, imageURL string, transactions []entity.SavedTransaction, replyToID *uint, replyToRole, replyToContent string) (*entity.ChatMessage, error)
	GetHistory(userID uint, limit int) ([]entity.ChatMessage, error)
	GetMessageByID(id uint, userID uint) (*entity.ChatMessage, error)
	FindNextAssistantMessage(userID uint, userMsgID uint) (*entity.ChatMessage, error)
	UpdateMessageContent(id uint, userID uint, content string) (*entity.ChatMessage, error)
	DeleteMessage(id uint, userID uint) error
	ClearHistory(userID uint) error
}

type chatHistoryService struct {
	repo repository.ChatRepository
}

func NewChatHistoryService(repo repository.ChatRepository) ChatHistoryService {
	return &chatHistoryService{repo: repo}
}

func (s *chatHistoryService) SaveMessage(userID uint, role, content, audioURL, imageURL string, transactions []entity.SavedTransaction) error {
	_, err := s.SaveMessageWithReply(userID, role, content, audioURL, imageURL, transactions, nil, "", "")
	return err
}

func (s *chatHistoryService) SaveMessageWithReply(userID uint, role, content, audioURL, imageURL string, transactions []entity.SavedTransaction, replyToID *uint, replyToRole, replyToContent string) (*entity.ChatMessage, error) {
	msg := &entity.ChatMessage{
		UserID:         userID,
		Role:           role,
		Content:        content,
		AudioURL:       audioURL,
		ImageURL:       imageURL,
		Transactions:   transactions,
		ReplyToID:      replyToID,
		ReplyToRole:    replyToRole,
		ReplyToContent: replyToContent,
	}
	err := s.repo.Save(msg)
	return msg, err
}

func (s *chatHistoryService) GetHistory(userID uint, limit int) ([]entity.ChatMessage, error) {
	if limit <= 0 {
		limit = DefaultChatHistoryLimit
	}
	return s.repo.FindByUserID(userID, limit)
}

func (s *chatHistoryService) GetMessageByID(id uint, userID uint) (*entity.ChatMessage, error) {
	return s.repo.FindByID(id, userID)
}

func (s *chatHistoryService) FindNextAssistantMessage(userID uint, userMsgID uint) (*entity.ChatMessage, error) {
	return s.repo.FindNextAssistantMessage(userID, userMsgID)
}

func (s *chatHistoryService) UpdateMessageContent(id uint, userID uint, content string) (*entity.ChatMessage, error) {
	msg, err := s.repo.FindByID(id, userID)
	if err != nil {
		return nil, err
	}
	msg.Content = content
	msg.IsEdited = true
	if err := s.repo.Update(msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (s *chatHistoryService) DeleteMessage(id uint, userID uint) error {
	return s.repo.DeleteByID(id, userID)
}

func (s *chatHistoryService) ClearHistory(userID uint) error {
	return s.repo.DeleteByUserID(userID)
}
