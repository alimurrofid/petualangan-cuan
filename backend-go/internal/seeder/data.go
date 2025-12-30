package seeder

import (
	"cuan-backend/internal/entity"
	"time"
)

var (
	Users = []entity.User{
		{
			Name:  "Budiono Siregar",
			Email: "budiono@example.com",
		},
		{
			Name:  "Rudi Tabudi",
			Email: "rudi@example.com",
		},
	}

	Wallets = []entity.Wallet{
		{
			Name:    "Dompet Cash",
			Type:    "Cash",
			Balance: 500000,
			Icon:    "Em_DollarBill",
		},
		{
			Name:    "BCA",
			Type:    "Bank",
			Balance: 15000000,
			Icon:    "Em_Bank",
		},
		{
			Name:    "GoPay",
			Type:    "E-Wallet",
			Balance: 250000,
			Icon:    "Em_MoneyWing",
		},
	}

	Categories = []entity.Category{
		{
			Name:        "Gaji",
			Type:        "income",
			Icon:        "Em_MoneyBag",
			BudgetLimit: 0,
		},
		{
			Name:        "Makan & Minum",
			Type:        "expense",
			Icon:        "Em_Pizza",
			BudgetLimit: 3000000,
		},
		{
			Name:        "Transportasi",
			Type:        "expense",
			Icon:        "Bus",
			BudgetLimit: 1000000,
		},
		{
			Name:        "Belanja",
			Type:        "expense",
			Icon:        "ShoppingCart",
			BudgetLimit: 2000000,
		},
	}

	Transactions = []entity.Transaction{
		{
			Description:   "Gaji Bulanan",
			Type:   "income",
			Amount: 15000000,
			Date:   time.Now().AddDate(0, 0, -5), // 5 days ago
		},
		{
			Description:   "Makan Siang",
			Type:   "expense",
			Amount: 50000,
			Date:   time.Now().AddDate(0, 0, -2), // 2 days ago
		},
		{
			Description:   "Naik Grab",
			Type:   "expense",
			Amount: 35000,
			Date:   time.Now().AddDate(0, 0, -1), // Yesterday
		},
		{
			Description:   "Belanja Bulanan",
			Type:   "expense",
			Amount: 1500000,
			Date:   time.Now().AddDate(0, 0, -3), // 3 days ago
		},
	}
)
