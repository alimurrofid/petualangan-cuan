package handler

import (
	"cuan-backend/internal/service"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ChatHandler struct {
	Service *service.ChatService
}

func NewChatHandler(service *service.ChatService) *ChatHandler {
	return &ChatHandler{Service: service}
}

func (h *ChatHandler) SendMessage(c *fiber.Ctx) error {
	message := c.FormValue("message")
	
	// Handle file upload
	file, err := c.FormFile("file")
	var fileData []byte
	var mimeType string
	var attachmentPath string

	if err == nil {
		// File exists
		f, err := file.Open()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to open file"})
		}
		defer f.Close()

		fileData, err = io.ReadAll(f)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to read file"})
		}
		mimeType = file.Header.Get("Content-Type")

		// Save File Locally
		uploadDir := "./uploads"
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			os.MkdirAll(uploadDir, 0755)
		}

		// Generate unique filename
		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
		filePath := filepath.Join(uploadDir, filename)
		
		// Save file
		if err := c.SaveFile(file, filePath); err != nil {
			// Log error?
		} else {
			// Convert to URL path (assuming /uploads is static served)
			attachmentPath = "/uploads/" + filename
		}
	}

	// Get UserID from context (set by middleware)
	userID := uint(1) 
	
	if id, ok := c.Locals("userID").(float64); ok {
		userID = uint(id)
	}

	response, err := h.Service.ProcessMessage(userID, message, fileData, mimeType, attachmentPath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"response": response,
	})
}
