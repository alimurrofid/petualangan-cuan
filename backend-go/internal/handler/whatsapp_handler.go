package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"cuan-backend/internal/entity"
	"cuan-backend/internal/service"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type WhatsAppHandler interface {
	HandleWebhook(c *fiber.Ctx) error
}

type whatsAppHandler struct {
	waSvc         service.WhatsAppService
	webhookSecret string
}

func NewWhatsAppHandler(waSvc service.WhatsAppService, webhookSecret string) WhatsAppHandler {
	return &whatsAppHandler{
		waSvc:         waSvc,
		webhookSecret: webhookSecret,
	}
}

// HandleWebhook godoc
// @Summary Handle WhatsApp webhook
// @Description Receives events from wa-gateway and processes incoming messages through AI chatbot
// @Tags webhook
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/webhook/whatsapp [post]
func (h *whatsAppHandler) HandleWebhook(c *fiber.Ctx) error {
	if h.webhookSecret != "" {
		signature := c.Get("X-Hub-Signature-256")
		if !h.verifySignature(c.Body(), signature) {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid webhook signature",
			})
		}
	}

	var event entity.WAWebhookEvent
	if err := json.Unmarshal(c.Body(), &event); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid payload",
		})
	}

	if event.Event != "message" {
		return c.JSON(fiber.Map{"status": "ignored", "event": event.Event})
	}
	go func() {
		if err := h.waSvc.ProcessMessage(event); err != nil {
			fmt.Printf("[WA][ERROR] ProcessMessage gagal: %v\n", err)
		}
	}()

	return c.JSON(fiber.Map{"status": "received"})
}

func (h *whatsAppHandler) verifySignature(body []byte, signature string) bool {
	expected := strings.TrimPrefix(signature, "sha256=")
	if expected == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(body)
	computed := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(computed), []byte(expected))
}
