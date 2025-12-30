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
	}

	Wallets = []entity.Wallet{
		{
			Name:    "Cash",
			Type:    "Cash",
			Balance: 200_000,
			Icon:    "Em_DollarBill",
		},
		{
			Name:    "BCA",
			Type:    "Bank",
			Balance: 3_000_000,
			Icon:    "Em_Bank",
		},
		{
			Name:    "GoPay",
			Type:    "E-Wallet",
			Balance: 300_000,
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
			BudgetLimit: 1_200_000,
		},
		{
			Name:        "Transportasi",
			Type:        "expense",
			Icon:        "Bus",
			BudgetLimit: 500_000,
		},
		{
			Name:        "Belanja",
			Type:        "expense",
			Icon:        "ShoppingCart",
			BudgetLimit: 800_000,
		},
	}

	Transactions = []entity.Transaction{
		{
			Description: "Gaji Bulanan",
			Type:        "income",
			Amount:      4_500_000,
			Date:        time.Now().AddDate(0, 0, -10),
		},
		{
			Description: "Sarapan - Nasi Uduk",
			Type:        "expense",
			Amount:      15_000,
			Date:        time.Now().AddDate(0, 0, -5),
		},
		{
			Description: "Kopi Pagi",
			Type:        "expense",
			Amount:      12_000,
			Date:        time.Now().AddDate(0, 0, -5),
		},
		{
			Description: "Makan Siang - Ayam Geprek",
			Type:        "expense",
			Amount:      25_000,
			Date:        time.Now().AddDate(0, 0, -4),
		},
		{
			Description: "Minum Boba",
			Type:        "expense",
			Amount:      20_000,
			Date:        time.Now().AddDate(0, 0, -4),
		},
		{
			Description: "Makan Malam - Pecel",
			Type:        "expense",
			Amount:      18_000,
			Date:        time.Now().AddDate(0, 0, -3),
		},
		{
			Description: "Naik Grab ke Kantor",
			Type:        "expense",
			Amount:      35_000,
			Date:        time.Now().AddDate(0, 0, -3),
		},
		{
			Description: "Naik Grab Pulang",
			Type:        "expense",
			Amount:      30_000,
			Date:        time.Now().AddDate(0, 0, -3),
		},
		{
			Description: "Belanja - Beras 5kg",
			Type:        "expense",
			Amount:      65_000,
			Date:        time.Now().AddDate(0, 0, -6),
		},
		{
			Description: "Belanja - Telur 1kg",
			Type:        "expense",
			Amount:      30_000,
			Date:        time.Now().AddDate(0, 0, -6),
		},
		{
			Description: "Belanja - Minyak Goreng 2L",
			Type:        "expense",
			Amount:      38_000,
			Date:        time.Now().AddDate(0, 0, -6),
		},
		{
			Description: "Roti Maryam",
			Type:        "expense",
			Amount:      15_000,
			Date:        time.Now().AddDate(0, 0, -1),
		},
		{
			Description: "Belanja - Sabun & Shampoo",
			Type:        "expense",
			Amount:      30_000,
			Date:        time.Now().AddDate(0, 0, 0),
		},
	}
)
