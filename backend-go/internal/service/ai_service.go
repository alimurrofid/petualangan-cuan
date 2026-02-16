package service

import (
	"bytes"
	"cuan-backend/internal/entity"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type AIService struct {
	llmURL      string
	whisperPath string
	mu          sync.Mutex
}

func NewAIService(llmURL, whisperPath string) *AIService {
	return &AIService{
		llmURL:      llmURL,
		whisperPath: whisperPath,
	}
}

type chatMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type contentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *imageURL `json:"image_url,omitempty"`
}

type imageURL struct {
	URL string `json:"url"`
}

type chatCompletionRequest struct {
	Model       string        `json:"model,omitempty"`
	Messages    []chatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	Stream      bool          `json:"stream"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (s *AIService) Chat(message string, imageBase64 string, userContext string) (*entity.ChatAIResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var userContent interface{}

	if imageBase64 != "" {
		parts := []contentPart{
			{Type: "text", Text: message},
			{Type: "image_url", ImageURL: &imageURL{URL: "data:image/jpeg;base64," + imageBase64}},
		}
		userContent = parts
	} else {
		userContent = message
	}

	systemPrompt := fmt.Sprintf(SystemPromptChat, userContext)

	payload := chatCompletionRequest{
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userContent},
		},
		MaxTokens:   1024,
		Temperature: 0.3,
		Stream:      false,
	}

	content, err := s.doChatUnlocked(payload)
	if err != nil {
		return nil, err
	}

	fmt.Printf("[DEBUG] AI Raw Response: %q\n", content)

	startIdx := strings.Index(content, "{")
	endIdx := strings.LastIndex(content, "}")
	if startIdx == -1 || endIdx == -1 {
		return &entity.ChatAIResponse{
			Reply:         content,
			IsTransaction: false,
			Transactions:  nil,
		}, nil
	}

	jsonPart := content[startIdx : endIdx+1]

	var response entity.ChatAIResponse
	if err := json.Unmarshal([]byte(jsonPart), &response); err != nil {
		fmt.Printf("[WARN] Failed to parse AI JSON: %v, raw: %s\n", err, jsonPart)
		return &entity.ChatAIResponse{
			Reply:         content,
			IsTransaction: false,
			Transactions:  nil,
		}, nil
	}

	return &response, nil
}

func (s *AIService) ProcessVoice(path string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return "", fmt.Errorf("ffmpeg not found: %w", err)
	}

	wavPath := strings.TrimSuffix(path, filepath.Ext(path)) + ".wav"

	convertCmd := exec.Command("ffmpeg", "-y", "-i", path, "-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le", wavPath)
	var convertOut bytes.Buffer
	convertCmd.Stdout = &convertOut
	convertCmd.Stderr = &convertOut

	if err := convertCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to convert audio: %w, output: %s", err, convertOut.String())
	}

	defer os.Remove(wavPath)

	cmd := exec.Command(s.whisperPath, "-m", "/app/bin/ggml-small.bin", "-f", wavPath, "-nt", "-l", "id")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run whisper: %w, stderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(out.String()), nil
}


func (s *AIService) doChatUnlocked(payload chatCompletionRequest) (string, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", s.llmURL+"/v1/chat/completions", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("llm returned status %d: %s", resp.StatusCode, string(body))
	}

	var result chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode llm response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("llm returned no choices")
	}

	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}