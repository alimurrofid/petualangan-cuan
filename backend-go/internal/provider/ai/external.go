package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type ExternalProvider struct {
	url    string
	apiKey string
	model  string
}

func NewExternalProvider(url, apiKey, model string) *ExternalProvider {
	return &ExternalProvider{url: url, apiKey: apiKey, model: model}
}

func (p *ExternalProvider) GenerateCompletion(ctx context.Context, req AIRequest) (string, error) {
	log.Printf("[External AI] Sending request to %s (model: %s)", p.url, p.model)

	payload := buildExternalPayload(req)
	payload.Model = p.model
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("[External AI] failed to marshal payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		p.url,
		bytes.NewBuffer(jsonPayload),
	)
	if err != nil {
		return "", fmt.Errorf("[External AI] failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("[External AI] request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("[External AI] returned status %d: %s", resp.StatusCode, string(body))
	}

	return decodeCompletion(resp.Body, "[External AI]")
}

func buildExternalPayload(req AIRequest) completionPayload {
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
		messages = append(messages, localChatMessage{
			Role:    "user",
			Content: "[INSTRUKSI SISTEM - IKUTI SELALU]:\n" + req.System,
		})
		messages = append(messages, localChatMessage{
			Role:    "assistant",
			Content: "Baik, saya mengerti dan akan mengikuti instruksi tersebut.",
		})
	}
	messages = append(messages, localChatMessage{Role: "user", Content: userContent})

	return completionPayload{
		Messages:    messages,
		MaxTokens:   1024,
		Temperature: 0.3,
		Stream:      false,
	}
}

