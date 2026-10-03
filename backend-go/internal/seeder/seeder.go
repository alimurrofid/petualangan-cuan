package seeder

import (
	"cuan-backend/internal/entity"
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

		intendedPayday := user.Payday
		if err := db.FirstOrCreate(&user, entity.User{Email: user.Email}).Error; err != nil {
			log.Error().Err(err).Str("email", user.Email).Msg("Failed to seed user")
		} else {
			if intendedPayday != nil && (user.Payday == nil || *user.Payday != *intendedPayday) {
				db.Model(&user).Update("payday", intendedPayday)
				user.Payday = intendedPayday
			}
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

	plantedTxByDesc := make(map[string]uint)
	log.Info().Msg("🌱 Seeding Transactions...")
	if len(plantedWallets) > 0 && len(plantedCategories) > 0 {
		for _, tx := range Transactions {
			tx.UserID = mainUser.ID

			loc, _ := time.LoadLocation("Asia/Jakarta")
			tx.Date = tx.Date.In(loc)

			// MAPPING WALLET: Mengambil ID asli dari database berdasarkan urutan di data.go
			// (tx.WalletID dari data.go adalah index 1-based)
			if tx.WalletID > 0 && int(tx.WalletID) <= len(plantedWallets) {
				tx.WalletID = plantedWallets[tx.WalletID-1].ID
			} else {
				tx.WalletID = plantedWallets[0].ID // fallback aman
			}

			// MAPPING CATEGORY: Mengambil ID asli dari database berdasarkan urutan di data.go
			if tx.CategoryID > 0 && int(tx.CategoryID) <= len(plantedCategories) {
				tx.CategoryID = plantedCategories[tx.CategoryID-1].ID
			} else {
				tx.CategoryID = plantedCategories[0].ID // fallback aman
			}

			if err := db.Create(&tx).Error; err != nil {
				log.Error().Err(err).Str("transaction", tx.Description).Msg("Failed to seed transaction")
			} else {
				plantedTxByDesc[tx.Description] = tx.ID
			}
		}
	}

	var plantedDebts []entity.Debt
	log.Info().Msg("🌱 Seeding Debts...")
	if len(plantedWallets) > 0 {
		for _, debt := range Debts {
			debt.UserID = mainUser.ID
			if debt.WalletID > 0 && int(debt.WalletID) <= len(plantedWallets) {
				debt.WalletID = plantedWallets[debt.WalletID-1].ID
			} else {
				debt.WalletID = plantedWallets[0].ID
			}
			if err := db.Create(&debt).Error; err != nil {
				log.Error().Err(err).Str("debt", debt.Name).Msg("Failed to seed debt")
			} else {
				plantedDebts = append(plantedDebts, debt)
			}
		}
	}

	log.Info().Msg("🌱 Seeding Debt Payments...")
	for _, p := range DebtPayments {
		if p.DebtIndex < 0 || p.DebtIndex >= len(plantedDebts) {
			continue
		}
		txID, hasTx := plantedTxByDesc[p.TransactionDesc]
		if !hasTx {
			log.Warn().Str("tx", p.TransactionDesc).Msg("Transaction not found for debt payment")
			continue
		}

		var wID uint
		if p.WalletIndex > 0 && p.WalletIndex <= len(plantedWallets) {
			wID = plantedWallets[p.WalletIndex-1].ID
		} else {
			wID = plantedWallets[0].ID
		}

		payment := entity.DebtPayment{
			DebtID:        plantedDebts[p.DebtIndex].ID,
			TransactionID: txID,
			WalletID:      wID,
			Amount:        p.Amount,
			Date:          *relativeDate(-p.DaysAgo),
			Note:          p.Note,
		}
		if err := db.Create(&payment).Error; err != nil {
			log.Error().Err(err).Str("note", p.Note).Msg("Failed to seed debt payment")
		}
	}

	var plantedGoals []entity.SavingGoal
	log.Info().Msg("🌱 Seeding Saving Goals...")
	if len(plantedCategories) > 0 {
		for _, goal := range SavingGoals {
			goal.UserID = mainUser.ID
			if goal.CategoryID > 0 && int(goal.CategoryID) <= len(plantedCategories) {
				goal.CategoryID = plantedCategories[goal.CategoryID-1].ID
			} else {
				goal.CategoryID = plantedCategories[0].ID
			}
			if err := db.Create(&goal).Error; err != nil {
				log.Error().Err(err).Str("goal", goal.Name).Msg("Failed to seed saving goal")
			} else {
				plantedGoals = append(plantedGoals, goal)
			}
		}
	}

	log.Info().Msg("🌱 Seeding Saving Contributions...")
	for _, c := range SavingContributions {
		if c.GoalIndex < 0 || c.GoalIndex >= len(plantedGoals) {
			continue
		}
		txID, hasTx := plantedTxByDesc[c.TransactionDesc]
		if !hasTx {
			log.Warn().Str("tx", c.TransactionDesc).Msg("Transaction not found for saving contribution")
			continue
		}

		var wID uint
		if c.WalletIndex > 0 && c.WalletIndex <= len(plantedWallets) {
			wID = plantedWallets[c.WalletIndex-1].ID
		} else {
			wID = plantedWallets[0].ID
		}

		contrib := entity.SavingContribution{
			GoalID:        plantedGoals[c.GoalIndex].ID,
			WalletID:      wID,
			TransactionID: txID,
			Amount:        c.Amount,
			Date:          *relativeDate(-c.DaysAgo),
		}
		if err := db.Create(&contrib).Error; err != nil {
			log.Error().Err(err).Uint("goal_id", contrib.GoalID).Msg("Failed to seed saving contribution")
		}
	}

	log.Info().Msg("🌱 Seeding Wishlists...")
	if len(plantedCategories) > 0 {
		for _, item := range Wishlists {
			item.UserID = mainUser.ID
			if item.CategoryID > 0 && int(item.CategoryID) <= len(plantedCategories) {
				item.CategoryID = plantedCategories[item.CategoryID-1].ID
			} else {
				item.CategoryID = plantedCategories[0].ID
			}
			if err := db.Create(&item).Error; err != nil {
				log.Error().Err(err).Str("wishlist", item.Name).Msg("Failed to seed wishlist item")
			}
		}
	}

	log.Info().Msg("✅ Seeding Finished!")
}
