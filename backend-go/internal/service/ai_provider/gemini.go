package ai_provider

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GeminiProvider struct {
	APIKey string
	Model  string
}

func NewGeminiProvider(apiKey, model string) *GeminiProvider {
	if model == "" {
		model = "gemini-2.0-flash" 
	}
	return &GeminiProvider{
		APIKey: apiKey,
		Model:  model,
	}
}

type geminiPart struct {
	Text       string      `json:"text,omitempty"`
	InlineData *inlineData `json:"inline_data,omitempty"`
}

type inlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	Contents          []geminiContent    `json:"contents"`
	SystemInstruction *geminiContent     `json:"system_instruction,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func (p *GeminiProvider) GenerateResponse(message string, fileData []byte, mimeType string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", p.Model, p.APIKey)

	parts := []geminiPart{}
	
	if len(fileData) > 0 {
		parts = append(parts, geminiPart{
			InlineData: &inlineData{
				MimeType: mimeType,
				Data:     base64.StdEncoding.EncodeToString(fileData),
			},
		})
	}

	parts = append(parts, geminiPart{Text: message})

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{Parts: parts},
		},
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{
				{Text: SystemPrompt},
			},
		},
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
		if resp.StatusCode == 429 {
			return "", fmt.Errorf("quota exceeded (429). please try again later or switch models")
		}
		return "", fmt.Errorf("gemini api error (url: %s): %s, %s", url, resp.Status, string(bodyBytes))
	}

	var geminiResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", err
	}

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		return geminiResp.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("no response candidate found")
}
