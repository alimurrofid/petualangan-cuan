package entity

type ChatAIResponse struct {
	Reply         string              `json:"reply"`
	IsTransaction bool                `json:"is_transaction"`
	Transactions  []TransactionItemAI `json:"transactions"`
}
type TransactionItemAI struct {
	Type         string  `json:"type"`
	Amount       float64 `json:"amount"`
	Description  string  `json:"description"`
	CategoryName string  `json:"category_name"`
	WalletName   string  `json:"wallet_name"`
}

type ChatResponse struct {
	Reply        string             `json:"reply"`
	Transactions []SavedTransaction `json:"transactions,omitempty"`
	AudioURL     string             `json:"audio_url,omitempty"`
	ImageURL     string             `json:"image_url,omitempty"`
}
type SavedTransaction struct {
	ID           uint    `json:"id"`
	Description  string  `json:"description"`
	Amount       float64 `json:"amount"`
	Type         string  `json:"type"`
	CategoryName string  `json:"category_name"`
	WalletName   string  `json:"wallet_name"`
}