package repository

import (
	"cuan-backend/internal/entity"
	"fmt"

	"gorm.io/gorm"
)

type TransactionRepository interface {
	Create(transaction *entity.Transaction) error
	FindAll(userID uint, params entity.TransactionFilterParams) ([]entity.Transaction, int64, error)
	FindByID(id uint, userID uint) (*entity.Transaction, error)
	Update(transaction *entity.Transaction) error
	Delete(id uint, userID uint) error
	FindSummaryByDateRange(userID uint, startDate, endDate string, walletID *uint, categoryID *uint, search string) ([]entity.TransactionSummary, error)
	GetCategoryBreakdown(userID uint, startDate, endDate string, walletID *uint, filterType *string) ([]entity.CategoryBreakdown, error)
	GetMonthlyTrend(userID uint, startDate, endDate string) ([]entity.MonthlyTrend, error)
	GetRecentTransactions(userID uint, limit int) ([]entity.Transaction, error)
	// Used for transactions, we need access to DB transaction object
	WithTx(tx *gorm.DB) TransactionRepository
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db}
}

func (r *transactionRepository) WithTx(tx *gorm.DB) TransactionRepository {
	return &transactionRepository{db: tx}
}

func (r *transactionRepository) Create(transaction *entity.Transaction) error {
	return r.db.Create(transaction).Error
}

func (r *transactionRepository) FindAll(userID uint, params entity.TransactionFilterParams) ([]entity.Transaction, int64, error) {
	var transactions []entity.Transaction
	var total int64

	query := r.db.Model(&entity.Transaction{}).Where("transactions.user_id = ?", userID)

	if params.StartDate != "" && params.EndDate != "" {
		query = query.Where("transactions.date BETWEEN ? AND ?", params.StartDate, params.EndDate)
	}
	if params.WalletID != 0 {
		query = query.Where("transactions.wallet_id = ?", params.WalletID)
	}
	if params.CategoryID != 0 {
		query = query.Where("transactions.category_id = ?", params.CategoryID)
	}
	if params.Search != "" {
		query = query.Joins("LEFT JOIN categories ON categories.id = transactions.category_id").
			Where("transactions.description ILIKE ? OR categories.name ILIKE ?", "%"+params.Search+"%", "%"+params.Search+"%")
	}
	if params.Type != "" {
		query = query.Where("transactions.type = ?", params.Type)
	}

	// Count total before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	offset := (params.Page - 1) * params.Limit
	if params.Limit > 0 {
		query = query.Offset(offset).Limit(params.Limit)
	}

	err := query.
		Preload("Wallet").
		Preload("Category").
		Order("date desc, created_at desc").
		Find(&transactions).Error

	return transactions, total, err
}

func (r *transactionRepository) FindByID(id uint, userID uint) (*entity.Transaction, error) {
	var transaction entity.Transaction
	err := r.db.Where("id = ? AND user_id = ?", id, userID).
		Preload("Wallet").
		Preload("Category").
		First(&transaction).Error
	return &transaction, err
}

func (r *transactionRepository) Update(transaction *entity.Transaction) error {
	return r.db.Save(transaction).Error
}

func (r *transactionRepository) Delete(id uint, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&entity.Transaction{}).Error
}

func (r *transactionRepository) FindSummaryByDateRange(userID uint, startDate, endDate string, walletID *uint, categoryID *uint, search string) ([]entity.TransactionSummary, error) {
	var results []entity.TransactionSummary
	
	// Check if single day (naive check: start == end) or small range
    // Ideally user passes granularity, but for now we infer:
    // If startDate == endDate, group by hour
    
	dateFormat := "YYYY-MM-DD"

	// Check if single day (first 10 chars "YYYY-MM-DD" match)
	if len(startDate) >= 10 && len(endDate) >= 10 && startDate[:10] == endDate[:10] {
        // Hourly format: "2024-01-01 10:00"
        dateFormat = "YYYY-MM-DD HH24:00" 
	}
    
	// Use Sprintf for Group clause to ensure identical expression in Select, Group, and Order
	// Postgres requires exact match for Group By to work with Select identifiers
    dateExpr := fmt.Sprintf("TO_CHAR(transactions.date, '%s')", dateFormat)

	query := r.db.Model(&entity.Transaction{}).
		Select(fmt.Sprintf("%s as date, SUM(CASE WHEN transactions.type = 'income' THEN transactions.amount ELSE 0 END) as income, SUM(CASE WHEN transactions.type = 'expense' THEN transactions.amount ELSE 0 END) as expense", dateExpr)).
		Where("transactions.user_id = ? AND transactions.date >= ? AND transactions.date <= ?", userID, startDate, endDate)

	if search != "" {
		query = query.Joins("LEFT JOIN categories ON categories.id = transactions.category_id").
			Where("transactions.description ILIKE ? OR categories.name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if walletID != nil && *walletID != 0 {
		query = query.Where("transactions.wallet_id = ?", *walletID)
	}
	if categoryID != nil && *categoryID != 0 {
		query = query.Where("transactions.category_id = ?", *categoryID)
	}

	err := query.Group(dateExpr).
        Order("1 ASC"). // Order by the first column (date) to avoid ambiguity
		Scan(&results).Error

	return results, err
}

func (r *transactionRepository) GetCategoryBreakdown(userID uint, startDate, endDate string, walletID *uint, filterType *string) ([]entity.CategoryBreakdown, error) {
	results := make([]entity.CategoryBreakdown, 0)

	query := r.db.Table("transactions as t").
		Select("c.name as category_name, c.icon as category_icon, t.type, SUM(t.amount) as total_amount, c.budget_limit").
		Joins("JOIN categories c ON c.id = t.category_id").
		Where("t.user_id = ? AND t.date BETWEEN ? AND ?", userID, startDate, endDate)

	if walletID != nil {
		query = query.Where("t.wallet_id = ?", *walletID)
	}

	if filterType != nil && *filterType != "all" {
		query = query.Where("t.type = ?", *filterType)
	} else {
		query = query.Where("t.type IN (?, ?)", "income", "expense")
	}

	err := query.Group("c.name, c.icon, t.type, c.budget_limit").Scan(&results).Error
	if err != nil {
		return nil, err
	}

    // Calculate IsOverBudget
    for i := range results {
        if results[i].Type == "expense" && results[i].BudgetLimit > 0 {
            if results[i].TotalAmount > results[i].BudgetLimit {
                results[i].IsOverBudget = true
            }
             // Optional: Calculate Percentage if needed in backend, otherwise frontend can do it
             // results[i].Percentage = (results[i].TotalAmount / results[i].BudgetLimit) * 100
        }
    }

	return results, nil
}

func (r *transactionRepository) GetMonthlyTrend(userID uint, startDate, endDate string) ([]entity.MonthlyTrend, error) {
	results := make([]entity.MonthlyTrend, 0)

	// PostgreSQL: TO_CHAR(date, 'YYYY-MM')
	err := r.db.Model(&entity.Transaction{}).
		Select("TO_CHAR(date, 'YYYY-MM') as date, SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END) as income, SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END) as expense").
		Where("user_id = ? AND date BETWEEN ? AND ?", userID, startDate, endDate).
		Group("TO_CHAR(date, 'YYYY-MM')").
		Order("date ASC").
		Scan(&results).Error

	return results, err
}

func (r *transactionRepository) GetRecentTransactions(userID uint, limit int) ([]entity.Transaction, error) {
	var transactions []entity.Transaction
	err := r.db.Where("user_id = ?", userID).
		Preload("Wallet").
		Preload("Category").
		Order("date desc, created_at desc").
		Limit(limit).
		Find(&transactions).Error
	return transactions, err
}
