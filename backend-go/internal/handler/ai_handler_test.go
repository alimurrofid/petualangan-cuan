package handler

import (
	"bytes"
	"context"
	"cuan-backend/internal/entity"
	aiprovider "cuan-backend/internal/provider/ai"
	"cuan-backend/internal/service"
	"encoding/json"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAIService struct {
	mock.Mock
}

func (m *mockAIService) Chat(params service.ChatParams) (*entity.ChatAIResponse, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ChatAIResponse), args.Error(1)
}

func (m *mockAIService) ChatStream(params service.ChatParams, onToken func(string) error) (*entity.ChatAIResponse, error) {
	args := m.Called(params, onToken)
	if onToken != nil {
		_ = onToken("Halo ")
		_ = onToken("User")
	}
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ChatAIResponse), args.Error(1)
}

func (m *mockAIService) ProcessVoice(path string) (string, error) {
	args := m.Called(path)
	return args.String(0), args.Error(1)
}

type mockAIProvider struct{}

func (m *mockAIProvider) GenerateCompletion(_ context.Context, _ aiprovider.AIRequest) (string, error) {
	return "", nil
}

type mockChatHistoryService struct{ mock.Mock }

func (m *mockChatHistoryService) SaveMessage(userID uint, role, content, audioURL, imageURL string, transactions []entity.SavedTransaction) error {
	return nil
}

func (m *mockChatHistoryService) SaveMessageWithReply(userID uint, role, content, audioURL, imageURL string, transactions []entity.SavedTransaction, replyToID *uint, replyToRole, replyToContent string) (*entity.ChatMessage, error) {
	return &entity.ChatMessage{
		ID:             1,
		UserID:         userID,
		Role:           role,
		Content:        content,
		ReplyToID:      replyToID,
		ReplyToRole:    replyToRole,
		ReplyToContent: replyToContent,
	}, nil
}

func (m *mockChatHistoryService) GetHistory(userID uint, limit int) ([]entity.ChatMessage, error) {
	return nil, nil
}

func (m *mockChatHistoryService) GetMessageByID(id uint, userID uint) (*entity.ChatMessage, error) {
	return &entity.ChatMessage{ID: id, UserID: userID, Role: "user", Content: "test"}, nil
}

func (m *mockChatHistoryService) FindNextAssistantMessage(userID uint, userMsgID uint) (*entity.ChatMessage, error) {
	return nil, nil
}

func (m *mockChatHistoryService) UpdateMessageContent(id uint, userID uint, content string) (*entity.ChatMessage, error) {
	return &entity.ChatMessage{ID: id, UserID: userID, Role: "user", Content: content, IsEdited: true}, nil
}

func (m *mockChatHistoryService) DeleteMessage(id uint, userID uint) error {
	return nil
}

func (m *mockChatHistoryService) ClearHistory(userID uint) error {
	return nil
}

func setupAIApp(aiSvc service.AIService, chatbotSvc *service.ChatbotService) (*fiber.App, AIHandler) {
	app := fiber.New()

	chatHistSvc := &mockChatHistoryService{}

	h := NewAIHandler(aiSvc, chatbotSvc, chatHistSvc)

	app.Post("/api/ai/chat", func(c *fiber.Ctx) error {
		c.Locals("userID", uint(1))
		return h.ChatMessage(c)
	})

	app.Post("/api/ai/chat/unauth", func(c *fiber.Ctx) error {
		return h.ChatMessage(c)
	})

	app.Post("/api/ai/chat/stream", func(c *fiber.Ctx) error {
		c.Locals("userID", uint(1))
		return h.ChatMessageStream(c)
	})

	app.Post("/api/ai/chat/stream/unauth", func(c *fiber.Ctx) error {
		return h.ChatMessageStream(c)
	})

	app.Put("/api/ai/chat/messages/:id", func(c *fiber.Ctx) error {
		c.Locals("userID", uint(1))
		return h.UpdateChatMessage(c)
	})

	app.Delete("/api/ai/chat/messages/:id", func(c *fiber.Ctx) error {
		c.Locals("userID", uint(1))
		return h.DeleteChatMessage(c)
	})

	return app, h
}

func TestAIHandler_ChatMessage_Unauthorized(t *testing.T) {
	mockAI := new(mockAIService)
	app, _ := setupAIApp(mockAI, &service.ChatbotService{})

	req := httptest.NewRequest("POST", "/api/ai/chat/unauth", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAIHandler_ChatMessage_EmptyMessage(t *testing.T) {
	mockAI := new(mockAIService)
	app, _ := setupAIApp(mockAI, &service.ChatbotService{})

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	_ = writer.Close()

	req := httptest.NewRequest("POST", "/api/ai/chat", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAIHandler_ChatMessageStream_Unauthorized(t *testing.T) {
	mockAI := new(mockAIService)
	app, _ := setupAIApp(mockAI, &service.ChatbotService{})

	req := httptest.NewRequest("POST", "/api/ai/chat/stream/unauth", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAIHandler_FormatActionSummary(t *testing.T) {
	saved := []entity.SavedTransaction{
		{
			Action:       "transfer",
			Description:  "Transfer saldo",
			Amount:       50000,
			WalletName:   "BCA",
			ToWalletName: "GoPay",
		},
		{
			Action:      "pay_debt",
			Description: "Bayar Utang Budi",
			Amount:      30000,
			WalletName:  "BCA",
		},
		{
			Action:      "save_goal",
			Description: "Setor Tabungan Laptop",
			Amount:      100000,
			WalletName:  "Mandiri",
		},
		{
			Action:      "create_wishlist",
			Description: "Sepatu Lari [medium]",
			Amount:      1200000,
		},
	}

	summary := formatActionSummary(saved)
	assert.Contains(t, summary, "Transfer saldo berhasil diproses!")
	assert.Contains(t, summary, "BCA ➡️ GoPay")
	assert.Contains(t, summary, "Bayar Utang Budi")
	assert.Contains(t, summary, "Setor Tabungan Laptop")
	assert.Contains(t, summary, "Sepatu Lari")
}

func TestAIHandler_ChatMessage_Success(t *testing.T) {
	mockAI := new(mockAIService)
	mockAI.On("Chat", mock.MatchedBy(func(p service.ChatParams) bool {
		return p.Message == "halo cuan"
	})).Return(&entity.ChatAIResponse{
		Reply:         "Halo juga! Ada yang bisa saya bantu?",
		IsTransaction: false,
	}, nil)

	app, _ := setupAIApp(mockAI, &service.ChatbotService{})

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("message", "halo cuan")
	_ = writer.Close()

	req := httptest.NewRequest("POST", "/api/ai/chat", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var chatResp entity.ChatResponse
	_ = json.NewDecoder(resp.Body).Decode(&chatResp)
	assert.Equal(t, "Halo juga! Ada yang bisa saya bantu?", chatResp.Reply)
	assert.Equal(t, uint(1), chatResp.UserMessageID)
	assert.Equal(t, uint(1), chatResp.AssistantMessageID)
}

func TestAIHandler_DeleteChatMessage_Success(t *testing.T) {
	mockAI := new(mockAIService)
	app, _ := setupAIApp(mockAI, &service.ChatbotService{})

	req := httptest.NewRequest("DELETE", "/api/ai/chat/messages/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestAIHandler_DeleteChatMessage_LargeID(t *testing.T) {
	mockAI := new(mockAIService)
	app, _ := setupAIApp(mockAI, &service.ChatbotService{})

	// 1791095355005 exceeds 32-bit uint and tests 64-bit parsing
	req := httptest.NewRequest("DELETE", "/api/ai/chat/messages/1791095355005", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	// Should not return 400 Bad Request ("ID pesan tidak valid")
	assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAIHandler_DeleteChatMessage_InvalidID(t *testing.T) {
	mockAI := new(mockAIService)
	app, _ := setupAIApp(mockAI, &service.ChatbotService{})

	req := httptest.NewRequest("DELETE", "/api/ai/chat/messages/invalid-id", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAIHandler_UpdateChatMessage_InvalidID(t *testing.T) {
	mockAI := new(mockAIService)
	app, _ := setupAIApp(mockAI, &service.ChatbotService{})

	req := httptest.NewRequest("PUT", "/api/ai/chat/messages/not-a-number", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAIHandler_UpdateChatMessage_Success(t *testing.T) {
	mockAI := new(mockAIService)
	mockAI.On("ChatStream", mock.MatchedBy(func(p service.ChatParams) bool {
		return p.Message == "beli es teh 5rb"
	}), mock.Anything).Return(&entity.ChatAIResponse{
		Reply: "Oke, dicatat es teh 5rb",
	}, nil)

	app, _ := setupAIApp(mockAI, &service.ChatbotService{})

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("message", "beli es teh 5rb")
	_ = writer.Close()

	req := httptest.NewRequest("PUT", "/api/ai/chat/messages/1", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}
