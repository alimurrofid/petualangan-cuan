package service

import (
	"bytes"
	"cuan-backend/internal/entity"
	"cuan-backend/internal/repository"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}


type WhatsAppService interface {
	ProcessMessage(event entity.WAWebhookEvent) error
}

type whatsAppService struct {
	userRepo       repository.UserRepository
	aiSvc          AIService
	chatbotSvc     *ChatbotService
	chatHistSvc    ChatHistoryService
	waGatewayURL  string
	waGatewayUser string
	waGatewayPass string
}

func NewWhatsAppService(
	userRepo repository.UserRepository,
	aiSvc AIService,
	chatbotSvc *ChatbotService,
	chatHistSvc ChatHistoryService,
	waGatewayURL string,
) WhatsAppService {
	return &whatsAppService{
		userRepo:      userRepo,
		aiSvc:         aiSvc,
		chatbotSvc:    chatbotSvc,
		chatHistSvc:   chatHistSvc,
		waGatewayURL:  waGatewayURL,
		waGatewayUser: os.Getenv("WA_GATEWAY_USERNAME"),
		waGatewayPass: os.Getenv("WA_GATEWAY_PASSWORD"),
	}
}

func (s *whatsAppService) ProcessMessage(event entity.WAWebhookEvent) error {
	msg := event.Payload

	if msg.IsFromMe {
		return nil
	}

	phone := extractPhone(msg.From)
	if phone == "" {
		return fmt.Errorf("gagal mengekstrak nomor dari JID: %s", msg.From)
	}

	user, err := s.userRepo.FindByPhone(phone)
	if err != nil {
		return s.sendWAMessage(msg.ChatID, event.DeviceID,
			"Halo! 👋 Nomor WA Anda belum terhubung ke akun Petualangan Cuan.\n\n"+
				"Silakan login ke aplikasi web, lalu masuk ke menu *Profil* dan isi kolom *Nomor HP* dengan nomor ini: "+phone)
	}

	var processingMsg string
	switch {
	case msg.Audio != "":
		processingMsg = "🎙️ Pesan suara diterima, sedang ditranskripsi dan diproses..."
	case msg.Image != "":
		processingMsg = "🖼️ Gambar diterima, sedang dianalisis..."
	default:
		processingMsg = "⏳ Pesan diterima, sedang diproses..."
	}
	_ = s.sendWAMessage(msg.ChatID, event.DeviceID, processingMsg)

	text := strings.TrimSpace(msg.Body)
	var imageBase64 string
	var audioTranscription string

	var savedAudioURL, savedImageURL string

	if msg.Audio != "" {
		audioData, err := s.downloadMediaFromGateway(msg.Audio)
		if err != nil {
			fmt.Printf("[WA] Gagal download audio: %v\n", err)
			_ = s.sendWAMessage(msg.ChatID, event.DeviceID, "❌ Gagal mengunduh pesan suara. Coba kirim ulang.")
			return err
		}
		
		// Simpan audio ke file sistem
		audioFilename := fmt.Sprintf("wa_audio_%d_%d.ogg", user.ID, time.Now().UnixNano())
		audioPath := filepath.Join("uploads", "audio", audioFilename)
		_ = os.MkdirAll(filepath.Dir(audioPath), 0755)
		if err := os.WriteFile(audioPath, audioData, 0644); err == nil {
			savedAudioURL = "/uploads/audio/" + audioFilename
		} else {
			fmt.Printf("[WA][WARN] Gagal menyimpan audio WA ke disk: %v\n", err)
		}

		transcription, err := s.transcribeAudio(audioData, msg.Audio)
		if err != nil {
			fmt.Printf("[WA] Gagal transkripsi audio: %v\n", err)
			_ = s.sendWAMessage(msg.ChatID, event.DeviceID, "❌ Gagal membaca pesan suara. Coba kirim ulang atau ketik pesannya.")
			return err
		}
		audioTranscription = transcription
	}

	if msg.Image != "" {
		imgData, err := s.downloadMediaFromGateway(msg.Image)
		if err != nil {
			fmt.Printf("[WA] Gagal download gambar: %v\n", err)
			_ = s.sendWAMessage(msg.ChatID, event.DeviceID, "❌ Gagal mengunduh gambar. Coba kirim ulang.")
			return err
		}
		
		// Simpan gambar ke file sistem
		imageFilename := fmt.Sprintf("wa_image_%d_%d.jpg", user.ID, time.Now().UnixNano())
		imagePath := filepath.Join("uploads", "images", imageFilename)
		_ = os.MkdirAll(filepath.Dir(imagePath), 0755)
		if err := os.WriteFile(imagePath, imgData, 0644); err == nil {
			savedImageURL = "/uploads/images/" + imageFilename
		} else {
			fmt.Printf("[WA][WARN] Gagal menyimpan gambar WA ke disk: %v\n", err)
		}

		imageBase64 = base64.StdEncoding.EncodeToString(imgData)
		if text == "" {
			text = "Tolong analisis gambar ini. Jika ini struk belanja, identifikasi item dan harganya."
		}
	}

	if audioTranscription != "" {
		prefix := "[TRANSKRIPSI SUARA - mungkin ada kesalahan fonetik, tolong koreksi]: "
		if text == "" {
			text = prefix + audioTranscription
		} else {
			text = text + "\n\n" + prefix + audioTranscription
		}
	}

	if text == "" {
		return nil
	}

	if err := s.chatHistSvc.SaveMessage(user.ID, "user", text, savedAudioURL, savedImageURL); err != nil {
		fmt.Printf("[WA][WARN] Gagal simpan pesan user: %v\n", err)
	}

	userContext := s.chatbotSvc.GetUserContext(user.ID, text)

	aiResp, err := s.aiSvc.Chat(text, imageBase64, userContext)
	if err != nil {
		fmt.Printf("[WA][ERROR] AI Chat gagal: %v\n", err)
		_ = s.sendWAMessage(msg.ChatID, event.DeviceID,
			"❌ AI tidak dapat merespons saat ini. Mohon coba beberapa saat lagi. 🙏")
		return err
	}

	replyText := aiResp.Reply

	if aiResp.IsTransaction && len(aiResp.Transactions) > 0 {
		saved, err := s.chatbotSvc.SaveTransactions(user.ID, aiResp.Transactions)
		if err != nil {
			fmt.Printf("[WA][ERROR] SaveTransactions gagal: %v\n", err)
			replyText += "\n\n⚠️ Transaksi terdeteksi tapi gagal disimpan."
		} else if len(saved) > 0 {
			summary := "\n\n✅ Transaksi berhasil dicatat!"
			for _, s := range saved {
				summary += fmt.Sprintf("\n📝 %s — Rp%.0f (%s) | 🏦 %s | 📂 %s",
					s.Description, s.Amount, s.Type, s.WalletName, s.CategoryName)
			}
			replyText += summary
		}
	}

	if err := s.chatHistSvc.SaveMessage(user.ID, "assistant", replyText, "", ""); err != nil {
		fmt.Printf("[WA][WARN] Gagal simpan balasan AI: %v\n", err)
	}
	return s.sendWAMessage(msg.ChatID, event.DeviceID, replyText)
}

func extractPhone(jid string) string {
	parts := strings.Split(jid, "@")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func (s *whatsAppService) downloadMediaFromGateway(mediaPath string) ([]byte, error) {
	url := s.waGatewayURL + "/" + strings.TrimPrefix(mediaPath, "/")
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat request: %w", err)
	}
	if s.waGatewayUser != "" {
		req.SetBasicAuth(s.waGatewayUser, s.waGatewayPass)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal mengunduh media: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wa-gateway mengembalikan status %d untuk %s", resp.StatusCode, mediaPath)
	}

	return io.ReadAll(resp.Body)
}

func (s *whatsAppService) transcribeAudio(data []byte, originalPath string) (string, error) {
	ext := ".ogg"
	lower := strings.ToLower(originalPath)
	for _, candidate := range []string{".ogg", ".mp4", ".webm", ".m4a", ".aac", ".wav"} {
		if strings.HasSuffix(lower, candidate) {
			ext = candidate
			break
		}
	}

	tmpPath := fmt.Sprintf("/tmp/wa_audio_%d%s", time.Now().UnixNano(), ext)
	if err := writeFile(tmpPath, data); err != nil {
		return "", fmt.Errorf("gagal tulis file temp: %w", err)
	}

	return s.aiSvc.ProcessVoice(tmpPath)
}

func (s *whatsAppService) sendWAMessage(chatID, deviceID, text string) error {
	payload := map[string]interface{}{
		"phone":   chatID,
		"message": text,
	}

	url := s.waGatewayURL + "/send/message"
	fmt.Printf("[WA][DEBUG] sendWAMessage -> chatID:%s, deviceID:%s\n", chatID, deviceID)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("gagal marshal payload: %w", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("gagal membuat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if s.waGatewayUser != "" {
		req.SetBasicAuth(s.waGatewayUser, s.waGatewayPass)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("gagal mengirim pesan WA: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("wa-gateway error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
