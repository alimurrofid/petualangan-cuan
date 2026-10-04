package handler

import (
	"bufio"
	"bytes"
	"cuan-backend/internal/entity"
	"cuan-backend/internal/service"
	"cuan-backend/pkg/utils"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/gofiber/fiber/v2"
)

const (
	MaxImageSize = 5 << 20  // 5MB
	MaxAudioSize = 10 << 20 // 10MB
)

type AIHandler interface {
	ChatMessage(c *fiber.Ctx) error
	ChatMessageStream(c *fiber.Ctx) error
	GetChatHistory(c *fiber.Ctx) error
	ClearChatHistory(c *fiber.Ctx) error
	UpdateChatMessage(c *fiber.Ctx) error
	DeleteChatMessage(c *fiber.Ctx) error
}

type aiHandler struct {
	aiService      service.AIService
	chatbotService *service.ChatbotService
	chatHistorySvc service.ChatHistoryService
}

func NewAIHandler(aiService service.AIService, chatbotService *service.ChatbotService, chatHistorySvc service.ChatHistoryService) AIHandler {
	return &aiHandler{
		aiService:      aiService,
		chatbotService: chatbotService,
		chatHistorySvc: chatHistorySvc,
	}
}

// ChatMessage godoc
// @Summary Send chat message
// @Description Send text, image, or voice message to the AI chatbot
// @Tags ai
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param message formData string false "Text message"
// @Param image formData file false "Image attachment"
// @Param voice formData file false "Voice attachment"
// @Success 200 {object} entity.ChatResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 413 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/ai/chat [post]
func (h *aiHandler) ChatMessage(c *fiber.Ctx) error {
	message := c.FormValue("message")
	var imageBase64 string

	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var audioURL string
	var savedImageURL string
	var audioBase64 string
	var audioFormat string
	var audioMimeType string

	voiceFile, err := c.FormFile("voice")
	if err == nil && voiceFile != nil {
		if voiceFile.Size > MaxAudioSize {
			return c.Status(http.StatusRequestEntityTooLarge).JSON(fiber.Map{
				"error": fmt.Sprintf("Ukuran audio maksimal %dMB", MaxAudioSize>>20),
			})
		}
		b64, err := fileHeaderToBase64(voiceFile)
		if err != nil {
			reqID, _ := c.Locals("requestid").(string)
			log.Error().Str("request_id", reqID).Err(err).Msg("Failed to read audio as base64")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Gagal memproses audio: " + err.Error(),
			})
		}
		audioBase64 = b64

		savedPath, err := processAndSaveFile(voiceFile)
		if err != nil {
			reqID, _ := c.Locals("requestid").(string)
			log.Error().Str("request_id", reqID).Err(err).Msg("Failed to save audio")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Gagal menyimpan audio: " + err.Error(),
			})
		}
		audioURL = savedPath
		meta := utils.ResolveAudioMetadata(voiceFile.Filename)
		audioFormat = meta.Format
		audioMimeType = meta.MimeType

		if message == "" {
			message = "Tolong dengarkan rekaman audio ini. Jika ini transaksi keuangan, catat dan identifikasi rinciannya."
		}
	}

	imageFile, err := c.FormFile("image")
	if err == nil && imageFile != nil {
		if imageFile.Size > MaxImageSize {
			return c.Status(http.StatusRequestEntityTooLarge).JSON(fiber.Map{
				"error": fmt.Sprintf("Ukuran gambar maksimal %dMB", MaxImageSize>>20),
			})
		}
		b64, err := fileHeaderToBase64(imageFile)
		if err != nil {
			reqID, _ := c.Locals("requestid").(string)
			log.Error().Str("request_id", reqID).Err(err).Msg("Failed to process image")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Gagal memproses gambar: " + err.Error(),
			})
		}
		imageBase64 = b64

		savedPath, err := processAndSaveFile(imageFile)
		if err != nil {
			reqID, _ := c.Locals("requestid").(string)
			log.Error().Str("request_id", reqID).Err(err).Msg("Failed to save image")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Gagal menyimpan gambar: " + err.Error(),
			})
		}
		savedImageURL = savedPath

		if message == "" {
			message = "Tolong analisis gambar ini. Jika ini struk belanja, identifikasi item dan harganya."
		}
	}

	if message == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Pesan tidak boleh kosong",
		})
	}

	// Cek apakah pesan mengutip/membalas pesan lain
	replyToIDStr := c.FormValue("reply_to_id")
	var replyToIDPtr *uint
	var replyToRole string
	var replyToContent string
	promptMessage := message

	if replyToIDStr != "" {
		if idNum, err := strconv.ParseUint(replyToIDStr, 10, 64); err == nil {
			id := uint(idNum)
			if quoted, err := h.chatHistorySvc.GetMessageByID(id, userID); err == nil && quoted != nil {
				replyToIDPtr = &quoted.ID
				replyToRole = quoted.Role
				replyToContent = quoted.Content
				promptMessage = fmt.Sprintf("[Membalas %s: \"%s\"]\n%s", replyToRole, replyToContent, message)
			}
		}
	}

	// Simpan pesan user ke history
	userContent := message
	savedUserMsg, err := h.chatHistorySvc.SaveMessageWithReply(userID, "user", userContent, audioURL, savedImageURL, nil, replyToIDPtr, replyToRole, replyToContent)
	if err != nil {
		log.Warn().Err(err).Msg("Gagal menyimpan pesan user ke history")
	}

	userContext := h.chatbotService.GetUserContext(userID, promptMessage)
	aiResponse, err := h.aiService.Chat(service.ChatParams{
		Message:       promptMessage,
		ImageBase64:   imageBase64,
		AudioBase64:   audioBase64,
		AudioFormat:   audioFormat,
		AudioMimeType: audioMimeType,
		UserContext:   userContext,
	})
	if err != nil {
		reqID, _ := c.Locals("requestid").(string)
		log.Error().Str("request_id", reqID).Err(err).Msg("AI Chat failed")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mendapatkan respons AI: " + err.Error(),
		})
	}

	response := entity.ChatResponse{
		Reply:    aiResponse.Reply,
		AudioURL: audioURL,
		ImageURL: savedImageURL,
	}

	if aiResponse.IsTransaction && len(aiResponse.Transactions) > 0 {
		saved, err := h.chatbotService.SaveTransactions(userID, aiResponse.Transactions)
		if err != nil {
			reqID, _ := c.Locals("requestid").(string)
			log.Error().Str("request_id", reqID).Err(err).Msg("SaveTransactions failed")
			response.Reply += "\n\n⚠️ Transaksi terdeteksi tapi gagal diproses: " + err.Error()
		} else if len(saved) > 0 {
			response.Transactions = saved
			summary := formatActionSummary(saved)
			response.Reply += summary
		}
	}

	// Simpan balasan AI ke history
	savedAssistantMsg, err := h.chatHistorySvc.SaveMessageWithReply(userID, "assistant", response.Reply, "", "", response.Transactions, nil, "", "")
	if err != nil {
		log.Warn().Err(err).Msg("Gagal menyimpan balasan AI ke history")
	}

	if savedUserMsg != nil {
		response.UserMessageID = savedUserMsg.ID
	}
	if savedAssistantMsg != nil {
		response.AssistantMessageID = savedAssistantMsg.ID
		response.ID = savedAssistantMsg.ID
	}

	return c.JSON(response)
}

// ChatMessageStream godoc
// @Summary Send chat message via stream
// @Description Stream AI chatbot response using Server-Sent Events (SSE)
// @Tags ai
// @Accept multipart/form-data
// @Produce text/event-stream
// @Security BearerAuth
// @Param message formData string false "Text message"
// @Param image formData file false "Image attachment"
// @Param voice formData file false "Voice attachment"
// @Success 200 {string} string "SSE stream"
// @Router /api/ai/chat/stream [post]
func (h *aiHandler) ChatMessageStream(c *fiber.Ctx) error {
	message := c.FormValue("message")
	var imageBase64 string
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	reqID, _ := c.Locals("requestid").(string)

	var audioURL string
	var savedImageURL string
	var diskVoicePath string
	var diskImagePath string
	var audioBase64 string
	var audioFormat string
	var audioMimeType string

	voiceFile, err := c.FormFile("voice")
	if err == nil && voiceFile != nil {
		if voiceFile.Size > MaxAudioSize {
			return c.Status(http.StatusRequestEntityTooLarge).JSON(fiber.Map{"error": "Audio too large"})
		}
		b64, err := fileHeaderToBase64(voiceFile)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Gagal memproses audio: %s", err.Error())})
		}
		audioBase64 = b64

		savedPath, err := processAndSaveFile(voiceFile)
		if err != nil {
			log.Error().Str("request_id", reqID).Err(err).Msg("Failed to save audio")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Gagal menyimpan audio: %s", err.Error())})
		}
		audioURL = savedPath
		diskVoicePath = "." + savedPath
	}

	imageFile, err := c.FormFile("image")
	if err == nil && imageFile != nil {
		if imageFile.Size > MaxImageSize {
			return c.Status(http.StatusRequestEntityTooLarge).JSON(fiber.Map{"error": "Image too large"})
		}
		b64, err := fileHeaderToBase64(imageFile)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Gagal memproses gambar: %s", err.Error())})
		}
		imageBase64 = b64

		savedPath, err := processAndSaveFile(imageFile)
		if err != nil {
			log.Error().Str("request_id", reqID).Err(err).Msg("Failed to save image")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Gagal menyimpan gambar: %s", err.Error())})
		}
		savedImageURL = savedPath
		diskImagePath = "." + savedPath
	}

	replyToIDStr := c.FormValue("reply_to_id")
	var replyToIDPtr *uint
	var replyToRole string
	var replyToContent string

	if replyToIDStr != "" {
		if idNum, err := strconv.ParseUint(replyToIDStr, 10, 64); err == nil {
			id := uint(idNum)
			if quoted, err := h.chatHistorySvc.GetMessageByID(id, userID); err == nil && quoted != nil {
				replyToIDPtr = &quoted.ID
				replyToRole = quoted.Role
				replyToContent = quoted.Content
			}
		}
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		writeSSE(w, "status", "Mempersiapkan...")

		if diskVoicePath != "" {
			meta := utils.ResolveAudioMetadata(diskVoicePath)
			audioFormat = meta.Format
			audioMimeType = meta.MimeType

			if message == "" {
				message = "Tolong dengarkan rekaman audio ini. Jika ini transaksi keuangan, catat dan identifikasi rinciannya."
			}
		}

		if diskImagePath != "" {
			if message == "" {
				message = "Tolong analisis gambar ini. Jika ini struk belanja, identifikasi item dan harganya."
			}
		}

		if message == "" {
			safeError, _ := json.Marshal(map[string]string{"error": "Message required"})
			writeSSE(w, "error", string(safeError))
			return
		}

		promptMessage := message
		if replyToContent != "" {
			promptMessage = fmt.Sprintf("[Membalas %s: \"%s\"]\n%s", replyToRole, replyToContent, message)
		}

		// Simpan pesan user ke history
		savedUserMsg, err := h.chatHistorySvc.SaveMessageWithReply(userID, "user", message, audioURL, savedImageURL, nil, replyToIDPtr, replyToRole, replyToContent)
		if err != nil {
			log.Warn().Err(err).Msg("Gagal menyimpan pesan user ke history")
		} else if savedUserMsg != nil {
			userMsgData, _ := json.Marshal(map[string]interface{}{
				"id":              savedUserMsg.ID,
				"user_message_id": savedUserMsg.ID,
			})
			writeSSE(w, "user_message", string(userMsgData))
		}

		botStatus := "Sedang berpikir..."
		if diskImagePath != "" {
			botStatus = "Menganalisis gambar..."
		} else if diskVoicePath != "" {
			botStatus = "Mendengarkan audio..."
		}
		writeSSE(w, "status", botStatus)

		userContext := h.chatbotService.GetUserContext(userID, promptMessage)

		aiResponse, err := h.aiService.ChatStream(service.ChatParams{
			Message:       promptMessage,
			ImageBase64:   imageBase64,
			AudioBase64:   audioBase64,
			AudioFormat:   audioFormat,
			AudioMimeType: audioMimeType,
			UserContext:   userContext,
		}, func(token string) error {
			safeToken, _ := json.Marshal(map[string]string{"content": token})
			return writeSSE(w, "token", string(safeToken))
		})

		if err != nil {
			log.Error().Str("request_id", reqID).Err(err).Msg("Stream failed")
			safeError, _ := json.Marshal(map[string]string{"error": err.Error()})
			writeSSE(w, "error", string(safeError))
			return
		}

		response := entity.ChatResponse{
			Reply:    aiResponse.Reply,
			AudioURL: audioURL,
			ImageURL: savedImageURL,
		}

		if aiResponse.IsTransaction && len(aiResponse.Transactions) > 0 {
			writeSSE(w, "status", "Memproses transaksi...")
			saved, err := h.chatbotService.SaveTransactions(userID, aiResponse.Transactions)
			if err != nil {
				errMsg := "\n\n(Gagal memproses transaksi: " + err.Error() + ")"
				response.Reply += errMsg
				safeToken, _ := json.Marshal(map[string]string{"content": errMsg})
				writeSSE(w, "token", string(safeToken))
			} else if len(saved) > 0 {
				response.Transactions = saved
				summary := formatActionSummary(saved)

				for _, char := range summary {
					safeToken, _ := json.Marshal(map[string]string{"content": string(char)})
					writeSSE(w, "token", string(safeToken))
				}

				response.Reply += summary
			}
		}

		// Simpan balasan AI ke history setelah streaming selesai
		savedAssistantMsg, err := h.chatHistorySvc.SaveMessageWithReply(userID, "assistant", response.Reply, "", "", response.Transactions, nil, "", "")
		if err != nil {
			log.Warn().Err(err).Msg("Gagal menyimpan balasan AI ke history")
		}

		if savedUserMsg != nil {
			response.UserMessageID = savedUserMsg.ID
		}
		if savedAssistantMsg != nil {
			response.AssistantMessageID = savedAssistantMsg.ID
			response.ID = savedAssistantMsg.ID
		}

		finalJSON, _ := json.Marshal(response)
		writeSSE(w, "done", string(finalJSON))
	})

	return nil
}

// GetChatHistory godoc
// @Summary Get chat history
// @Description Get the chat message history for the authenticated user
// @Tags ai
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of messages to return (default 100)"
// @Success 200 {array} entity.ChatMessage
// @Failure 401 {object} map[string]interface{}
// @Router /api/ai/chat/history [get]
func (h *aiHandler) GetChatHistory(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	limitStr := c.Query("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 100
	}

	messages, err := h.chatHistorySvc.GetHistory(userID, limit)
	if err != nil {
		reqID, _ := c.Locals("requestid").(string)
		log.Error().Str("request_id", reqID).Err(err).Msg("Failed to retrieve chat history")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil riwayat chat"})
	}

	return c.JSON(messages)
}

// ClearChatHistory godoc
// @Summary Clear chat history
// @Description Delete all chat messages for the authenticated user
// @Tags ai
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/ai/chat/history [delete]
func (h *aiHandler) ClearChatHistory(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	if err := h.chatHistorySvc.ClearHistory(userID); err != nil {
		reqID, _ := c.Locals("requestid").(string)
		log.Error().Str("request_id", reqID).Err(err).Msg("Failed to clear chat history")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menghapus riwayat chat"})
	}

	return c.JSON(fiber.Map{"message": "Riwayat chat berhasil dihapus"})
}

// UpdateChatMessage godoc
// @Summary Edit user chat message and regenerate response
// @Description Updates user message content, cancels previous assistant actions if any, and regenerates response
// @Tags ai
// @Accept json
// @Accept mpfd
// @Produce text/event-stream
// @Produce json
// @Security BearerAuth
// @Param id path int true "Message ID"
// @Param message formData string false "New message text"
// @Success 200 {object} entity.ChatResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/ai/chat/messages/{id} [put]
func (h *aiHandler) UpdateChatMessage(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	idParam := c.Params("id")
	idNum, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "ID pesan tidak valid",
		})
	}
	messageID := uint(idNum)

	msg, err := h.chatHistorySvc.GetMessageByID(messageID, userID)
	if err != nil || msg == nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Pesan tidak ditemukan",
		})
	}

	if msg.Role != "user" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Hanya pesan kamu yang dapat diedit",
		})
	}

	newText := strings.TrimSpace(c.FormValue("message"))
	if newText == "" {
		var body struct {
			Message string `json:"message"`
			Content string `json:"content"`
		}
		if err := c.BodyParser(&body); err == nil {
			if body.Message != "" {
				newText = strings.TrimSpace(body.Message)
			} else {
				newText = strings.TrimSpace(body.Content)
			}
		}
	}

	if newText == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Teks pesan baru tidak boleh kosong",
		})
	}

	// 1. Batalkan aksi/transaksi asisten berikutnya dan hapus pesan asisten lama jika ada
	nextAssistant, _ := h.chatHistorySvc.FindNextAssistantMessage(userID, messageID)
	if nextAssistant != nil {
		if len(nextAssistant.Transactions) > 0 {
			if err := h.chatbotService.RollbackSavedTransactions(userID, nextAssistant.Transactions); err != nil {
				log.Warn().Err(err).Msg("Rollback transaksi asisten lama gagal sebagian")
			}
		}
		_ = h.chatHistorySvc.DeleteMessage(nextAssistant.ID, userID)
	}

	// 2. Perbarui konten pesan pengguna
	if _, err := h.chatHistorySvc.UpdateMessageContent(messageID, userID, newText); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal memperbarui pesan: " + err.Error(),
		})
	}

	// 3. Re-generate AI response via SSE
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		writeSSE(w, "status", "Memperbarui respons...")

		promptMessage := newText
		if msg.ReplyToContent != "" {
			promptMessage = fmt.Sprintf("[Membalas %s: \"%s\"]\n%s", msg.ReplyToRole, msg.ReplyToContent, newText)
		}

		userContext := h.chatbotService.GetUserContext(userID, promptMessage)

		aiResponse, err := h.aiService.ChatStream(service.ChatParams{
			Message:     promptMessage,
			UserContext: userContext,
		}, func(token string) error {
			safeToken, _ := json.Marshal(map[string]string{"content": token})
			return writeSSE(w, "token", string(safeToken))
		})

		if err != nil {
			log.Error().Err(err).Msg("Stream failed on UpdateChatMessage")
			safeError, _ := json.Marshal(map[string]string{"error": err.Error()})
			writeSSE(w, "error", string(safeError))
			return
		}

		response := entity.ChatResponse{
			Reply: aiResponse.Reply,
		}

		if aiResponse.IsTransaction && len(aiResponse.Transactions) > 0 {
			writeSSE(w, "status", "Memproses transaksi baru...")
			saved, err := h.chatbotService.SaveTransactions(userID, aiResponse.Transactions)
			if err != nil {
				errMsg := "\n\n(Gagal memproses transaksi: " + err.Error() + ")"
				response.Reply += errMsg
				safeToken, _ := json.Marshal(map[string]string{"content": errMsg})
				writeSSE(w, "token", string(safeToken))
			} else if len(saved) > 0 {
				response.Transactions = saved
				summary := formatActionSummary(saved)

				for _, char := range summary {
					safeToken, _ := json.Marshal(map[string]string{"content": string(char)})
					writeSSE(w, "token", string(safeToken))
				}

				response.Reply += summary
			}
		}

		// Simpan balasan AI baru ke history
		savedAssistantMsg, err := h.chatHistorySvc.SaveMessageWithReply(userID, "assistant", response.Reply, "", "", response.Transactions, nil, "", "")
		if err != nil {
			log.Warn().Err(err).Msg("Gagal menyimpan balasan AI baru ke history")
		}

		response.UserMessageID = messageID
		if savedAssistantMsg != nil {
			response.AssistantMessageID = savedAssistantMsg.ID
			response.ID = savedAssistantMsg.ID
		}

		finalJSON, _ := json.Marshal(response)
		writeSSE(w, "done", string(finalJSON))
	})

	return nil
}

// DeleteChatMessage godoc
// @Summary Delete a chat message
// @Description Deletes a chat message by ID, optionally rolling back associated transactions
// @Tags ai
// @Produce json
// @Security BearerAuth
// @Param id path int true "Message ID"
// @Param rollback query bool false "Rollback associated transactions"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/ai/chat/messages/{id} [delete]
func (h *aiHandler) DeleteChatMessage(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	idParam := c.Params("id")
	idNum, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "ID pesan tidak valid",
		})
	}
	messageID := uint(idNum)

	msg, err := h.chatHistorySvc.GetMessageByID(messageID, userID)
	if err != nil || msg == nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Pesan tidak ditemukan",
		})
	}

	shouldRollback := c.Query("rollback") == "true" || c.Query("rollback") == "1"
	if shouldRollback && len(msg.Transactions) > 0 {
		if err := h.chatbotService.RollbackSavedTransactions(userID, msg.Transactions); err != nil {
			log.Warn().Err(err).Msg("Gagal membatalkan sebagian transaksi saat menghapus pesan")
		}
	}

	if err := h.chatHistorySvc.DeleteMessage(messageID, userID); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal menghapus pesan: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Pesan berhasil dihapus",
	})
}

func writeSSE(w *bufio.Writer, event, data string) error {
	_, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
	if err != nil {
		return err
	}
	return w.Flush()
}

func formatCurrency(amount float64) string {
	n := int64(amount)
	if n < 0 {
		return "-" + formatCurrency(-amount)
	}
	str := fmt.Sprintf("%d", n)
	result := make([]byte, 0, len(str)+len(str)/3)
	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result = append(result, '.')
		}
		result = append(result, byte(c))
	}
	return string(result)
}

func readFileAsBase64(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", err
	}

	estimatedSize := int(info.Size()*4/3) + 4
	var buf bytes.Buffer
	buf.Grow(estimatedSize)

	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	if _, err := io.Copy(encoder, file); err != nil {
		return "", err
	}
	encoder.Close()

	return buf.String(), nil
}

func fileHeaderToBase64(file *multipart.FileHeader) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	if _, err := io.Copy(encoder, f); err != nil {
		return "", err
	}
	encoder.Close()
	return buf.String(), nil
}

func removeTempWavFile(path string) {
	wavPath := strings.TrimSuffix(path, ".ogg") + ".wav"
	os.Remove(wavPath)
	wavPath = strings.TrimSuffix(path, ".webm") + ".wav"
	os.Remove(wavPath)
}

func formatActionSummary(saved []entity.SavedTransaction) string {
	if len(saved) == 0 {
		return ""
	}
	summary := "\n\n"
	hasCreate, hasUpdate, hasDelete := false, false, false
	hasTransfer, hasPayDebt, hasSaveGoal, hasWishlist := false, false, false, false
	hasDebt, hasGoal := false, false

	for _, s := range saved {
		switch {
		case strings.HasPrefix(s.Action, "delete"):
			hasDelete = true
		case strings.HasPrefix(s.Action, "update"):
			hasUpdate = true
		case s.Action == "transfer":
			hasTransfer = true
		case s.Action == "pay_debt":
			hasPayDebt = true
		case s.Action == "save_goal":
			hasSaveGoal = true
		case s.Action == "create_wishlist":
			hasWishlist = true
		case s.Action == "create_debt":
			hasDebt = true
		case s.Action == "create_goal":
			hasGoal = true
		default:
			hasCreate = true
		}
	}

	if hasTransfer {
		summary += "✅ Transfer saldo berhasil diproses!"
	} else if hasPayDebt {
		summary += "✅ Pembayaran utang berhasil dicatat!"
	} else if hasSaveGoal {
		summary += "✅ Setoran tabungan berhasil dicatat!"
	} else if hasWishlist {
		summary += "✅ Item berhasil ditambahkan ke Wishlist!"
	} else if hasDebt {
		summary += "✅ Utang/piutang berhasil dicatat!"
	} else if hasGoal {
		summary += "✅ Target tabungan berhasil dibuat!"
	} else if hasDelete {
		summary += "✅ Data berhasil dihapus!"
	} else if hasUpdate {
		summary += "✅ Data berhasil diperbarui!"
	} else if hasCreate {
		summary += "✅ Transaksi berhasil dicatat!"
	}

	for _, s := range saved {
		switch s.Action {
		case "create_transaction":
			summary += fmt.Sprintf("\n📝 %s — Rp%s", s.Description, formatCurrency(s.Amount))
		case "update_transaction":
			summary += fmt.Sprintf("\n✏️ %s — Rp%s", s.Description, formatCurrency(s.Amount))
		case "delete_transaction":
			summary += fmt.Sprintf("\n🗑️ %s (Dihapus)", s.Description)
		case "transfer":
			summary += fmt.Sprintf("\n🔄 %s: Rp%s (%s ➡️ %s)", s.Description, formatCurrency(s.Amount), s.WalletName, s.ToWalletName)
		case "pay_debt":
			summary += fmt.Sprintf("\n🤝 %s — Rp%s (%s)", s.Description, formatCurrency(s.Amount), s.WalletName)
		case "save_goal":
			summary += fmt.Sprintf("\n🎯 %s — Rp%s (%s)", s.Description, formatCurrency(s.Amount), s.WalletName)
		case "create_wishlist":
			summary += fmt.Sprintf("\n⭐ %s — Rp%s", s.Description, formatCurrency(s.Amount))
		case "update_wishlist":
			summary += fmt.Sprintf("\n✏️ Wishlist: %s — Rp%s", s.Description, formatCurrency(s.Amount))
		case "delete_wishlist":
			summary += fmt.Sprintf("\n🗑️ Wishlist: %s (Dihapus)", s.Description)
		case "create_debt":
			summary += fmt.Sprintf("\n📝 Utang/Piutang: %s — Rp%s (%s)", s.Description, formatCurrency(s.Amount), s.WalletName)
		case "update_debt":
			summary += fmt.Sprintf("\n✏️ Utang/Piutang: %s — Rp%s (%s)", s.Description, formatCurrency(s.Amount), s.WalletName)
		case "delete_debt":
			summary += fmt.Sprintf("\n🗑️ Utang/Piutang: %s (Dihapus)", s.Description)
		case "create_goal":
			summary += fmt.Sprintf("\n🎯 Target Tabungan: %s — Rp%s", s.Description, formatCurrency(s.Amount))
		case "update_goal":
			summary += fmt.Sprintf("\n✏️ Target Tabungan: %s — Rp%s", s.Description, formatCurrency(s.Amount))
		case "delete_goal":
			summary += fmt.Sprintf("\n🗑️ Target Tabungan: %s (Dihapus)", s.Description)
		default:
			summary += fmt.Sprintf("\n📝 %s — Rp%s", s.Description, formatCurrency(s.Amount))
		}
	}
	return summary
}

