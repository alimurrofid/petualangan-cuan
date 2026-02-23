package service

import (
	"bufio"
	"bytes"
	"context"
	"cuan-backend/internal/entity"
	aiprovider "cuan-backend/internal/provider/ai"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type AIService interface {
	Chat(message string, imageBase64 string, userContext string) (*entity.ChatAIResponse, error)
	ChatStream(message string, imageBase64 string, userContext string, onToken func(string) error) (*entity.ChatAIResponse, error)
	ProcessVoice(path string) (string, error)
}

type aiService struct {
	provider   aiprovider.Provider
	whisperURL string
	llmSem     chan struct{}
}

func NewAIService(provider aiprovider.Provider, whisperURL string) AIService {
	return &aiService{
		provider:   provider,
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

func (s *aiService) Chat(message string, imageBase64 string, userContext string) (*entity.ChatAIResponse, error) {
	select {
	case s.llmSem <- struct{}{}:
		defer func() { <-s.llmSem }()
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("AI Server sedang sibuk, mohon coba beberapa saat lagi")
	}

	systemPrompt := fmt.Sprintf(SystemPromptChat, userContext)
	log.Debug().Str("context", userContext).Msg("User Context sent to LLM")

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	req := aiprovider.AIRequest{
		Prompt:      message,
		Base64Image: imageBase64,
		System:      systemPrompt,
	}

	content, err := s.provider.GenerateCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("provider error: %w", err)
	}

	log.Debug().Str("response", content).Msg("AI Raw Response")

	return parseAIResponse(content), nil
}

func (s *aiService) ProcessVoice(path string) (string, error) {
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

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", s.whisperURL+"/v1/audio/transcriptions", pr)
	if err != nil {
		return "", fmt.Errorf("failed to create whisper request: %w", err)
	}
	req.Header.Set("Content-Type", m.FormDataContentType())

	client := &http.Client{}
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

func (s *aiService) ChatStream(message string, imageBase64 string, userContext string, onToken func(string) error) (*entity.ChatAIResponse, error) {
	select {
	case s.llmSem <- struct{}{}:
		defer func() { <-s.llmSem }()
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("AI Server sedang sibuk, mohon coba beberapa saat lagi")
	}

	systemPrompt := fmt.Sprintf(SystemPromptChat, userContext)
	log.Debug().Str("context", userContext).Msg("User Context sent to LLM (Stream)")

	payload := chatCompletionRequest{
		Messages:    buildMessages(systemPrompt, message, imageBase64),
		MaxTokens:   1024,
		Temperature: 0.3,
		Stream:      true,
	}

	return s.doChatStream(payload, onToken)
}

func buildMessages(system, prompt, imageBase64 string) []chatMessage {
	var userContent interface{}
	if imageBase64 != "" {
		userContent = []contentPart{
			{Type: "text", Text: prompt},
			{Type: "image_url", ImageURL: &imageURL{URL: "data:image/jpeg;base64," + imageBase64}},
		}
	} else {
		userContent = prompt
	}

	return []chatMessage{
		{Role: "system", Content: system},
		{Role: "user", Content: userContent},
	}
}

func parseAIResponse(content string) *entity.ChatAIResponse {
	startIdx := strings.Index(content, "{")
	endIdx := strings.LastIndex(content, "}")
	if startIdx == -1 || endIdx == -1 {
		return &entity.ChatAIResponse{Reply: content, IsTransaction: false}
	}

	jsonPart := content[startIdx : endIdx+1]
	var response entity.ChatAIResponse
	if err := json.Unmarshal([]byte(jsonPart), &response); err != nil {
		log.Warn().Err(err).Str("raw", jsonPart).Msg("Failed to parse AI JSON")
		return &entity.ChatAIResponse{Reply: content, IsTransaction: false}
	}
	return &response
}

func (s *aiService) doChatStream(payload chatCompletionRequest, onToken func(string) error) (*entity.ChatAIResponse, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	type urlGetter interface {
		GetURL() string
	}

	var streamURL string
	if ug, ok := s.provider.(urlGetter); ok {
		streamURL = ug.GetURL() + "/v1/chat/completions"
	} else {
		return s.chatStreamViaProvider(payload, onToken)
	}

	req, err := http.NewRequest("POST", streamURL, bytes.NewBuffer(jsonPayload))
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

	return parseSSEStream(resp.Body, onToken)
}

func (s *aiService) chatStreamViaProvider(payload chatCompletionRequest, onToken func(string) error) (*entity.ChatAIResponse, error) {
	var prompt, system, imageBase64 string
	for _, msg := range payload.Messages {
		switch msg.Role {
		case "system":
			if s, ok := msg.Content.(string); ok {
				system = s
			}
		case "user":
			if s, ok := msg.Content.(string); ok {
				prompt = s
			} else if parts, ok := msg.Content.([]contentPart); ok {
				for _, p := range parts {
					if p.Type == "text" {
						prompt = p.Text
					} else if p.Type == "image_url" && p.ImageURL != nil {
						imageBase64 = strings.TrimPrefix(p.ImageURL.URL, "data:image/jpeg;base64,")
					}
				}
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	req := aiprovider.AIRequest{
		Prompt:      prompt,
		Base64Image: imageBase64,
		System:      system,
	}

	content, err := s.provider.GenerateCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("provider error: %w", err)
	}

	log.Debug().Str("content", content).Msg("AI Stream (via provider) Final Content")

	response := parseAIResponse(content)
	for _, char := range response.Reply {
		if err := onToken(string(char)); err != nil {
			return nil, err
		}
	}

	return response, nil
}

func parseSSEStream(body io.Reader, onToken func(string) error) (*entity.ChatAIResponse, error) {
	reader := bufio.NewReader(body)
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
	log.Debug().Str("content", finalContent).Msg("AI Stream Final Content")

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