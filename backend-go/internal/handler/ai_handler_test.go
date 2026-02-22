package handler

import (
	"bytes"
	"cuan-backend/internal/service"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// mockAIService and mockChatbotService can be instantiated as empty or properly mocked structures
// depending on how deep the handler test needs to be. For fiber handler routing tests,
// we can test the failure cases where message/files are missing, or contexts are unauthorized.

func setupAIApp() (*fiber.App, AIHandler) {
	app := fiber.New()

	// Create mock services
	aiSvc := &service.AIService{}           // Simplified mock
	chatbotSvc := &service.ChatbotService{} // Simplified mock

	h := NewAIHandler(aiSvc, chatbotSvc)

	// Register route with mock JWT Local setup
	app.Post("/api/ai/chat", func(c *fiber.Ctx) error {
		// Mock valid user login
		c.Locals("userID", uint(1))
		return h.ChatMessage(c)
	})

	app.Post("/api/ai/chat/unauth", func(c *fiber.Ctx) error {
		// Mock unauthorized (missing userID)
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
