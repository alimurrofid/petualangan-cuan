package entity

type DashboardData struct {
	TotalBalance          float64             `json:"total_balance"`
	TotalAvailableBalance float64             `json:"total_available_balance"`
	TotalIncomeMonth      float64             `json:"total_income_month"`
	TotalExpenseMonth     float64             `json:"total_expense_month"`
	Wallets               []Wallet            `json:"wallets"`
	RecentTransactions    []Transaction       `json:"recent_transactions"`
	MonthlyTrend          []MonthlyTrend      `json:"monthly_trend"`
	ExpenseBreakdown      []CategoryBreakdown `json:"expense_breakdown"`
	PaceMetric            *PaceMetric         `json:"pace_metric"`
}

type PaceMetric struct {
	TotalBudget        float64 `json:"total_budget"`
	RemainingBudget    float64 `json:"remaining_budget"`
	DaysRemaining      int     `json:"days_remaining"`
	DailySafeToSpend   float64 `json:"daily_safe_to_spend"`
	TodayExpense       float64 `json:"today_expense"`
	TodayRemainingSafe float64 `json:"today_remaining_safe"`
	Status             string  `json:"status"` // "SAFE", "WARNING", "OVERPACE"
}

type MonthlyTrend struct {
	Date    string  `json:"date"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
}

