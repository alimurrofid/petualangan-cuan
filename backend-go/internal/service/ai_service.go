package service

import (
	"bytes"
	"cuan-backend/internal/entity"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type AIService struct {
	ocrURL      string
	llmURL      string
	whisperPath string
	mu          sync.Mutex
}

func NewAIService(ocrURL, llmURL, whisperPath string) *AIService {
	return &AIService{
		ocrURL:      ocrURL,
		llmURL:      llmURL,
		whisperPath: whisperPath,
	}
}

func (s *AIService) ProcessVoice(path string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return "", fmt.Errorf("ffmpeg not found: %w", err)
	}

	// 2. Convert OGG/Opus to WAV (16kHz, mono) for Whisper using ffmpeg
	// Input: path (e.g. /app/uploads/voice.ogg)
	// Output: /app/uploads/voice.wav
	wavPath := strings.TrimSuffix(path, filepath.Ext(path)) + ".wav"
	
	// ffmpeg -i input.ogg -ar 16000 -ac 1 -c:a pcm_s16le output.wav
	convertCmd := exec.Command("ffmpeg", "-y", "-i", path, "-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le", wavPath)
	var convertOut bytes.Buffer
	convertCmd.Stdout = &convertOut
	convertCmd.Stderr = &convertOut

	if err := convertCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to convert audio: %w, output: %s", err, convertOut.String())
	}

	// Ensure cleanup of the WAV file usually, but keeping it for debugging might be okay for now.
	// For production, best to defer os.Remove(wavPath)
	defer os.Remove(wavPath)

	// 3. Run Whisper on the converted file
	cmd := exec.Command(s.whisperPath, "-m", "/app/bin/ggml-small.bin", "-f", wavPath, "-nt", "-l", "id")
	
	// Pisahkan Stdout (Hasil Text) dan Stderr (Log System)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Kalau error, baru kita butuh liat stderr nya
		return "", fmt.Errorf("failed to run whisper: %w, stderr: %s", err, stderr.String())
	}
	
	// Return hanya hasil text bersih
	return strings.TrimSpace(out.String()), nil
}

func (s *AIService) ScanReceipt(file io.Reader, filename string) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil { return "", err }
	_, err = io.Copy(part, file)
	if err != nil { return "", err }
	writer.Close()

	req, err := http.NewRequest("POST", s.ocrURL+"/scan", body)
	if err != nil { return "", err }
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil { return "", err }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ocr service returned status: %d", resp.StatusCode)
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Text, nil
}

// --- LOGIC UTAMA: LEBIH BERSIH KARENA PROMPT DIPISAH ---
func (s *AIService) ExtractFinancialData(rawText string) (entity.AIProcessResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// PANGGIL PROMPT DARI FILE prompt.go
	fullPrompt := fmt.Sprintf("<|im_start|>system\n%s<|im_end|>\n<|im_start|>user\n%s<|im_end|>\n<|im_start|>assistant\n", SystemPromptFinancialExtraction, rawText)

	payload := map[string]interface{}{
		"prompt":      fullPrompt,
		"n_predict":   256,
		"temperature": 0.05,
		"stop":        []string{"<|im_end|>"},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil { return nil, err }

	req, err := http.NewRequest("POST", s.llmURL+"/completion", bytes.NewBuffer(jsonPayload))
	if err != nil { return nil, err }
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("llm service returned status: %d", resp.StatusCode)
	}

	var llmResp struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
		return nil, err
	}

	// Log Raw Content
	fmt.Printf("[DEBUG] LLM Raw Response: %q\n", llmResp.Content)
	content := llmResp.Content

	startIdx := strings.Index(content, "[")
	endIdx := strings.LastIndex(content, "]")

	if startIdx == -1 || endIdx == -1 {
		if strings.Contains(content, "{") {
			content = "[" + content + "]"
			startIdx = strings.Index(content, "[")
			endIdx = strings.LastIndex(content, "]")
		} else {
			return nil, fmt.Errorf("AI output invalid: %s", content)
		}
	}

	jsonPart := content[startIdx : endIdx+1]

	var extraction entity.AIProcessResponse
	if err := json.Unmarshal([]byte(jsonPart), &extraction); err != nil {
		return nil, fmt.Errorf("failed to parse json: %v, raw: %s", err, jsonPart)
	}

	return extraction, nil
}