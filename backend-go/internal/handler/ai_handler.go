package handler

import (
	"bufio"
	"bytes"
	"cuan-backend/internal/entity"
	"cuan-backend/internal/service"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	MaxImageSize = 5 << 20  // 5MB
	MaxAudioSize = 10 << 20 // 10MB
)

type AIHandler interface {
	ChatMessage(c *fiber.Ctx) error
	ChatMessageStream(c *fiber.Ctx) error
}

type aiHandler struct {
	aiService      service.AIService
	chatbotService *service.ChatbotService
}

func NewAIHandler(aiService service.AIService, chatbotService *service.ChatbotService) AIHandler {
	return &aiHandler{
		aiService:      aiService,
		chatbotService: chatbotService,
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

	voiceFile, err := c.FormFile("voice")
	if err == nil && voiceFile != nil {
		if voiceFile.Size > MaxAudioSize {
			return c.Status(http.StatusRequestEntityTooLarge).JSON(fiber.Map{
				"error": fmt.Sprintf("Ukuran audio maksimal %dMB", MaxAudioSize>>20),
			})
		}
		savedPath, err := processAndSaveFile(voiceFile)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Gagal menyimpan audio: " + err.Error(),
			})
		}
		audioURL = savedPath

		diskPath := "." + savedPath
		transcription, err := h.aiService.ProcessVoice(diskPath)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Gagal memproses audio: " + err.Error(),
			})
		}
		if message == "" {
			message = "[TRANSKRIPSI SUARA - mungkin ada kesalahan fonetik, tolong koreksi]: " + transcription
		} else {
			message = message + "\n\n[TRANSKRIPSI SUARA - mungkin ada kesalahan fonetik, tolong koreksi]: " + transcription
		}
	}

	imageFile, err := c.FormFile("image")
	if err == nil && imageFile != nil {
		if imageFile.Size > MaxImageSize {
			return c.Status(http.StatusRequestEntityTooLarge).JSON(fiber.Map{
				"error": fmt.Sprintf("Ukuran gambar maksimal %dMB", MaxImageSize>>20),
			})
		}
		savedPath, err := processAndSaveFile(imageFile)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Gagal menyimpan gambar: " + err.Error(),
			})
		}
		savedImageURL = savedPath

		diskPath := "." + savedPath
		b64, err := readFileAsBase64(diskPath)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Gagal memproses gambar: " + err.Error(),
			})
		}
		imageBase64 = b64

		if message == "" {
			message = "Tolong analisis gambar ini. Jika ini struk belanja, identifikasi item dan harganya."
		}
	}

	if message == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Pesan tidak boleh kosong",
		})
	}

	userContext := h.chatbotService.GetUserContext(userID, message)
	aiResponse, err := h.aiService.Chat(message, imageBase64, userContext)
	if err != nil {
		fmt.Printf("[ERROR] AI Chat failed: %v\n", err)
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
			fmt.Printf("[ERROR] SaveTransactions failed: %v\n", err)
			response.Reply += "\n\n⚠️ Transaksi terdeteksi tapi gagal disimpan: " + err.Error()
		} else if len(saved) > 0 {
			response.Transactions = saved
			summary := "\n\n✅ Transaksi berhasil dicatat!"
			for _, s := range saved {
				summary += fmt.Sprintf("\n📝 %s — Rp%s (%s) | 🏦 %s | 📂 %s",
					s.Description,
					formatCurrency(s.Amount),
					s.Type,
					s.WalletName,
					s.CategoryName,
				)
			}
			response.Reply += summary
		}
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

	var audioURL string
	var savedImageURL string
	var diskVoicePath string
	var diskImagePath string

	voiceFile, err := c.FormFile("voice")
	if err == nil && voiceFile != nil {
		if voiceFile.Size > MaxAudioSize {
			return c.Status(http.StatusRequestEntityTooLarge).JSON(fiber.Map{"error": "Audio too large"})
		}
		savedPath, err := processAndSaveFile(voiceFile)
		if err != nil {
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
		savedPath, err := processAndSaveFile(imageFile)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Gagal menyimpan gambar: %s", err.Error())})
		}
		savedImageURL = savedPath
		diskImagePath = "." + savedPath
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		writeSSE(w, "status", "Mempersiapkan...")

		if diskVoicePath != "" {
			writeSSE(w, "status", "Mentranskripsi suara...")
			transcription, err := h.aiService.ProcessVoice(diskVoicePath)
			if err != nil {
				safeError, _ := json.Marshal(map[string]string{"error": "Gagal memproses audio: " + err.Error()})
				writeSSE(w, "error", string(safeError))
				return
			}
			if message == "" {
				message = "[TRANSKRIPSI SUARA - mungkin ada kesalahan fonetik, tolong koreksi]: " + transcription
			} else {
				message = message + "\n\n[TRANSKRIPSI SUARA - mungkin ada kesalahan fonetik, tolong koreksi]: " + transcription
			}
		}

		if diskImagePath != "" {
			b64, err := readFileAsBase64(diskImagePath)
			if err != nil {
				safeError, _ := json.Marshal(map[string]string{"error": "Gagal memproses gambar: " + err.Error()})
				writeSSE(w, "error", string(safeError))
				return
			}
			imageBase64 = b64
			if message == "" {
				message = "Tolong analisis gambar ini. Jika ini struk belanja, identifikasi item dan harganya."
			}
		}

		if message == "" {
			safeError, _ := json.Marshal(map[string]string{"error": "Message required"})
			writeSSE(w, "error", string(safeError))
			return
		}

		botStatus := "Sedang berpikir..."
		if diskImagePath != "" {
			botStatus = "Menganalisis gambar..."
		}
		writeSSE(w, "status", botStatus)

		userContext := h.chatbotService.GetUserContext(userID, message)

		aiResponse, err := h.aiService.ChatStream(message, imageBase64, userContext, func(token string) error {
			safeToken, _ := json.Marshal(map[string]string{"content": token})
			return writeSSE(w, "token", string(safeToken))
		})

		if err != nil {
			fmt.Printf("[ERROR] Stream failed: %v\n", err)
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
			writeSSE(w, "status", "Menyimpan transaksi...")
			saved, err := h.chatbotService.SaveTransactions(userID, aiResponse.Transactions)
			if err != nil {
				errMsg := "\n\n(Gagal menyimpan transaksi: " + err.Error() + ")"
				response.Reply += errMsg
				safeToken, _ := json.Marshal(map[string]string{"content": errMsg})
				writeSSE(w, "token", string(safeToken))
			} else if len(saved) > 0 {
				response.Transactions = saved
				summary := "\n\n✅ Transaksi berhasil dicatat!"
				for _, s := range saved {
					summary += fmt.Sprintf("\n📝 %s — Rp%s", s.Description, formatCurrency(s.Amount))
				}

				for _, char := range summary {
					safeToken, _ := json.Marshal(map[string]string{"content": string(char)})
					writeSSE(w, "token", string(safeToken))
				}

				response.Reply += summary
			}
		}

		finalJSON, _ := json.Marshal(response)
		writeSSE(w, "done", string(finalJSON))
	})

	return nil
}

func writeSSE(w *bufio.Writer, event, data string) error {
	_, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
	if err != nil {
		return err
	}
	return w.Flush()
}

func formatCurrency(amount float64) string {
	if amount == float64(int64(amount)) {
		return fmt.Sprintf("%d", int64(amount))
	}
	return fmt.Sprintf("%.0f", amount)
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

func removeTempWavFile(path string) {
	wavPath := strings.TrimSuffix(path, ".ogg") + ".wav"
	os.Remove(wavPath)
	wavPath = strings.TrimSuffix(path, ".webm") + ".wav"
	os.Remove(wavPath)
}
