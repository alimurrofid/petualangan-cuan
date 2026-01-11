package ai_provider

// AIProvider defines the contract for different AI backends
type AIProvider interface {
	GenerateResponse(message string, fileData []byte, mimeType string) (string, error)
}
