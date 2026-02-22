package mock

import (
	"cuan-backend/internal/entity"

	"github.com/stretchr/testify/mock"
)

type ChatRepositoryMock struct {
	mock.Mock
}

func (m *ChatRepositoryMock) Save(msg *entity.ChatMessage) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *ChatRepositoryMock) FindByUserID(userID uint, limit int) ([]entity.ChatMessage, error) {
	args := m.Called(userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.ChatMessage), args.Error(1)
}

func (m *ChatRepositoryMock) DeleteByUserID(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}
