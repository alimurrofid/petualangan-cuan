package service

import (
	"bufio"
	"bytes"
	"cuan-backend/internal/entity"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type AIService struct {
	llmURL     string
	whisperURL string
	llmSem     chan struct{}
}

func NewAIService(llmURL, whisperURL string) *AIService {
	return &AIService{
		llmURL:     llmURL,
		whisperURL: whisperURL,
		llmSem:     make(chan struct{}, 2), // max 2 concurrent LLM inference
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
	select {
	case s.llmSem <- struct{}{}:
		defer func() { <-s.llmSem }()
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("AI Server sedang sibuk, mohon coba beberapa saat lagi")
	}

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
	fmt.Printf("[DEBUG] User Context sent to LLM:\n%s\n", userContext)

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
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	pr, pw := io.Pipe()
	m := multipart.NewWriter(pw)

	go func() {
		defer pw.Close()
		defer m.Close()
		fw, err := m.CreateFormFile("file", filepath.Base(path))
		if err == nil {
			io.Copy(fw, file)
		}
		m.WriteField("model", "small")
	}()

	req, err := http.NewRequest("POST", s.whisperURL+"/v1/audio/transcriptions", pr)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", m.FormDataContentType())

	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("whisper request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("whisper returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode whisper response: %w", err)
	}

	return strings.TrimSpace(result.Text), nil
}

func (s *AIService) ChatStream(message string, imageBase64 string, userContext string, onToken func(string) error) (*entity.ChatAIResponse, error) {
	select {
	case s.llmSem <- struct{}{}:
		defer func() { <-s.llmSem }()
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("AI Server sedang sibuk, mohon coba beberapa saat lagi")
	}

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
	fmt.Printf("[DEBUG] User Context sent to LLM (Stream):\n%s\n", userContext)

	payload := chatCompletionRequest{
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userContent},
		},
		MaxTokens:   1024,
		Temperature: 0.3,
		Stream:      true,
	}

	return s.doChatStream(payload, onToken)
}

func (s *AIService) doChatStream(payload chatCompletionRequest, onToken func(string) error) (*entity.ChatAIResponse, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", s.llmURL+"/v1/chat/completions", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llm stream request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("llm returned status %d: %s", resp.StatusCode, string(body))
	}

	reader := bufio.NewReader(resp.Body)
	var fullContent strings.Builder

	var (
		buffer       strings.Builder
		inString     bool
		isReplyField bool
		escape       bool
		jsonDepth    int
	)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error reading stream: %w", err)
		}

		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}

		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			if content != "" {
				fullContent.WriteString(content)

				for _, r := range content {
					if !inString {
						if r == '{' {
							jsonDepth++
						} else if r == '}' {
							jsonDepth--
						}
					}

					buffer.WriteRune(r)
					bufStr := buffer.String()

					if !inString && !isReplyField {
						if strings.HasSuffix(bufStr, `"reply": "`) || strings.HasSuffix(bufStr, `"reply":"`) {
							isReplyField = true
							inString = true
							buffer.Reset()
						}
					} else if isReplyField && inString {
						if escape {
							escape = false
							if err := onToken(string(r)); err != nil {
								return nil, err
							}
						} else {
							if r == '\\' {
								escape = true
								if err := onToken(string(r)); err != nil {
								    return nil, err
								}
							} else if r == '"' {
								inString = false
								isReplyField = false
							} else {
								if err := onToken(string(r)); err != nil {
									return nil, err
								}
							}
						}
					}
				}
				if buffer.Len() > 20 && !isReplyField {
					temp := buffer.String()
					buffer.Reset()
					buffer.WriteString(temp[len(temp)-20:])
				}
			}
		}
	}

	finalContent := fullContent.String()
	fmt.Printf("[DEBUG] AI Stream Final Content: %q\n", finalContent)

	startIdx := strings.Index(finalContent, "{")
	endIdx := strings.LastIndex(finalContent, "}")
	
	response := &entity.ChatAIResponse{
		Reply:         finalContent,
		IsTransaction: false,
		Transactions:  nil,
	}

	if startIdx != -1 && endIdx != -1 {
		jsonPart := finalContent[startIdx : endIdx+1]
		var parsed entity.ChatAIResponse
		if err := json.Unmarshal([]byte(jsonPart), &parsed); err == nil {
			response.Reply = parsed.Reply
			response.IsTransaction = parsed.IsTransaction
			response.Transactions = parsed.Transactions
		}
	}

	return response, nil
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