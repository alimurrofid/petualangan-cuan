package handler

import (
	"bytes"
	"context"
	"cuan-backend/internal/entity"
	aiprovider "cuan-backend/internal/provider/ai"
	"cuan-backend/internal/service"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

type mockAIService struct{}

func (m *mockAIService) Chat(_ string, _ string, _ string) (*entity.ChatAIResponse, error) {
	return nil, nil
}

func (m *mockAIService) ChatStream(_ string, _ string, _ string, _ func(string) error) (*entity.ChatAIResponse, error) {
	return nil, nil
}

func (m *mockAIService) ProcessVoice(_ string) (string, error) {
	return "", nil
}

type mockAIProvider struct{}

func (m *mockAIProvider) GenerateCompletion(_ context.Context, _ aiprovider.AIRequest) (string, error) {
	return "", nil
}

type mockChatHistoryService struct{}

func (m *mockChatHistoryService) SaveMessage(_ uint, _, _, _, _ string) error {
	return nil
}

func (m *mockChatHistoryService) GetHistory(_ uint, _ int) ([]entity.ChatMessage, error) {
	return nil, nil
}

func (m *mockChatHistoryService) ClearHistory(_ uint) error {
	return nil
}

func setupAIApp() (*fiber.App, AIHandler) {
	app := fiber.New()

	aiSvc := &mockAIService{}
	chatbotSvc := &service.ChatbotService{}
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

	return app, h
}

func TestAIHandler_ChatMessage_Unauthorized(t *testing.T) {
	app, _ := setupAIApp()

	req := httptest.NewRequest("POST", "/api/ai/chat/unauth", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAIHandler_ChatMessage_EmptyMessage(t *testing.T) {
	app, _ := setupAIApp()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	_ = writer.Close()

	req := httptest.NewRequest("POST", "/api/ai/chat", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAIHandler_ChatMessageStream_Unauthorized(t *testing.T) {
	app, _ := setupAIApp()

	req := httptest.NewRequest("POST", "/api/ai/chat/stream/unauth", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
