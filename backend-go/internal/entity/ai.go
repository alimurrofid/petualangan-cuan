package entity

type AIProcessResponse []TransactionItem

type TransactionItem struct {
	Item     string `json:"item"`
	Category string `json:"category"`
	Amount   int    `json:"amount"`
	Wallet   string `json:"wallet"`
	Note     string `json:"note"`
}