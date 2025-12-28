package entity

type CategoryBreakdown struct {
	CategoryName string  `json:"category_name"`
	CategoryIcon string  `json:"category_icon"`
	Type         string  `json:"type"`
	TotalAmount  float64 `json:"total_amount"`
	BudgetLimit  float64 `json:"budget_limit"`
	IsOverBudget bool    `json:"is_over_budget"` // Computed or returned? Computed in backend is safer.
	Percentage   float64 `json:"percentage"`     // Optional, can be computed in frontend or backend
}
