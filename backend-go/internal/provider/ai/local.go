package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type LocalProvider struct {
	url string
}

func NewLocalProvider(url string) *LocalProvider {
	return &LocalProvider{url: url}
}

func (p *LocalProvider) GetURL() string {
	return p.url
}

func (p *LocalProvider) GenerateCompletion(ctx context.Context, req AIRequest) (string, error) {
	reqID, _ := ctx.Value("request_id").(string)
	log.Info().Str("request_id", reqID).Str("url", p.url).Msg("[Local AI] Sending request")

	payload := buildChatPayload(req)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("[Local AI] failed to marshal payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		p.url+"/v1/chat/completions",
		bytes.NewBuffer(jsonPayload),
	)
	if err != nil {
		return "", fmt.Errorf("[Local AI] failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("[Local AI] request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("[Local AI] returned status %d: %s", resp.StatusCode, string(body))
	}

	return decodeCompletion(resp.Body, "[Local AI]")
}

type localChatMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type localContentPart struct {
	Type     string        `json:"type"`
	Text     string        `json:"text,omitempty"`
	ImageURL *localImgURL  `json:"image_url,omitempty"`
}

type localImgURL struct {
	URL string `json:"url"`
}

type completionPayload struct {
	Model       string             `json:"model,omitempty"`
	Messages    []localChatMessage `json:"messages"`
	MaxTokens   int                `json:"max_tokens,omitempty"`
	Temperature float64            `json:"temperature,omitempty"`
	Stream      bool               `json:"stream"`
}

type completionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func buildChatPayload(req AIRequest) completionPayload {
	var userContent interface{}

	if req.Base64Image != "" {
		userContent = []localContentPart{
			{Type: "text", Text: req.Prompt},
			{Type: "image_url", ImageURL: &localImgURL{URL: "data:image/jpeg;base64," + req.Base64Image}},
		}
	} else {
		userContent = req.Prompt
	}

	messages := []localChatMessage{}
	if req.System != "" {
		messages = append(messages, localChatMessage{Role: "system", Content: req.System})
	}
	messages = append(messages, localChatMessage{Role: "user", Content: userContent})

	return completionPayload{
		Messages:    messages,
		MaxTokens:   1024,
		Temperature: 0.3,
		Stream:      false,
	}
}

func decodeCompletion(body io.Reader, logPrefix string) (string, error) {
	var result completionResponse
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return "", fmt.Errorf("%s failed to decode response: %w", logPrefix, err)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("%s returned no choices", logPrefix)
	}
	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}
