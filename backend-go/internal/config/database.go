package config

import (
	"fmt"
	"os"
	"time"

	"cuan-backend/internal/entity"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("❌ Gagal connect ke Database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("❌ Gagal mendapatkan instance sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	fmt.Println("✅ Terhubung ke Database dengan Connection Pool!")

	return db, nil
}

func MigrateFresh(db *gorm.DB) {
	fmt.Println("🚧 Dropping all tables...")
	db.Migrator().DropTable(&entity.SavingContribution{})
	db.Migrator().DropTable(&entity.SavingGoal{})
	db.Migrator().DropTable(&entity.WishlistItem{})
	db.Migrator().DropTable(&entity.Transaction{})
	db.Migrator().DropTable(&entity.DebtPayment{})
	db.Migrator().DropTable(&entity.Debt{})
	db.Migrator().DropTable(&entity.Category{})
	db.Migrator().DropTable(&entity.Wallet{})
	db.Migrator().DropTable(&entity.User{})

	fmt.Println("✅ All tables dropped!")
	fmt.Println("🆕 Re-running Auto Migration...")
	db.AutoMigrate(&entity.Transaction{}, &entity.User{}, &entity.Wallet{}, &entity.Category{}, &entity.Debt{}, &entity.DebtPayment{}, &entity.WishlistItem{}, &entity.SavingGoal{}, &entity.SavingContribution{})
}

func RunMigration(db *gorm.DB) error {
	fmt.Println("Running Auto Migration...")
	return db.AutoMigrate(&entity.Transaction{}, &entity.User{}, &entity.Wallet{}, &entity.Category{}, &entity.Debt{}, &entity.DebtPayment{}, &entity.WishlistItem{}, &entity.SavingGoal{}, &entity.SavingContribution{})
}
