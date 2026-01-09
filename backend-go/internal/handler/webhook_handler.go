package handler

import (
	"bytes"
	"cuan-backend/internal/entity"
	"cuan-backend/internal/service"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// --- Memory Cache (Anti-Duplikasi) ---
var (
	processedMsg sync.Map
)

// Goroutine pembersih cache
func init() {
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			processedMsg.Range(func(key, value interface{}) bool {
				if time.Since(value.(time.Time)) > 5*time.Minute {
					processedMsg.Delete(key)
				}
				return true
			})
		}
	}()
}

type WebhookHandler struct {
	aiService          *service.AIService
	transactionService service.TransactionService
	userService        service.UserService
}

func NewWebhookHandler(aiSvc *service.AIService, txSvc service.TransactionService, userSvc service.UserService) *WebhookHandler {
	return &WebhookHandler{
		aiService:          aiSvc,
		transactionService: txSvc,
		userService:        userSvc,
	}
}

type WAWebhookRequest struct {
	Event   string                 `json:"event"`
	Payload map[string]interface{} `json:"payload"`
}

// --- FUNGSI KIRIM BALASAN KE WA ---
func (h *WebhookHandler) ReplyToUser(toJID, message, replyID string) error {
	baseURL := os.Getenv("WA_URL")
	authUser := os.Getenv("WA_GATEWAY_USERNAME")
	authPass := os.Getenv("WA_GATEWAY_PASSWORD")

	if baseURL == "" {
		return fmt.Errorf("WA_URL environment variable is trying to be accessed but is not set")
	}

	url := fmt.Sprintf("%s/send/message", strings.TrimRight(baseURL, "/"))

	payload := map[string]interface{}{
		"phone":            toJID,
		"message":          message,
		"reply_message_id": replyID,
	}

	requestBody, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil { return err }
	
	req.Header.Set("Content-Type", "application/json")
	if authUser != "" {
		req.SetBasicAuth(authUser, authPass)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gateway error: %d - %s", resp.StatusCode, string(body))
	}
	return nil
}

// --- HELPER: RESOLVE WALLET ID ---
func (h *WebhookHandler) resolveWalletID(name string) (uint, string) {
	name = strings.ToLower(name)
	// Mapping manual sesuai ID di Database
	if strings.Contains(name, "bca") { return 2, "Bank BCA" }
	if strings.Contains(name, "gopay") { return 3, "GoPay" }
	if strings.Contains(name, "ovo") { return 4, "OVO" }
	// Default
	return 1, "Dompet Tunai"
}

// --- HELPER: FORMAT PESAN BALASAN ---
func (h *WebhookHandler) buildSuccessMessage(items entity.AIProcessResponse) string {
	dateStr := time.Now().Format("02 Jan 2006")
	p := message.NewPrinter(language.Indonesian)
	
	totalAmount := 0
	var detailLines []string

	for _, item := range items {
		totalAmount += item.Amount
		
		// Resolusi nama wallet agar tampil cantik di chat
		_, walletDisplayName := h.resolveWalletID(item.Wallet)

		desc := item.Item
		if item.Note != "" {
			desc = fmt.Sprintf("%s (%s)", item.Item, item.Note)
		}

		line := fmt.Sprintf("• %s\n   └ 💸 *Rp %s* via %s", desc, p.Sprintf("%d", item.Amount), walletDisplayName)
		detailLines = append(detailLines, line)
	}

	dashboardURL := os.Getenv("FRONTEND_URL")
	if dashboardURL == "" { dashboardURL = "(Setting di .env)" }
	dashboardURL = strings.TrimSpace(strings.TrimRight(dashboardURL, "/"))

	msg := fmt.Sprintf(`✅ *%d Transaksi Berhasil!*

💰 Total: *Rp %s*

🧾 Rincian:
%s

📅 Tanggal: %s
📊 Cek: %s/transactions`, 
		len(items),
		p.Sprintf("%d", totalAmount),
		strings.Join(detailLines, "\n"),
		dateStr,
		dashboardURL,
	)

	return msg
}

// --- HANDLER UTAMA ---
func (h *WebhookHandler) HandleWhatsAppWebhook(c *fiber.Ctx) error {
	// [STEP 0] Terima Request
	var req WAWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		fmt.Println("❌ [STEP 0] Error parsing body:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}
	p := req.Payload

	msgID, _ := p["id"].(string)
	
	// Ignore status updates / acks
	if req.Event != "message" {
		return c.Status(200).JSON(fiber.Map{"status": "ignored"})
	}

	// 1. Anti Duplikasi
	if _, exists := processedMsg.Load(msgID); exists && msgID != "" {
		return c.Status(200).JSON(fiber.Map{"status": "duplicate_ignored"})
	}
	if msgID != "" {
		processedMsg.Store(msgID, time.Now())
	}

	// GO ROUTINE: Proses Asynchronous agar tidak timeout
	go func() {
		h.processMessage(req, msgID)
	}()

	// Return 200 OK segera agar WA tidak retry/timeout
	return c.JSON(fiber.Map{"status": "processing_started"})
}

func (h *WebhookHandler) processMessage(req WAWebhookRequest, msgID string) {
	p := req.Payload
	from, _ := p["from"].(string) 
	fromName, _ := p["from_name"].(string)
	body, _ := p["body"].(string)

	// Detect Media
	var mediaURL string
	var mediaType string
	if audioInfo, ok := p["audio"].(map[string]interface{}); ok {
		mediaType = "audio"
		if path, ok := audioInfo["media_path"].(string); ok {
			baseURL := os.Getenv("WA_URL")
			if baseURL != "" {
				baseURL = strings.TrimRight(baseURL, "/")
				path = strings.TrimLeft(path, "/")
				parts := strings.Split(path, "/")
				for i, part := range parts { parts[i] = url.PathEscape(part) }
				encodedPath := strings.Join(parts, "/")
				mediaURL = fmt.Sprintf("%s/%s", baseURL, encodedPath)
			}
		}
	} else if audioPath, ok := p["audio"].(string); ok { 
		mediaType = "audio"
		mediaURL = audioPath 
	}

	// Detect Image
	if imageInfo, ok := p["image"].(map[string]interface{}); ok {
		mediaType = "image"
		// Default structure often has 'url' or 'id', but here we follow the pattern used in audio if possible
		// However, WA Webhook usually sends 'id' for media.
		// If using the local gateway as seen in audio logic:
		if path, ok := imageInfo["media_path"].(string); ok {
			baseURL := os.Getenv("WA_URL")
			if baseURL != "" {
				baseURL = strings.TrimRight(baseURL, "/")
				path = strings.TrimLeft(path, "/")
				parts := strings.Split(path, "/")
				for i, part := range parts { parts[i] = url.PathEscape(part) }
				encodedPath := strings.Join(parts, "/")
				mediaURL = fmt.Sprintf("%s/%s", baseURL, encodedPath)
			}
		} else if u, ok := imageInfo["url"].(string); ok {
			mediaURL = u
		}
	}

	if body == "" && mediaURL == "" {
		fmt.Println("⚠️ Empty body and media, ignoring.")
		return 
	}

	// [STEP 1] Log Incoming
	logMsg := body
	if mediaURL != "" { logMsg = fmt.Sprintf("[%s Media]", strings.ToUpper(mediaType)) }
	fmt.Printf("\n[STEP 1/6] 📩 Webhook: %s (%s) | Isi: %s\n", fromName, from, logMsg)

	userID := uint(1)
	var extracted entity.AIProcessResponse
	var err error

	// [STEP 2] Proses Media / Text
	isVoice := mediaType == "audio" && mediaURL != ""
	isImage := mediaType == "image" && mediaURL != ""

	if isVoice {
		fmt.Printf("[STEP 2/6] 🎤 Downloading & Processing Voice...\n")
		
		audioPath, err := h.downloadMedia(mediaURL, "ogg")
		if err != nil {
			fmt.Printf("❌ Download failed: %v\n", err)
			h.ReplyToUser(from, "❌ Gagal download audio", msgID)
			return
		}
		defer os.Remove(audioPath)

		// [STEP 3] Transkripsi
		text, err := h.aiService.ProcessVoice(audioPath)
		if err != nil {
			fmt.Printf("❌ Whisper failed: %v\n", err)
			h.ReplyToUser(from, "❌ Gagal transkrip", msgID)
			return
		}
		fmt.Printf("[STEP 3/6] 📝 Transcript: %q\n", text)
		
		extracted, err = h.aiService.ExtractFinancialData(text)

	} else if isImage {
		fmt.Printf("[STEP 2/6] 📸 Downloading & Processing Image (OCR)...\n")
		
		imagePath, err := h.downloadMedia(mediaURL, "jpg")
		if err != nil {
			fmt.Printf("❌ Download failed: %v\n", err)
			h.ReplyToUser(from, "❌ Gagal download gambar", msgID)
			return
		}
		defer os.Remove(imagePath)

		// [STEP 3] OCR
		f, err := os.Open(imagePath)
		if err != nil {
			fmt.Printf("❌ Failed to open image: %v\n", err)
			return 
		}
		
		ocrText, err := h.aiService.ScanReceipt(f, filepath.Base(imagePath))
		f.Close() // Close immediately after use

		if err != nil {
			fmt.Printf("❌ OCR failed: %v\n", err)
			h.ReplyToUser(from, "❌ Gagal baca gambar (OCR)", msgID)
			return
		}
		fmt.Printf("[STEP 3/6] 🧾 OCR Text: %q\n", ocrText)

		// [STEP 4] Ekstraksi AI
		extracted, err = h.aiService.ExtractFinancialData(ocrText)

	} else {
		fmt.Printf("[STEP 2/6] 🤖 Processing Text...\n")
		// Step 3 skipped for text
		// [STEP 4] Ekstraksi
		extracted, err = h.aiService.ExtractFinancialData(body)
	}

	if err != nil {
		fmt.Printf("❌ AI Extraction Error: %v\n", err)
		h.ReplyToUser(from, "⚠️ Gagal baca format. Coba: 'Beli [item] [harga]'", msgID)
		return
	}

	// Log Raw JSON for Debugging
	jsonDebug, _ := json.Marshal(extracted)
	fmt.Printf("[STEP 4/6] 🧠 AI Result: %s\n", string(jsonDebug))

	// [STEP 5] Simpan DB
	successCount := 0
	for _, item := range extracted {
		categoryID := uint(1)
		cat := strings.ToLower(item.Category)
		if strings.Contains(cat, "makan") { categoryID = 2 }
		if strings.Contains(cat, "transport") { categoryID = 3 }
		
		walletID, _ := h.resolveWalletID(item.Wallet)

		finalDesc := item.Item
		if item.Note != "" {
			finalDesc = fmt.Sprintf("%s (%s)", item.Item, item.Note)
		}

		input := service.CreateTransactionInput{
			WalletID:    walletID,
			CategoryID:  categoryID,
			Amount:      float64(item.Amount),
			Type:        "expense",
			Description: finalDesc,
			Date:        time.Now(),
		}

		_, err = h.transactionService.CreateTransaction(userID, input)
		if err == nil { successCount++ }
	}
	fmt.Printf("[STEP 5/6] 💾 Saved %d/%d transactions to DB.\n", successCount, len(extracted))

	// [STEP 6] Kirim Balasan
	replyMessage := h.buildSuccessMessage(extracted)
	err = h.ReplyToUser(from, replyMessage, msgID)
	if err != nil {
		fmt.Printf("❌ Failed to send reply: %v\n", err)
	} else {
		fmt.Printf("[STEP 6/6] ✅ Reply sent to User.\n")
	}
}

// Renamed to generic downloadMedia
func (h *WebhookHandler) downloadMedia(url, ext string) (string, error) {
    // Basic validation
    if url == "" {
        return "", fmt.Errorf("empty media url")
    }

    resp, err := http.Get(url)
    if err != nil { return "", err }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return "", fmt.Errorf("failed to download: status %d", resp.StatusCode)
    }

    tmpObj, err := os.CreateTemp("", fmt.Sprintf("wa_media_*.%s", ext))
    if err != nil { return "", err }
    defer tmpObj.Close()

    _, err = io.Copy(tmpObj, resp.Body)
    if err != nil { return "", err }
    
    return tmpObj.Name(), nil
}