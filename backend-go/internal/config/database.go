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

	// Set GORM Logger to Silent in production if needed, or default
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

	// Connection Pool Settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	fmt.Println("✅ Terhubung ke Database dengan Connection Pool!")

	// AutoMigrate removed from Connect to allow --fresh flag to work if migration fails
	// fmt.Println("Running Auto Migration...")
	// err = db.AutoMigrate(&entity.Transaction{}, &entity.User{}, &entity.Wallet{}, &entity.Category{}, &entity.Debt{})
	// if err != nil {
	// 	return nil, fmt.Errorf("❌ Gagal migrasi database: %w", err)
	// }

	return db, nil
}

func MigrateFresh(db *gorm.DB) {
	fmt.Println("🚧 Dropping all tables...")
	// Drop tables in reverse dependency order to avoid FK issues
	db.Migrator().DropTable(&entity.WishlistItem{}) // Dependent on User & Category
	db.Migrator().DropTable(&entity.Transaction{})
	db.Migrator().DropTable(&entity.DebtPayment{})
	db.Migrator().DropTable(&entity.Debt{})
	db.Migrator().DropTable(&entity.Category{})
	db.Migrator().DropTable(&entity.Wallet{})
	db.Migrator().DropTable(&entity.User{})

	fmt.Println("✅ All tables dropped!")
	fmt.Println("🆕 Re-running Auto Migration...")
	db.AutoMigrate(&entity.Transaction{}, &entity.User{}, &entity.Wallet{}, &entity.Category{}, &entity.Debt{}, &entity.DebtPayment{}, &entity.WishlistItem{})
}

func RunMigration(db *gorm.DB) error {
	fmt.Println("Running Auto Migration...")
	return db.AutoMigrate(&entity.Transaction{}, &entity.User{}, &entity.Wallet{}, &entity.Category{}, &entity.Debt{}, &entity.DebtPayment{}, &entity.WishlistItem{})
}
