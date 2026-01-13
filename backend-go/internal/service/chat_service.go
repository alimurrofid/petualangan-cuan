package service

import (
	"cuan-backend/internal/service/ai_provider"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type ChatService struct {
	Provider       ai_provider.AIProvider
	TransactionSvc TransactionService
	WalletSvc      WalletService
	CategorySvc    CategoryService
}

// NewAIProvider creates an AI provider based on configuration
func NewAIProvider() (ai_provider.AIProvider, error) {
	baseURL := os.Getenv("AI_BASE_URL")
	model := os.Getenv("AI_MODEL")
	return ai_provider.NewLocalAIProvider(baseURL, model), nil
}

func NewChatService(provider ai_provider.AIProvider, transactionSvc TransactionService, walletSvc WalletService, categorySvc CategoryService) *ChatService {
	return &ChatService{
		Provider:       provider,
		TransactionSvc: transactionSvc,
		WalletSvc:      walletSvc,
		CategorySvc:    categorySvc,
	}
}

// TransactionJSON structure matching the system prompt
type TransactionJSON struct {
	IsTransaction bool    `json:"is_transaction"`
	Type          string  `json:"type"`
	Amount        float64 `json:"amount"`
	Description   string  `json:"description"`
	CategoryName  string  `json:"category_name"`
	WalletName    string  `json:"wallet_name"`
}

func (s *ChatService) ProcessMessage(userID uint, message string, fileData []byte, mimeType string, attachmentPath string) (string, error) {
	wallets, _ := s.WalletSvc.GetUserWallets(userID)
	categories, _ := s.CategorySvc.GetCategories(userID)

	walletNames := []string{}
	for _, w := range wallets {
		walletNames = append(walletNames, w.Name)
	}

	categoryNames := []string{}
	for _, c := range categories {
		categoryNames = append(categoryNames, c.Name)
	}

	contextMsg := fmt.Sprintf("%s\n\n[System Info]\nAvailable Wallets: %v\nAvailable Categories: %v", message, walletNames, categoryNames)

	response, err := s.Provider.GenerateResponse(contextMsg, fileData, mimeType)
	if err != nil {
		return "", err
	}

	jsonStr := extractJSON(response)
	cleanResponse := response
	if jsonStr != "" {
		cleanResponse = strings.Replace(response, "```json", "", 1)
		cleanResponse = strings.Replace(cleanResponse, jsonStr, "", 1)
		cleanResponse = strings.Replace(cleanResponse, "```", "", 1)
		cleanResponse = strings.TrimSpace(cleanResponse)
	}
	cleanResponse = strings.ReplaceAll(cleanResponse, "**", "")
	cleanResponse = strings.ReplaceAll(cleanResponse, "*", "")
	cleanResponse = strings.TrimSpace(cleanResponse)
	
	if jsonStr != "" {
		var txDataList []TransactionJSON
		if err := json.Unmarshal([]byte(jsonStr), &txDataList); err != nil {
			var singleTx TransactionJSON
			if errSingle := json.Unmarshal([]byte(jsonStr), &singleTx); errSingle == nil {
				txDataList = []TransactionJSON{singleTx}
			}
		}

		savedCount := 0
		failedCount := 0
		
		for _, txData := range txDataList {
			if !txData.IsTransaction {
				continue
			}

			var walletID uint
			var categoryID uint

			for _, w := range wallets {
				if strings.EqualFold(w.Name, txData.WalletName) {
					walletID = w.ID
					break
				}
			}
			if walletID == 0 {
				for _, w := range wallets {
					if strings.Contains(strings.ToLower(txData.WalletName), strings.ToLower(w.Name)) || strings.Contains(strings.ToLower(w.Name), strings.ToLower(txData.WalletName)) {
						walletID = w.ID
						break
					}
				}
			}

			for _, c := range categories {
				if strings.EqualFold(c.Name, txData.CategoryName) {
					categoryID = c.ID
					break
				}
			}
			if categoryID == 0 {
				for _, c := range categories {
					if strings.Contains(strings.ToLower(txData.CategoryName), strings.ToLower(c.Name)) || strings.Contains(strings.ToLower(c.Name), strings.ToLower(txData.CategoryName)) {
						categoryID = c.ID
						break
					}
				}
			}

			if walletID != 0 && categoryID != 0 {
				_, err := s.TransactionSvc.CreateTransaction(userID, CreateTransactionInput{
					WalletID:    walletID,
					CategoryID:  categoryID,
					Amount:      txData.Amount,
					Type:        txData.Type,
					Description: txData.Description,
					Attachment:  attachmentPath,
					Date:        time.Now(),
				})

				if err == nil {
					savedCount++
				} else {
					failedCount++
				}
			} else {
				failedCount++
			}
		}

		if savedCount > 0 {
			cleanResponse += fmt.Sprintf("\n\n✅ Berhasil menyimpan %d transaksi!", savedCount)
		}
		if failedCount > 0 {
			cleanResponse += fmt.Sprintf("\n(⚠️ Gagal menyimpan %d item)", failedCount)
		}
	}

	return cleanResponse, nil
}

func extractJSON(s string) string {
	start := strings.Index(s, "```json")
	if start == -1 {
		return ""
	}
	start += 7
	end := strings.Index(s[start:], "```")
	if end == -1 {
		return ""
	}
	return strings.TrimSpace(s[start : start+end])
}
