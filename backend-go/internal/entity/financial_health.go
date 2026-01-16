package entity

type FinancialHealthStatus string

const (
	StatusHealthy FinancialHealthStatus = "Sehat"
	StatusWarning FinancialHealthStatus = "Waspada"
	StatusDanger  FinancialHealthStatus = "Bahaya"
)

type FinancialHealthRatio struct {
	Name           string                `json:"name"`
	Value          float64               `json:"value"`           // The calculated ratio value
	FormattedValue string                `json:"formatted_value"` // e.g., "25%" or "4.5 Bulan"
	Target         string                `json:"target"`
	Status         FinancialHealthStatus `json:"status"`
	Description    string                `json:"description"`
}

type FinancialHealthResponse struct {
	OverallScore  float64                `json:"overall_score"` // 0-100
	OverallStatus FinancialHealthStatus  `json:"overall_status"`
	Ratios        []FinancialHealthRatio `json:"ratios"`
}
