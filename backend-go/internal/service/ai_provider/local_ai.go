package ai_provider

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type LocalAIProvider struct {
	BaseURL string
	Model   string
}

func NewLocalAIProvider(baseURL, model string) *LocalAIProvider {
	if baseURL == "" {
		baseURL = "http://local-ai:8080/v1"
	}
	if model == "" {
		model = "phi-4-mini-instruct"
	}
	return &LocalAIProvider{
		BaseURL: baseURL,
		Model:   model,
	}
}

// OpenAI Chat Completion Request Structures
type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type openAIMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // string or []openAIContentPart
}

type openAIContentPart struct {
	Type       string             `json:"type"`
	Text       string             `json:"text,omitempty"`
	ImageURL   *openAIImageURL    `json:"image_url,omitempty"`
	InputAudio *openAIInputAudio  `json:"input_audio,omitempty"`
}

type openAIImageURL struct {
	URL string `json:"url"`
}

type openAIInputAudio struct {
	Data   string `json:"data"`
	Format string `json:"format"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (p *LocalAIProvider) GenerateResponse(message string, fileData []byte, mimeType string) (string, error) {
	url := fmt.Sprintf("%s/chat/completions", p.BaseURL)

	contentParts := []openAIContentPart{}

	// Add Text Part
	if message != "" {
		contentParts = append(contentParts, openAIContentPart{
			Type: "text",
			Text: message,
		})
	}

	// Add Media Part (Image or Audio)
	if len(fileData) > 0 {
		base64Data := base64.StdEncoding.EncodeToString(fileData)

		if mimeType == "audio/wav" || mimeType == "audio/mp3" || mimeType == "audio/ogg" {
			// Map mime to format. OpenAI usually expects "wav" or "mp3".
			// LocalAI might be flexible, but let's default to parsing simple extension or raw.
			// Assuming format "wav" for robustness if mime is audio/wav
			format := "wav"
			if mimeType == "audio/mp3" {
				format = "mp3"
			}
			
			contentParts = append(contentParts, openAIContentPart{
				Type: "input_audio",
				InputAudio: &openAIInputAudio{
					Data:   base64Data,
					Format: format, 
				},
			})
		} else {
			// Treat as Image
			dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)
			contentParts = append(contentParts, openAIContentPart{
				Type: "image_url",
				ImageURL: &openAIImageURL{
					URL: dataURI,
				},
			})
		}
	}

	// Construct system message (using the global SystemPrompt from constants.go)
	messages := []openAIMessage{
		{
			Role:    "system",
			Content: SystemPrompt,
		},
		{
			Role:    "user",
			Content: contentParts,
		},
	}

	reqBody := openAIRequest{
		Model:    p.Model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("localai api error (url: %s): %s, %s", url, resp.Status, string(bodyBytes))
	}

	var openAIResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return "", err
	}

	if len(openAIResp.Choices) > 0 {
		return openAIResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response choice found")
}
