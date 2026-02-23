package seeder

import (
	"cuan-backend/internal/entity"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedAll(db *gorm.DB) {
	log.Info().Msg("🌱 Seeding Users...")
	plantedUsers := []entity.User{}
	
	// Create Users
	for _, user := range Users {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal().Msg("Failed to hash password")
		}
		user.Password = string(hashedPassword)
		
		if err := db.FirstOrCreate(&user, entity.User{Email: user.Email}).Error; err != nil {
			log.Error().Err(err).Str("email", user.Email).Msg("Failed to seed user")
		} else {
			plantedUsers = append(plantedUsers, user)
		}
	}

	if len(plantedUsers) == 0 {
		log.Warn().Msg("⚠️ No users seeded/found to attach data to")
		return
	}
	
	mainUser := plantedUsers[0]
	log.Info().Str("name", mainUser.Name).Uint("id", mainUser.ID).Msg("👤 Using user for related data")

	log.Info().Msg("🌱 Seeding Wallets...")
	plantedWallets := []entity.Wallet{}
	for _, wallet := range Wallets {
		wallet.UserID = mainUser.ID
		if err := db.Create(&wallet).Error; err != nil {
			log.Error().Err(err).Str("wallet", wallet.Name).Msg("Failed to seed wallet")
		} else {
			plantedWallets = append(plantedWallets, wallet)
		}
	}

	log.Info().Msg("🌱 Seeding Categories...")
	plantedCategories := []entity.Category{}
	for _, cat := range Categories {
		cat.UserID = mainUser.ID
		if err := db.Create(&cat).Error; err != nil {
			log.Error().Err(err).Str("category", cat.Name).Msg("Failed to seed category")
		} else {
			plantedCategories = append(plantedCategories, cat)
		}
	}

	log.Info().Msg("🌱 Seeding Transactions...")
	if len(plantedWallets) > 0 && len(plantedCategories) > 0 {
		for i, tx := range Transactions {
			tx.UserID = mainUser.ID
			
			loc, _ := time.LoadLocation("Asia/Jakarta")
			tx.Date = tx.Date.In(loc)

			tx.WalletID = plantedWallets[i%len(plantedWallets)].ID
			
			var validCategories []entity.Category
			for _, cat := range plantedCategories {
				if cat.Type == tx.Type {
					validCategories = append(validCategories, cat)
				}
			}

			if len(validCategories) > 0 {
				matched := false
				for _, cat := range validCategories {
					if strings.Contains(strings.ToLower(tx.Description), strings.ToLower(strings.Split(cat.Name, " ")[0])) {
						tx.CategoryID = cat.ID
						matched = true
						break
					}
				}
				
				if !matched {
					tx.CategoryID = validCategories[i%len(validCategories)].ID
				}
			}
			
			if err := db.Create(&tx).Error; err != nil {
				log.Error().Err(err).Str("transaction", tx.Description).Msg("Failed to seed transaction")
			}
		}
	}

	log.Info().Msg("✅ Seeding Finished!")
}
