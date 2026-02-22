package service

import "strings"

// ContextIntent merepresentasikan topik utama yang dideteksi dari pesan user.
type ContextIntent int

const (
	IntentGeneral     ContextIntent = iota // default: kirim semua context
	IntentTransaction                      // catat transaksi baru
	IntentDebt                             // tanya/kelola utang-piutang
	IntentReport                           // rekap / laporan keuangan
	IntentGoal                             // tabungan / target keuangan
	IntentHealth                           // skor / kesehatan keuangan
	IntentSmallTalk                        // sapaan / pertanyaan umum non-keuangan
)

// Token budget — batas karakter context yang dikirim ke LLM.
// Sesuaikan dengan context window model yang dipakai.
const (
	// MaxContextChars adalah batas total karakter context (lebih besar untuk IntentReport
	// yang menyertakan daily breakdown ~30 baris).
	MaxContextChars = 4000

	// MaxRecentTxns adalah jumlah maksimal transaksi terakhir yang disertakan.
	MaxRecentTxns = 5

	// MaxDebtsInContext adalah jumlah maksimal utang/piutang aktif yang disertakan.
	MaxDebtsInContext = 5

	// MaxGoalsInContext adalah jumlah maksimal target tabungan yang disertakan.
	MaxGoalsInContext = 5
)

// intentKeywords mendefinisikan kata kunci per intent.
// Urutan di dalam slice tidak berpengaruh; semua dicek dengan strings.Contains.
var intentKeywords = map[ContextIntent][]string{
	IntentTransaction: {
		"beli", "bayar", "beli", "catat", "masuk", "keluar", "transfer",
		"makan", "belanja", "jajan", "isi", "topup", "top up", "tagihan",
		"listrik", "air", "cicilan", "debit", "kredit", "saldo",
	},
	IntentDebt: {
		"utang", "piutang", "pinjam", "pinjaman", "bayar utang", "cicil",
		"minjem", "ngutang", "hutang", "receivable",
	},
	IntentReport: {
		"rekap", "laporan", "statistik", "ringkasan", "summary",
		"pengeluaran", "pemasukan", "minggu ini", "bulan ini",
		"berapa total", "berapa habis", "analisis", "analisa",
	},
	IntentGoal: {
		"target", "tabungan", "nabung", "saving", "goal", "tujuan",
		"impian", "rencana nabung",
	},
	IntentHealth: {
		"skor", "score", "kesehatan keuangan", "financial health",
		"kondisi keuangan", "evaluasi keuangan",
	},
	IntentSmallTalk: {
		"halo", "hai", "hi", "hello", "apa kabar", "selamat pagi",
		"selamat siang", "selamat malam", "terima kasih", "makasih",
		"thanks", "siapa kamu", "apa itu", "kamu bisa apa",
	},
}

// DetectIntent menganalisis pesan user dan mengembalikan intent yang paling relevan.
// Jika tidak ada kata kunci yang cocok, dikembalikan IntentGeneral.
func DetectIntent(message string) ContextIntent {
	msg := strings.ToLower(strings.TrimSpace(message))
	if msg == "" {
		return IntentGeneral
	}

	// Small talk diprioritaskan lebih awal supaya tidak trigger context keuangan.
	for _, kw := range intentKeywords[IntentSmallTalk] {
		if strings.Contains(msg, kw) {
			return IntentSmallTalk
		}
	}

	// Cari intent dengan jumlah keyword match terbanyak untuk akurasi.
	type scored struct {
		intent ContextIntent
		score  int
	}
	scores := make([]scored, 0, 5)

	for intent, keywords := range intentKeywords {
		if intent == IntentSmallTalk {
			continue
		}
		count := 0
		for _, kw := range keywords {
			if strings.Contains(msg, kw) {
				count++
			}
		}
		if count > 0 {
			scores = append(scores, scored{intent, count})
		}
	}

	if len(scores) == 0 {
		return IntentGeneral
	}

	best := scores[0]
	for _, s := range scores[1:] {
		if s.score > best.score {
			best = s
		}
	}
	return best.intent
}

// needsDebtContext mengembalikan true jika intent memerlukan data utang/piutang.
func needsDebtContext(intent ContextIntent) bool {
	return intent == IntentDebt || intent == IntentGeneral
}

// needsGoalContext mengembalikan true jika intent memerlukan data target tabungan.
func needsGoalContext(intent ContextIntent) bool {
	return intent == IntentGoal || intent == IntentGeneral
}

// needsHealthContext mengembalikan true jika intent memerlukan skor kesehatan keuangan.
func needsHealthContext(intent ContextIntent) bool {
	return intent == IntentHealth || intent == IntentGeneral
}

// needsReportContext mengembalikan true jika intent memerlukan ringkasan harian/mingguan.
func needsReportContext(intent ContextIntent) bool {
	return intent == IntentReport || intent == IntentGeneral || intent == IntentTransaction
}

// needsWalletContext mengembalikan true jika intent memerlukan daftar wallet.
// Wallet SELALU disertakan kecuali small talk supaya AI tahu ke mana transaksi disimpan.
func needsWalletContext(intent ContextIntent) bool {
	return intent != IntentSmallTalk
}

// needsDashboardContext mengembalikan true jika intent memerlukan ringkasan dashboard.
func needsDashboardContext(intent ContextIntent) bool {
	return intent != IntentSmallTalk
}
