package service

import (
	"bufio"
	"bytes"
	"context"
	"cuan-backend/internal/entity"
	aiprovider "cuan-backend/internal/provider/ai"
	"os"
	"strconv"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type ChatParams struct {
	Message       string
	ImageBase64   string
	AudioBase64   string
	AudioFormat   string
	AudioMimeType string
	UserContext   string
}

type AIService interface {
	Chat(params ChatParams) (*entity.ChatAIResponse, error)
	ChatStream(params ChatParams, onToken func(string) error) (*entity.ChatAIResponse, error)
	ProcessVoice(path string) (string, error)
}

type aiService struct {
	provider   aiprovider.Provider
	whisperURL string
	llmSem     chan struct{}
}

func NewAIService(provider aiprovider.Provider, whisperURL string) AIService {
	maxConcurrent := 2 // default
	if val := os.Getenv("LLM_MAX_CONCURRENT_REQUESTS"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			maxConcurrent = parsed
		}
	}

	return &aiService{
		provider:   provider,
		whisperURL: whisperURL,
		llmSem:     make(chan struct{}, maxConcurrent),
	}
}

type chatMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type inputAudio struct {
	Data   string `json:"data"`
	Format string `json:"format"`
}

type contentPart struct {
	Type       string      `json:"type"`
	Text       string      `json:"text,omitempty"`
	ImageURL   *imageURL   `json:"image_url,omitempty"`
	InputAudio *inputAudio `json:"input_audio,omitempty"`
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

func (s *aiService) Chat(params ChatParams) (*entity.ChatAIResponse, error) {
	select {
	case s.llmSem <- struct{}{}:
		defer func() { <-s.llmSem }()
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("AI Server sedang sibuk, mohon coba beberapa saat lagi")
	}

	systemPrompt := fmt.Sprintf(SystemPromptChat, params.UserContext)
	log.Debug().Str("context", params.UserContext).Msg("User Context sent to LLM")

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	req := aiprovider.AIRequest{
		Prompt:        params.Message,
		Base64Image:   params.ImageBase64,
		Base64Audio:   params.AudioBase64,
		AudioFormat:   params.AudioFormat,
		AudioMimeType: params.AudioMimeType,
		System:        systemPrompt,
	}

	content, err := s.provider.GenerateCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("provider error: %w", err)
	}

	log.Debug().Str("response", content).Msg("AI Raw Response")

	return parseAIResponse(content), nil
}

func (s *aiService) ProcessVoice(path string) (string, error) {
	if s.whisperURL == "" {
		return "", fmt.Errorf("whisper service tidak dikonfigurasi")
	}
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

func (s *aiService) ChatStream(params ChatParams, onToken func(string) error) (*entity.ChatAIResponse, error) {
	select {
	case s.llmSem <- struct{}{}:
		defer func() { <-s.llmSem }()
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("AI Server sedang sibuk, mohon coba beberapa saat lagi")
	}

	systemPrompt := fmt.Sprintf(SystemPromptChat, params.UserContext)
	log.Debug().Str("context", params.UserContext).Msg("User Context sent to LLM (Stream)")

	payload := chatCompletionRequest{
		Messages:    buildMessages(systemPrompt, params),
		MaxTokens:   1024,
		Temperature: 0.3,
		Stream:      true,
	}

	return s.doChatStream(payload, onToken)
}

func buildMessages(system string, params ChatParams) []chatMessage {
	var userContent interface{}
	if params.ImageBase64 != "" || params.AudioBase64 != "" {
		parts := []contentPart{}
		if params.Message != "" {
			parts = append(parts, contentPart{Type: "text", Text: params.Message})
		}
		if params.ImageBase64 != "" {
			parts = append(parts, contentPart{
				Type:     "image_url",
				ImageURL: &imageURL{URL: "data:image/jpeg;base64," + params.ImageBase64},
			})
		}
		if params.AudioBase64 != "" {
			format := params.AudioFormat
			if format != "wav" && format != "mp3" {
				format = "wav"
			}
			parts = append(parts, contentPart{
				Type: "input_audio",
				InputAudio: &inputAudio{
					Data:   params.AudioBase64,
					Format: format,
				},
			})
		}
		userContent = parts
	} else {
		userContent = params.Message
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
	var prompt, system, imageBase64, audioBase64, audioFormat string
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
					} else if p.Type == "input_audio" && p.InputAudio != nil {
						audioBase64 = p.InputAudio.Data
						audioFormat = p.InputAudio.Format
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
		Base64Audio: audioBase64,
		AudioFormat: audioFormat,
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