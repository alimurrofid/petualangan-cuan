package seeder

import (
	"cuan-backend/internal/entity"
	"time"
)

func relativeDate(days int) *time.Time {
	t := time.Now().AddDate(0, 0, days)
	return &t
}

func t(wID uint, cID uint, amount float64, typ string, desc string, daysAgo int) entity.Transaction {
	return entity.Transaction{
		WalletID:    wID,
		CategoryID:  cID,
		Amount:      amount,
		Type:        typ,
		Description: desc,
		Date:        *relativeDate(-daysAgo),
	}
}

func intPtr(i int) *int {
	return &i
}

var (
	Users = []entity.User{
		{
			Name:   "Budiono Siregar",
			Email:  "budiono@example.com",
			Payday: intPtr(25),
		},
	}

	Wallets = []entity.Wallet{
		{Name: "Seabank", Type: "Bank", Balance: 21416986, Icon: "Em_Coin", Seq: 1},             // ID: 1
		{Name: "Uang Tunai", Type: "Cash", Balance: 87700, Icon: "Em_DollarBill", Seq: 2},      // ID: 2
		{Name: "Mandiri", Type: "Bank", Balance: 109159, Icon: "Em_Bank", Seq: 3},              // ID: 3
		{Name: "Gopay", Type: "E-Wallet", Balance: 16402, Icon: "SmartphoneNfc", Seq: 4},       // ID: 4
		{Name: "Shopeepay", Type: "E-Wallet", Balance: 345, Icon: "SmartphoneNfc", Seq: 5},     // ID: 5
		{Name: "BTN", Type: "Bank", Balance: 51982, Icon: "Em_Bank", Seq: 6},                   // ID: 6
		{Name: "Jago Utama", Type: "Bank", Balance: 982, Icon: "Em_Bank", Seq: 7},              // ID: 7
		{Name: "Jago Learning", Type: "Bank", Balance: 17046, Icon: "Em_Bank", Seq: 8},         // ID: 8
	}

	Categories = []entity.Category{
		{Name: "Gaji", Type: "income", Icon: "Em_MoneyBag", BudgetLimit: 0, Seq: 1},                  // ID: 1
		{Name: "Pemasukan Lainnya", Type: "income", Icon: "HelpCircle", BudgetLimit: 0, Seq: 2},      // ID: 2
		{Name: "Makanan", Type: "expense", Icon: "Utensils", BudgetLimit: 0, Seq: 3},                 // ID: 3
		{Name: "Minuman", Type: "expense", Icon: "Wine", BudgetLimit: 0, Seq: 4},                     // ID: 4
		{Name: "Belanja Online", Type: "expense", Icon: "ShoppingCart", BudgetLimit: 0, Seq: 5},      // ID: 5
		{Name: "Belanja Kebutuhan", Type: "expense", Icon: "Em_Cart", BudgetLimit: 0, Seq: 6},        // ID: 6
		{Name: "Transfer", Type: "transfer", Icon: "Em_Exchange", BudgetLimit: 0, Seq: 7},            // ID: 7
		{Name: "Bensin", Type: "expense", Icon: "Fuel", BudgetLimit: 400000, Seq: 8},                 // ID: 8
		{Name: "Servis Motor", Type: "expense", Icon: "Em_Motor", BudgetLimit: 0, Seq: 9},            // ID: 9
		{Name: "Rokok", Type: "expense", Icon: "Em_Cigarette", BudgetLimit: 600000, Seq: 10},         // ID: 10
		{Name: "Bunga Bank", Type: "income", Icon: "BadgeDollarSign", BudgetLimit: 0, Seq: 11},       // ID: 11
		{Name: "Amal", Type: "expense", Icon: "HeartHandshake", BudgetLimit: 0, Seq: 12},             // ID: 12
		{Name: "Parkir", Type: "expense", Icon: "ParkingSquare", BudgetLimit: 0, Seq: 13},            // ID: 13
		{Name: "Jajan", Type: "expense", Icon: "Pizza", BudgetLimit: 0, Seq: 14},                     // ID: 14
		{Name: "Peralatan", Type: "expense", Icon: "Em_Tool", BudgetLimit: 0, Seq: 15},               // ID: 15
		{Name: "Kos", Type: "expense", Icon: "Building", BudgetLimit: 0, Seq: 16},                    // ID: 16
		{Name: "Listrik", Type: "expense", Icon: "Zap", BudgetLimit: 0, Seq: 17},                     // ID: 17
		{Name: "ATK", Type: "expense", Icon: "Tv", BudgetLimit: 0, Seq: 18},                          // ID: 18
		{Name: "Lifestyle", Type: "expense", Icon: "Em_Shirt", BudgetLimit: 0, Seq: 19},              // ID: 19
		{Name: "Lain lain", Type: "expense", Icon: "HelpCircle", BudgetLimit: 0, Seq: 20},            // ID: 20
		{Name: "Piutang", Type: "expense", Icon: "BanknoteArrowUp", BudgetLimit: 0, Seq: 21},         // ID: 21
		{Name: "Ngopi", Type: "expense", Icon: "Coffee", BudgetLimit: 150000, Seq: 22},               // ID: 23
		{Name: "Terima Piutang", Type: "income", Icon: "HandCoins", BudgetLimit: 0, Seq: 23},         // ID: 23
		{Name: "Internet", Type: "expense", Icon: "Wifi", BudgetLimit: 0, Seq: 24},                   // ID: 24
		{Name: "Biaya Admin", Type: "expense", Icon: "Em_MoneyWing", BudgetLimit: 0, Seq: 25},        // ID: 25
		{Name: "Admin Bulanan Bank", Type: "expense", Icon: "Em_Bank", BudgetLimit: 0, Seq: 26},      // ID: 26
		{Name: "Laundry", Type: "expense", Icon: "WashingMachine", BudgetLimit: 0, Seq: 27},          // ID: 27
		{Name: "Utang", Type: "income", Icon: "BanknoteArrowDown", BudgetLimit: 0, Seq: 28},          // ID: 28
		{Name: "Bayar Utang", Type: "expense", Icon: "CircleFadingArrowUp", BudgetLimit: 0, Seq: 29}, // ID: 29
		{Name: "Tabungan", Type: "saving_allocation", Icon: "PiggyBank", BudgetLimit: 0, Seq: 30},    // ID: 30
	}

	// Data transaksi diekstrak (12 Jan - 11 Feb) dan dimapping sempurna
	Transactions = []entity.Transaction{
		// --- 30 Hari Lalu (12 Jan) ---
		t(1, 2, 25000000, "income", "Pencairan Project Freelance Aplikasi", 30), // Seabank -> Pemasukan Lainnya
		t(3, 1, 4500000, "income", "Gaji Bulan Januari", 30), // Mandiri -> Gaji
		t(3, 8, 50000, "expense", "Bensin BP", 30),           // Mandiri -> Bensin
		t(1, 5, 44989, "expense", "Celana Jas Hujan", 30),    // Seabank -> Belanja Online
		t(1, 3, 18000, "expense", "Warteg Oren", 30),         // Seabank -> Makanan
		t(1, 4, 5000, "expense", "Gooday Capucino", 30),      // Seabank -> Minuman
		t(1, 29, 1200000, "expense", "Pelunasan Cicilan Tokopedia Helm", 30), // Seabank -> Bayar Utang

		// --- 29 Hari Lalu (13 Jan) ---
		t(3, 10, 25300, "expense", "Djarum Super", 29),                // Mandiri -> Rokok
		t(1, 3, 11000, "expense", "Sarapan udang butter garlic", 29),  // Seabank -> Makanan
		t(2, 10, 20000, "expense", "Gudang Garam Merah 16", 29),       // Uang Tunai -> Rokok
		t(2, 6, 8500, "expense", "Gula 1/2 Kilo", 29),                 // Uang Tunai -> Belanja Kebutuhan
		t(1, 4, 5000, "expense", "Gooday Capucino", 29),               // Seabank -> Minuman
		t(1, 30, 10000000, "saving_allocation", "Alokasi ke Upgrade Laptop M3 (Tahap 1)", 29), // Seabank -> Tabungan

		// --- 28 Hari Lalu (14 Jan) ---
		t(2, 10, 20500, "expense", "Djisamsu maestro edition", 28), // Uang Tunai -> Rokok
		t(1, 4, 10000, "expense", "Gooday Capucino", 28),           // Seabank -> Minuman
		t(2, 22, 5000, "expense", "Nongkrong di TKK", 28),          // Uang Tunai -> Ngopi
		t(1, 30, 2500000, "saving_allocation", "Alokasi ke Dana Darurat (Tahap 1)", 28), // Seabank -> Tabungan

		// --- 27 Hari Lalu (15 Jan) ---
		t(1, 3, 18000, "expense", "Nasi warteg oren", 27), // Seabank -> Makanan
		t(1, 3, 11000, "expense", "Nasi Kuning", 27),      // Seabank -> Makanan
		t(2, 10, 25000, "expense", "Pena Gold", 27),       // Uang Tunai -> Rokok
		t(1, 3, 3500, "expense", "Risol Mayo", 27),        // Seabank -> Makanan
		t(1, 3, 2500, "expense", "Martabak Mini", 27),     // Seabank -> Makanan
		t(1, 4, 5000, "expense", "Gooday Capucino", 27),   // Seabank -> Minuman

		// --- 26 Hari Lalu (16 Jan) ---
		t(2, 10, 18000, "expense", "Raptor", 26),                     // Uang Tunai -> Rokok
		t(1, 3, 12000, "expense", "Makan siang di warteg oren", 26),  // Seabank -> Makanan
		t(1, 3, 11000, "expense", "Mie kare + telur", 26),            // Seabank -> Makanan
		t(1, 22, 10000, "expense", "Nongkrong di wartawanti", 26),    // Seabank -> Ngopi
		t(2, 12, 5000, "expense", "Amal jumat", 26),                  // Uang Tunai -> Amal
		t(2, 3, 2000, "expense", "Es miki miki satian", 26),          // Uang Tunai -> Makanan

		// --- 25 Hari Lalu (17 Jan) ---
		t(1, 5, 101800, "expense", "Alat Cukur", 25),          // Seabank -> Belanja Online
		t(3, 8, 50000, "expense", "Bensin BP", 25),            // Mandiri -> Bensin
		t(2, 10, 24000, "expense", "Djarum king", 25),         // Uang Tunai -> Rokok
		t(1, 22, 16500, "expense", "Jus alpukat", 25),         // Seabank -> Ngopi
		t(2, 4, 12000, "expense", "Kopi saku indomaret", 25),  // Uang Tunai -> Minuman
		t(2, 3, 9000, "expense", "Nasi pecel", 25),            // Uang Tunai -> Makanan
		t(1, 4, 5000, "expense", "Gooday Capucino", 25),       // Seabank -> Minuman
		t(2, 15, 3500, "expense", "Korek Tokai", 25),          // Uang Tunai -> Peralatan
		t(1, 30, 1500000, "saving_allocation", "Alokasi ke Liburan Jogja (Tahap 1)", 25), // Seabank -> Tabungan

		// --- 24 Hari Lalu (18 Jan) ---
		t(2, 9, 575000, "expense", "Ban Belakang IRC", 24),           // Uang Tunai -> Servis Motor
		t(2, 9, 90000, "expense", "Oli Deltalube", 24),               // Uang Tunai -> Servis Motor
		t(2, 9, 50000, "expense", "Kampas Rem Belakang Beat", 24),    // Uang Tunai -> Servis Motor
		t(2, 12, 50000, "expense", "Sodakoh iil", 24),                // Uang Tunai -> Amal
		t(2, 18, 24000, "expense", "Print", 24),                      // Uang Tunai -> ATK
		t(2, 12, 20000, "expense", "Sodakoh azmi", 24),               // Uang Tunai -> Amal
		t(2, 12, 20000, "expense", "Sodakoh ibrahim", 24),            // Uang Tunai -> Amal
		t(2, 19, 18000, "expense", "Potong Rambut", 24),              // Uang Tunai -> Lifestyle
		t(2, 10, 17000, "expense", "Raptor", 24),                     // Uang Tunai -> Rokok
		t(2, 4, 7500, "expense", "Oat coffe", 24),                    // Uang Tunai -> Minuman

		// --- 23 Hari Lalu (19 Jan) ---
		t(1, 11, 28500, "income", "Bunga Tabungan Seabank", 23),         // Seabank -> Bunga Bank
		t(3, 8, 100000, "expense", "Bensin BP", 23),                     // Mandiri -> Bensin
		t(2, 20, 21000, "expense", "Lupa", 23),                          // Uang Tunai -> Lain lain
		t(2, 10, 20000, "expense", "Gudang garam merah 16", 23),         // Uang Tunai -> Rokok
		t(1, 3, 8000, "expense", "Nasi kuning", 23),                     // Seabank -> Makanan
		t(2, 3, 8000, "expense", "Bayar utang es teler ke rizaldi", 23), // Uang Tunai -> Makanan
		t(2, 3, 7000, "expense", "Bayar utang Ayam geprek ke ali", 23),  // Uang Tunai -> Makanan
		t(2, 3, 6000, "expense", "Bayar utang ke satian", 23),           // Uang Tunai -> Makanan
		t(1, 4, 5000, "expense", "Gooday Capucino", 23),                 // Seabank -> Minuman
		t(1, 14, 5000, "expense", "Gorengan", 23),                       // Seabank -> Jajan
		t(2, 6, 4000, "expense", "Bedak mbk", 23),                       // Uang Tunai -> Belanja Kebutuhan

		// --- 22 Hari Lalu (20 Jan) ---
		t(1, 6, 18000, "expense", "Beras", 22),            // Seabank -> Belanja Kebutuhan
		t(2, 10, 18000, "expense", "Raptor", 22),          // Uang Tunai -> Rokok
		t(1, 5, 17475, "expense", "Tempered glass", 22),   // Seabank -> Belanja Online
		t(1, 5, 16830, "expense", "Case HP", 22),          // Seabank -> Belanja Online
		t(1, 3, 8000, "expense", "Nasi Kuning", 22),       // Seabank -> Makanan

		// --- 21 Hari Lalu (21 Jan) ---
		t(1, 3, 20000, "expense", "Hotways", 21),       // Seabank -> Makanan
		t(3, 10, 18700, "expense", "Raptor", 21),       // Mandiri -> Rokok
		t(3, 10, 18700, "expense", "Raptor", 21),       // Mandiri -> Rokok
		t(2, 10, 18000, "expense", "Raptor", 21),       // Uang Tunai -> Rokok
		t(1, 3, 10000, "expense", "Nasi kuning", 21),   // Seabank -> Makanan
		t(1, 3, 9000, "expense", "Nasi bungkus", 21),   // Seabank -> Makanan
		t(1, 4, 5000, "expense", "Kopi hitam es", 21),  // Seabank -> Minuman

		// --- 20 Hari Lalu (22 Jan) ---
		t(2, 9, 20000, "expense", "Cuci motor", 20),              // Uang Tunai -> Servis Motor
		t(2, 10, 20000, "expense", "Gudang garam deluxe 16", 20), // Uang Tunai -> Rokok
		t(1, 3, 10000, "expense", "Nasi lada hitam", 20),         // Seabank -> Makanan
		t(2, 12, 7000, "expense", "Patungan Kue JB", 20),         // Uang Tunai -> Amal
		t(1, 4, 5000, "expense", "Es kopi hitam", 20),            // Seabank -> Minuman

		// --- 19 Hari Lalu (23 Jan) ---
		t(1, 4, 5000, "expense", "Gooday Capucino", 19), // Seabank -> Minuman

		// --- 18 Hari Lalu (24 Jan) ---
		t(1, 16, 800000, "expense", "Bayar kos", 18),             // Seabank -> Kos
		t(1, 22, 20000, "expense", "Caramel machiato", 18),       // Seabank -> Ngopi
		t(2, 10, 18000, "expense", "Raptor", 18),                 // Uang Tunai -> Rokok
		t(1, 5, 16930, "expense", "Sunscreen hanasui", 18),       // Seabank -> Belanja Online
		t(1, 22, 13000, "expense", "Kopi kenangan", 18),          // Seabank -> Ngopi
		t(1, 22, 13000, "expense", "Kopi kenangan", 18),          // Seabank -> Ngopi
		t(1, 3, 10000, "expense", "Nasi ayam lada hitam", 18),    // Seabank -> Makanan
		t(1, 30, 1500000, "saving_allocation", "Alokasi ke Liburan Jogja (Tahap 2)", 18), // Seabank -> Tabungan

		// --- 17 Hari Lalu (25 Jan) ---
		t(2, 10, 20000, "expense", "Gudang garam merah 16", 17), // Uang Tunai -> Rokok
		t(2, 10, 17000, "expense", "New star", 17),              // Uang Tunai -> Rokok

		// --- 16 Hari Lalu (26 Jan) ---
		t(1, 11, 32100, "income", "Bunga Tabungan Seabank", 16), // Seabank -> Bunga Bank
		t(2, 4, 20000, "expense", "Kopi Tubruk Gajah", 16), // Uang Tunai -> Minuman
		t(1, 10, 18000, "expense", "Gajah baru", 16),       // Seabank -> Rokok
		t(1, 3, 10000, "expense", "Nasi Kuning", 16),       // Seabank -> Makanan
		t(1, 4, 5000, "expense", "Gooday Capucino", 16),    // Seabank -> Minuman
		t(2, 3, 3500, "expense", "Kerupuk", 16),            // Uang Tunai -> Makanan

		// --- 15 Hari Lalu (27 Jan) ---
		t(1, 6, 47400, "expense", "Belanja di remaja", 15),       // Seabank -> Belanja Kebutuhan
		t(1, 5, 24320, "expense", "Celana pendek", 15),           // Seabank -> Belanja Online
		t(2, 10, 18000, "expense", "Raptor", 15),                 // Uang Tunai -> Rokok
		t(1, 3, 9000, "expense", "Nasi sambel goreng ati", 15),   // Seabank -> Makanan
		t(1, 3, 9000, "expense", "Nasi satian", 15),              // Seabank -> Makanan
		t(1, 4, 5000, "expense", "Marimas cincau", 15),           // Seabank -> Minuman
		t(1, 29, 1000000, "expense", "Cicilan Pinjaman Laptop Budi", 15), // Seabank -> Bayar Utang

		// --- 14 Hari Lalu (28 Jan) ---
		t(1, 3, 54000, "expense", "Pecel ponorogo", 14),     // Seabank -> Makanan
		t(1, 22, 46000, "expense", "Kopi bento", 14),        // Seabank -> Ngopi
		t(2, 10, 25000, "expense", "Djarum super", 14),      // Uang Tunai -> Rokok
		t(2, 10, 10000, "expense", "Merah delima", 14),      // Uang Tunai -> Rokok
		t(1, 3, 10000, "expense", "Nasi lada hitam", 14),    // Seabank -> Makanan
		t(2, 10, 10000, "expense", "Djarumku bold", 14),     // Uang Tunai -> Rokok
		t(2, 13, 6000, "expense", "Parkir kopi bento", 14),  // Uang Tunai -> Parkir
		t(1, 4, 5000, "expense", "Popice stoberi", 14),      // Seabank -> Minuman
		t(2, 13, 4000, "expense", "Parkir pecel", 14),       // Uang Tunai -> Parkir
		t(1, 14, 2000, "expense", "Kerupuk", 14),            // Seabank -> Jajan
		t(1, 30, 2000000, "saving_allocation", "Alokasi ke Dana Darurat (Tahap 2)", 14), // Seabank -> Tabungan

		// --- 13 Hari Lalu (29 Jan) ---
		t(1, 3, 10000, "expense", "Nasi kuning", 13),          // Seabank -> Makanan
		t(2, 6, 8500, "expense", "Tusuk gigi alfamart", 13),   // Uang Tunai -> Belanja Kebutuhan
		t(1, 4, 5000, "expense", "Gooday capucino", 13),       // Seabank -> Minuman
		t(1, 14, 2500, "expense", "Risoles", 13),              // Seabank -> Jajan
		t(1, 14, 2000, "expense", "Kerupuk", 13),              // Seabank -> Jajan

		// --- 12 Hari Lalu (30 Jan) ---
		t(4, 17, 20900, "expense", "Listrik bulanan", 12),       // Gopay -> Listrik
		t(2, 10, 20000, "expense", "Gudang garam merah 16", 12), // Uang Tunai -> Rokok
		t(1, 3, 10000, "expense", "Nasi kuning", 12),            // Seabank -> Makanan
		t(1, 4, 5000, "expense", "Gooday capucino", 12),         // Seabank -> Minuman
		t(2, 4, 5000, "expense", "Es degan", 12),                // Uang Tunai -> Minuman
		t(1, 14, 2500, "expense", "Tahu sumedang", 12),          // Seabank -> Jajan
		t(1, 5, 350000, "expense", "Mouse Wireless Logitech", 12), // Seabank -> Belanja Online

		// --- 11 Hari Lalu (31 Jan) ---
		t(1, 22, 15000, "expense", "Kopi bening upn", 11),       // Seabank -> Ngopi
		t(1, 3, 12000, "expense", "Mie sosis bening upn", 11),   // Seabank -> Makanan
		t(2, 3, 12000, "expense", "Tahu tek", 11),               // Uang Tunai -> Makanan
		t(2, 10, 10000, "expense", "Gajah merah", 11),           // Uang Tunai -> Rokok
		t(3, 26, 5000, "expense", "Admin bulanan", 11),          // Mandiri -> Admin Bulanan Bank
		t(1, 4, 5000, "expense", "Gooday capucino", 11),         // Seabank -> Minuman

		// --- 10 Hari Lalu (1 Feb) ---
		t(1, 20, 100000, "expense", "Urunan ke pandaan", 10),    // Seabank -> Lain lain
		t(1, 5, 61900, "expense", "Celana Pendek Disai", 10),    // Seabank -> Belanja Online
		t(2, 10, 26000, "expense", "Djarum super", 10),          // Uang Tunai -> Rokok
		t(2, 10, 25000, "expense", "Djarum super", 10),          // Uang Tunai -> Rokok
		t(1, 5, 24925, "expense", "Kabel rol 12m", 10),          // Seabank -> Belanja Online
		t(1, 3, 20000, "expense", "Nasi warteg", 10),            // Seabank -> Makanan
		t(1, 22, 20000, "expense", "Fore", 10),                  // Seabank -> Ngopi
		t(2, 3, 15000, "expense", "Nasi goreng", 10),            // Uang Tunai -> Makanan
		t(2, 3, 10000, "expense", "Soto daging", 10),            // Uang Tunai -> Makanan
		t(2, 3, 3500, "expense", "Kerupuk", 10),                 // Uang Tunai -> Makanan
		t(1, 30, 4000000, "saving_allocation", "Alokasi ke Upgrade Laptop M3 (Tahap 2)", 10), // Seabank -> Tabungan

		// --- 9 Hari Lalu (2 Feb) ---
		t(1, 11, 31400, "income", "Bunga Tabungan Seabank", 9),  // Seabank -> Bunga Bank
		t(2, 10, 20000, "expense", "Gudang garam merah 16", 9), // Uang Tunai -> Rokok
		t(2, 10, 18000, "expense", "Gajah baru", 9),            // Uang Tunai -> Rokok
		t(2, 3, 10000, "expense", "Ayam geprek", 9),            // Uang Tunai -> Makanan
		t(1, 4, 5000, "expense", "Gooday capucino", 9),         // Seabank -> Minuman

		// --- 8 Hari Lalu (3 Feb) ---
		t(2, 10, 20000, "expense", "Gudang garam 16", 8), // Uang Tunai -> Rokok
		t(2, 3, 15000, "expense", "Sate ayam", 8),        // Uang Tunai -> Makanan
		t(1, 3, 10000, "expense", "Nasi kuning", 8),      // Seabank -> Makanan
		t(1, 4, 5000, "expense", "Gooday capucino", 8),   // Seabank -> Minuman
		t(1, 3, 3000, "expense", "Jajan septian", 8),     // Seabank -> Makanan

		// --- 7 Hari Lalu (4 Feb) ---
		t(1, 3, 13000, "expense", "Bubur ayam", 7),       // Seabank -> Makanan
		t(1, 14, 10000, "expense", "Feilala 4 biji", 7),  // Seabank -> Jajan
		t(1, 4, 5000, "expense", "Gooday capucino", 7),   // Seabank -> Minuman

		// --- 6 Hari Lalu (5 Feb) ---
		t(1, 10, 18000, "expense", "Gajah baru", 6),      // Seabank -> Rokok
		t(1, 3, 10000, "expense", "Ricebowl", 6),         // Seabank -> Makanan
		t(1, 3, 10000, "expense", "Ayam geprek", 6),      // Seabank -> Makanan
		t(1, 4, 5000, "expense", "Gooday capucino", 6),   // Seabank -> Minuman
		t(1, 3, 2500, "expense", "Kerupuk", 6),           // Seabank -> Makanan

		// --- 5 Hari Lalu (6 Feb) ---
		t(3, 8, 100000, "expense", "Bensin bp", 5),         // Mandiri -> Bensin
		t(1, 10, 25000, "expense", "Djarum super", 5),      // Seabank -> Rokok
		t(1, 22, 18850, "expense", "Fore", 5),              // Seabank -> Ngopi
		t(2, 10, 18000, "expense", "Raptor", 5),            // Uang Tunai -> Rokok
		t(1, 3, 16000, "expense", "Nasi warteg oren", 5),   // Seabank -> Makanan
		t(1, 3, 10000, "expense", "Nasi ayam buldak", 5),   // Seabank -> Makanan
		t(2, 3, 10000, "expense", "Ayam geprek", 5),        // Uang Tunai -> Makanan
		t(1, 4, 5000, "expense", "Nescafe", 5),             // Seabank -> Minuman
		t(1, 4, 5000, "expense", "Nutrisari jeruk", 5),     // Seabank -> Minuman
		t(2, 13, 3000, "expense", "Parkir fore", 5),        // Uang Tunai -> Parkir
		t(2, 12, 2000, "expense", "Amal jumat", 5),         // Uang Tunai -> Amal
		t(1, 14, 2000, "expense", "Sosis", 5),              // Seabank -> Jajan
		t(2, 23, 50000, "income", "Cicilan talangan bensin dari Andi", 5), // Uang Tunai -> Terima Piutang

		// --- 4 Hari Lalu (7 Feb) ---
		t(1, 3, 10000, "expense", "Nasi ayam lada hitam", 4), // Seabank -> Makanan
		t(2, 4, 7800, "expense", "Cafino", 4),                // Uang Tunai -> Minuman

		// --- 3 Hari Lalu (8 Feb) ---
		t(1, 2, 30000, "income", "Uang bensin dari teman", 3),    // Seabank -> Pemasukan Lainnya
		t(2, 10, 20000, "expense", "Gudang garam merah 16", 3),   // Uang Tunai -> Rokok
		t(1, 22, 17850, "expense", "Vietnam drip", 3),            // Seabank -> Ngopi
		t(2, 13, 3000, "expense", "Parkir alas trawas", 3),       // Uang Tunai -> Parkir

		// --- 2 Hari Lalu (9 Feb) ---
		t(1, 11, 24986, "income", "Bunga Tabungan Seabank", 2),  // Seabank -> Bunga Bank
		t(4, 24, 45500, "expense", "Kuota internet", 2),  // Gopay -> Internet
		t(2, 10, 13000, "expense", "Smith hijau", 2),     // Uang Tunai -> Rokok
		t(2, 3, 12000, "expense", "Nasi goreng", 2),      // Uang Tunai -> Makanan
		t(2, 3, 10000, "expense", "Nasi pecel", 2),       // Uang Tunai -> Makanan
		t(1, 3, 10000, "expense", "Nasi daging", 2),      // Seabank -> Makanan
		t(1, 4, 5000, "expense", "Es milo", 2),           // Seabank -> Minuman

		// --- 1 Hari Lalu (10 Feb) ---
		t(1, 19, 50049, "expense", "Celana pendek disai", 1),   // Seabank -> Lifestyle
		t(2, 10, 20000, "expense", "Gudang garam merah 16", 1), // Uang Tunai -> Rokok
		t(1, 3, 10000, "expense", "Nasi kuning", 1),            // Seabank -> Makanan
		t(1, 4, 5000, "expense", "Nutrisari leci", 1),          // Seabank -> Minuman

		// --- Hari Ini / 0 Hari Lalu (11 Feb) ---
		t(1, 6, 34000, "expense", "Mie dan kopi", 0), // Seabank -> Belanja Kebutuhan
		t(1, 3, 10000, "expense", "Ayam Dkriuk", 0),  // Seabank -> Makanan
	}

	Debts = []entity.Debt{
		{
			WalletID:    1, // Seabank
			Name:        "Pinjaman Budi (Laptop)",
			Amount:      2500000,
			Remaining:   1500000,
			Type:        entity.DebtTypePayable,
			Description: "Pinjam untuk DP servis laptop",
			DueDate:     relativeDate(14), // Jatuh tempo 14 hari lagi
			IsPaid:      false,
		},
		{
			WalletID:    3, // Mandiri
			Name:        "Paylater Shopee",
			Amount:      450000,
			Remaining:   450000,
			Type:        entity.DebtTypePayable,
			Description: "Tagihan jaket & jas hujan",
			DueDate:     relativeDate(-3), // OVERDUE 3 hari lalu
			IsPaid:      false,
		},
		{
			WalletID:    2, // Uang Tunai
			Name:        "Talangan Bensin Andi",
			Amount:      100000,
			Remaining:   50000,
			Type:        entity.DebtTypeReceivable,
			Description: "Andi pinjam uang bensin saat touring",
			DueDate:     relativeDate(7), // Jatuh tempo 7 hari lagi
			IsPaid:      false,
		},
		{
			WalletID:    1, // Seabank
			Name:        "Cicilan Tokopedia (Selesai)",
			Amount:      1200000,
			Remaining:   0,
			Type:        entity.DebtTypePayable,
			Description: "Cicilan helm full face (sudah lunas)",
			DueDate:     relativeDate(-30),
			IsPaid:      true,
		},
	}

	Wishlists = []entity.WishlistItem{
		{
			CategoryID:     5, // Belanja Online
			Name:           "AirPods Pro 2",
			EstimatedPrice: 3500000,
			Priority:       entity.WishlistPriorityHigh,
			IsBought:       false,
		},
		{
			CategoryID:     19, // Lifestyle
			Name:           "Sepatu Lari Nike Pegasus",
			EstimatedPrice: 1650000,
			Priority:       entity.WishlistPriorityMedium,
			IsBought:       false,
		},
		{
			CategoryID:     15, // Peralatan
			Name:           "Mechanical Keyboard Keychron",
			EstimatedPrice: 1200000,
			Priority:       entity.WishlistPriorityLow,
			IsBought:       false,
		},
		{
			CategoryID:     5, // Belanja Online
			Name:           "Mouse Wireless Logitech",
			EstimatedPrice: 350000,
			Priority:       entity.WishlistPriorityMedium,
			IsBought:       true, // Sudah terbeli
		},
	}

	SavingGoals = []entity.SavingGoal{
		{
			CategoryID:    20, // Lain lain
			Name:          "Dana Darurat",
			TargetAmount:  15000000,
			CurrentAmount: 4500000, // 30% progress
			Deadline:      relativeDate(180),
			Icon:          "ShieldAlert",
			IsAchieved:    false,
			IsFinished:    false,
		},
		{
			CategoryID:    5, // Belanja Online
			Name:          "Upgrade Laptop M3",
			TargetAmount:  20000000,
			CurrentAmount: 14000000, // 70% progress
			Deadline:      relativeDate(90),
			Icon:          "Laptop",
			IsAchieved:    false,
			IsFinished:    false,
		},
		{
			CategoryID:    19, // Lifestyle
			Name:          "Liburan Jogja",
			TargetAmount:  3000000,
			CurrentAmount: 3000000, // 100% progress
			Deadline:      relativeDate(-15),
			Icon:          "Plane",
			IsAchieved:    true,
			IsFinished:    true,
		},
	}

	DebtPayments = []SeedDebtPayment{
		{
			DebtIndex:       3, // Cicilan Tokopedia (Selesai)
			TransactionDesc: "Pelunasan Cicilan Tokopedia Helm",
			WalletIndex:     1, // Seabank
			Amount:          1200000,
			DaysAgo:         30,
			Note:            "Pelunasan cicilan helm full face",
		},
		{
			DebtIndex:       0, // Pinjaman Budi (Laptop)
			TransactionDesc: "Cicilan Pinjaman Laptop Budi",
			WalletIndex:     1, // Seabank
			Amount:          1000000,
			DaysAgo:         15,
			Note:            "Cicilan ke-1 DP servis laptop",
		},
		{
			DebtIndex:       2, // Talangan Bensin Andi
			TransactionDesc: "Cicilan talangan bensin dari Andi",
			WalletIndex:     2, // Uang Tunai
			Amount:          50000,
			DaysAgo:         5,
			Note:            "Pengembalian uang bensin touring",
		},
	}

	SavingContributions = []SeedSavingContribution{
		{
			GoalIndex:       0, // Dana Darurat
			TransactionDesc: "Alokasi ke Dana Darurat (Tahap 1)",
			WalletIndex:     1, // Seabank
			Amount:          2500000,
			DaysAgo:         28,
		},
		{
			GoalIndex:       0, // Dana Darurat
			TransactionDesc: "Alokasi ke Dana Darurat (Tahap 2)",
			WalletIndex:     1, // Seabank
			Amount:          2000000,
			DaysAgo:         14,
		},
		{
			GoalIndex:       1, // Upgrade Laptop M3
			TransactionDesc: "Alokasi ke Upgrade Laptop M3 (Tahap 1)",
			WalletIndex:     1, // Seabank
			Amount:          10000000,
			DaysAgo:         29,
		},
		{
			GoalIndex:       1, // Upgrade Laptop M3
			TransactionDesc: "Alokasi ke Upgrade Laptop M3 (Tahap 2)",
			WalletIndex:     1, // Seabank
			Amount:          4000000,
			DaysAgo:         10,
		},
		{
			GoalIndex:       2, // Liburan Jogja
			TransactionDesc: "Alokasi ke Liburan Jogja (Tahap 1)",
			WalletIndex:     1, // Seabank
			Amount:          1500000,
			DaysAgo:         25,
		},
		{
			GoalIndex:       2, // Liburan Jogja
			TransactionDesc: "Alokasi ke Liburan Jogja (Tahap 2)",
			WalletIndex:     1, // Seabank
			Amount:          1500000,
			DaysAgo:         18,
		},
	}
)

type SeedDebtPayment struct {
	DebtIndex       int
	TransactionDesc string
	WalletIndex     int
	Amount          float64
	DaysAgo         int
	Note            string
}

type SeedSavingContribution struct {
	GoalIndex       int
	TransactionDesc string
	WalletIndex     int
	Amount          float64
	DaysAgo         int
}
