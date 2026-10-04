package service

import (
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository/mock"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
)

func TestChatHistoryService_SaveMessage(t *testing.T) {
	repo := new(mock.ChatRepositoryMock)
	svc := NewChatHistoryService(repo)

	repo.On("Save", testifymock.MatchedBy(func(msg *entity.ChatMessage) bool {
		return msg.UserID == 1 && msg.Role == "user" && msg.Content == "Halo"
	})).Return(nil)

	err := svc.SaveMessage(1, "user", "Halo", "", "", nil)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestChatHistoryService_SaveMessageWithReply(t *testing.T) {
	repo := new(mock.ChatRepositoryMock)
	svc := NewChatHistoryService(repo)

	replyID := uint(10)
	repo.On("Save", testifymock.MatchedBy(func(msg *entity.ChatMessage) bool {
		return msg.UserID == 1 && msg.ReplyToID != nil && *msg.ReplyToID == replyID
	})).Return(nil)

	msg, err := svc.SaveMessageWithReply(1, "assistant", "Balasan", "", "", nil, &replyID, "user", "Kutipan")
	assert.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, &replyID, msg.ReplyToID)
	assert.Equal(t, "user", msg.ReplyToRole)
	repo.AssertExpectations(t)
}

func TestChatHistoryService_GetHistory(t *testing.T) {
	repo := new(mock.ChatRepositoryMock)
	svc := NewChatHistoryService(repo)

	expected := []entity.ChatMessage{
		{ID: 1, UserID: 1, Content: "Pesan 1"},
		{ID: 2, UserID: 1, Content: "Pesan 2"},
	}

	repo.On("FindByUserID", uint(1), 50).Return(expected, nil)

	res, err := svc.GetHistory(1, 50)
	assert.NoError(t, err)
	assert.Len(t, res, 2)

	// Test default limit when limit <= 0
	repo.On("FindByUserID", uint(1), DefaultChatHistoryLimit).Return(expected, nil)
	resDefault, err := svc.GetHistory(1, 0)
	assert.NoError(t, err)
	assert.Len(t, resDefault, 2)
}

func TestChatHistoryService_GetMessageByID(t *testing.T) {
	repo := new(mock.ChatRepositoryMock)
	svc := NewChatHistoryService(repo)

	expected := &entity.ChatMessage{ID: 5, UserID: 1, Content: "Pesan 5"}
	repo.On("FindByID", uint(5), uint(1)).Return(expected, nil)

	res, err := svc.GetMessageByID(5, 1)
	assert.NoError(t, err)
	assert.Equal(t, uint(5), res.ID)
}

func TestChatHistoryService_FindNextAssistantMessage(t *testing.T) {
	repo := new(mock.ChatRepositoryMock)
	svc := NewChatHistoryService(repo)

	expected := &entity.ChatMessage{ID: 6, UserID: 1, Role: "assistant", Content: "Balasan AI"}
	repo.On("FindNextAssistantMessage", uint(1), uint(5)).Return(expected, nil)

	res, err := svc.FindNextAssistantMessage(1, 5)
	assert.NoError(t, err)
	assert.Equal(t, uint(6), res.ID)
}

func TestChatHistoryService_UpdateMessageContent(t *testing.T) {
	repo := new(mock.ChatRepositoryMock)
	svc := NewChatHistoryService(repo)

	original := &entity.ChatMessage{ID: 5, UserID: 1, Content: "Pesan lama", IsEdited: false}
	repo.On("FindByID", uint(5), uint(1)).Return(original, nil)
	repo.On("Update", testifymock.MatchedBy(func(m *entity.ChatMessage) bool {
		return m.ID == 5 && m.Content == "Pesan baru" && m.IsEdited == true
	})).Return(nil)

	res, err := svc.UpdateMessageContent(5, 1, "Pesan baru")
	assert.NoError(t, err)
	assert.Equal(t, "Pesan baru", res.Content)
	assert.True(t, res.IsEdited)

	// Test error when not found
	repo.On("FindByID", uint(99), uint(1)).Return(nil, errors.New("not found"))
	_, errNotFound := svc.UpdateMessageContent(99, 1, "Baru")
	assert.Error(t, errNotFound)
}

func TestChatHistoryService_DeleteMessage(t *testing.T) {
	repo := new(mock.ChatRepositoryMock)
	svc := NewChatHistoryService(repo)

	repo.On("DeleteByID", uint(5), uint(1)).Return(nil)

	err := svc.DeleteMessage(5, 1)
	assert.NoError(t, err)
}

func TestChatHistoryService_ClearHistory(t *testing.T) {
	repo := new(mock.ChatRepositoryMock)
	svc := NewChatHistoryService(repo)

	repo.On("DeleteByUserID", uint(1)).Return(nil)

	err := svc.ClearHistory(1)
	assert.NoError(t, err)
}
