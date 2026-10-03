package entity

type ChatAIResponse struct {
	Reply         string              `json:"reply"`
	IsTransaction bool                `json:"is_transaction"`
	Transactions  []TransactionItemAI `json:"transactions"`
}
type TransactionItemAI struct {
	Action       string  `json:"action"`                  // create_transaction, update_transaction, delete_transaction, transfer, pay_debt, save_goal, create_wishlist...
	ID           uint    `json:"id"`                      // target ID if action is update/delete/pay_debt/save_goal
	Type         string  `json:"type"`
	Amount       float64 `json:"amount"`
	Description  string  `json:"description"`
	CategoryName string  `json:"category_name"`
	WalletName   string  `json:"wallet_name"`
	ToWalletName string  `json:"to_wallet_name,omitempty"` // target wallet for transfer
	Priority     string  `json:"priority,omitempty"`       // low, medium, high for wishlist
}

type ChatResponse struct {
	Reply        string             `json:"reply"`
	Transactions []SavedTransaction `json:"transactions,omitempty"`
	AudioURL     string             `json:"audio_url,omitempty"`
	ImageURL     string             `json:"image_url,omitempty"`
}
type SavedTransaction struct {
	ID           uint    `json:"id"`
	Action       string  `json:"action"`                  // create_transaction, update_transaction, delete_transaction, transfer, pay_debt, save_goal, create_wishlist...
	Description  string  `json:"description"`
	Amount       float64 `json:"amount"`
	Type         string  `json:"type"`
	CategoryName string  `json:"category_name"`
	WalletName   string  `json:"wallet_name"`
	ToWalletName string  `json:"to_wallet_name,omitempty"`
}