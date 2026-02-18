package handler

import (
	"bytes"
	"cuan-backend/internal/entity"
	"cuan-backend/internal/service"
	"encoding/base64"
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
}

type aiHandler struct {
	aiService      *service.AIService
	chatbotService *service.ChatbotService
}

func NewAIHandler(aiService *service.AIService, chatbotService *service.ChatbotService) AIHandler {
	return &aiHandler{
		aiService:      aiService,
		chatbotService: chatbotService,
	}
}

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

	userContext := h.chatbotService.GetUserContext(userID)
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
