package entity

type DashboardData struct {
	TotalBalance       float64             `json:"total_balance"`
	TotalIncomeMonth   float64             `json:"total_income_month"`
	TotalExpenseMonth  float64             `json:"total_expense_month"`
	Wallets            []Wallet            `json:"wallets"`
	RecentTransactions []Transaction       `json:"recent_transactions"`
	MonthlyTrend       []MonthlyTrend      `json:"monthly_trend"`
	ExpenseBreakdown   []CategoryBreakdown `json:"expense_breakdown"`
}

type MonthlyTrend struct {
	Date    string  `json:"date"` // YYYY-MM
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
}
