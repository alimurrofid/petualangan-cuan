package seeder

import (
	"cuan-backend/internal/entity"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedAll(db *gorm.DB) {
	fmt.Println("🌱 Seeding Users...")
	plantedUsers := []entity.User{}
	
	// Create Users
	for _, user := range Users {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal("Failed to hash password")
		}
		user.Password = string(hashedPassword)
		
		if err := db.FirstOrCreate(&user, entity.User{Email: user.Email}).Error; err != nil {
			log.Printf("Failed to seed user %s: %v\n", user.Email, err)
		} else {
			plantedUsers = append(plantedUsers, user)
		}
	}

	if len(plantedUsers) == 0 {
		log.Println("⚠️ No users seeded/found to attach data to")
		return
	}
	
	mainUser := plantedUsers[0] // Use the first user for relational data
	fmt.Printf("👤 Using user: %s (ID: %d) for related data\n", mainUser.Name, mainUser.ID)

	fmt.Println("🌱 Seeding Wallets...")
	plantedWallets := []entity.Wallet{}
	for _, wallet := range Wallets {
		wallet.UserID = mainUser.ID
		if err := db.Create(&wallet).Error; err != nil {
			log.Printf("Failed to seed wallet %s: %v\n", wallet.Name, err)
		} else {
			plantedWallets = append(plantedWallets, wallet)
		}
	}

	fmt.Println("🌱 Seeding Categories...")
	plantedCategories := []entity.Category{}
	for _, cat := range Categories {
		cat.UserID = mainUser.ID
		if err := db.Create(&cat).Error; err != nil {
			log.Printf("Failed to seed category %s: %v\n", cat.Name, err)
		} else {
			plantedCategories = append(plantedCategories, cat)
		}
	}

	fmt.Println("🌱 Seeding Transactions...")
	if len(plantedWallets) > 0 && len(plantedCategories) > 0 {
		for i, tx := range Transactions {
			tx.UserID = mainUser.ID
			
			// Distribute transactions across wallets and categories (simple round-robin or random)
			tx.WalletID = plantedWallets[i%len(plantedWallets)].ID
			
			// Matches type (income/expense)
			for _, cat := range plantedCategories {
				if cat.Type == tx.Type {
					tx.CategoryID = cat.ID
					break
				}
			}
			
			if err := db.Create(&tx).Error; err != nil {
				log.Printf("Failed to seed transaction %s: %v\n", tx.Description, err)
			}
		}
	}

	fmt.Println("✅ Seeding Finished!")
}
