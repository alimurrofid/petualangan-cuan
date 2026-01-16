package main

import (
	"flag"
	"log"
	"os"

	"cuan-backend/internal/config"
	"cuan-backend/internal/handler"
	"cuan-backend/internal/repository"
	"cuan-backend/internal/seeder"
	"cuan-backend/internal/service"
	"cuan-backend/pkg/middleware"

	_ "cuan-backend/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"
)

// @title Petualangan Cuan API
// @version 1.0
// @description API for Petualangan Cuan Application
// @host localhost:8080
// @BasePath /
func main() {
	// Define flags
	freshPtr := flag.Bool("fresh", false, "Drop all tables and re-migrate")
	seedPtr := flag.Bool("seed", false, "Seed database with dummy data")
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Println("Info: No .env file found, relying on system env")
	}

	// Init Config (DB)
	db, err := config.Connect()
	if err != nil {
		log.Fatal(err)
	}

	// Handle Database Workflow Flags
	if *freshPtr {
		config.MigrateFresh(db)
	} else {
		if err := config.RunMigration(db); err != nil {
			log.Fatal(err)
		}
	}

	if *seedPtr {
		seeder.SeedAll(db)
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	// Init Layers
	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userSvc, frontendURL)

	repo := repository.NewTransactionRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	savingGoalRepo := repository.NewSavingGoalRepository(db) // New Repo

	svc := service.NewTransactionService(repo, walletRepo, db)
	h := handler.NewTransactionHandler(svc)
	
	walletSvc := service.NewWalletService(walletRepo, savingGoalRepo) // Updated Injection
	walletHandler := handler.NewWalletHandler(walletSvc)

	categoryRepo := repository.NewCategoryRepository(db)
	categorySvc := service.NewCategoryService(categoryRepo)
	categoryHandler := handler.NewCategoryHandler(categorySvc)

	// Dashboard
	dashboardSvc := service.NewDashboardService(repo, walletRepo)
	dashboardHandler := handler.NewDashboardHandler(dashboardSvc)

	// Debt
	debtRepo := repository.NewDebtRepository(db)
	debtSvc := service.NewDebtService(debtRepo, repo, walletRepo, db)
	debtHandler := handler.NewDebtHandler(debtSvc)
	// Wishlist
	wishlistRepo := repository.NewWishlistRepository(db)
	wishlistSvc := service.NewWishlistService(wishlistRepo)
	wishlistHandler := handler.NewWishlistHandler(wishlistSvc)

	// Saving Goals
	savingGoalSvc := service.NewSavingGoalService(savingGoalRepo, walletRepo, svc, db)
	savingGoalHandler := handler.NewSavingGoalHandler(savingGoalSvc)

	// Init Fiber
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10MB
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     frontendURL,
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, HEAD, PUT, DELETE, PATCH, OPTIONS",
	}))

	api := app.Group("/api")

	// Static Uploads
	app.Static("/uploads", "./uploads")

	api.Post("/webhook", h.WebhookReceiver)

	auth := api.Group("/auth")
	auth.Post("/register", userHandler.Register)
	auth.Post("/login", userHandler.Login)
	auth.Post("/logout", userHandler.Logout)
	auth.Get("/google", userHandler.GoogleLogin)
	auth.Get("/google/callback", userHandler.GoogleCallback)

	// Dashboard
	api.Get("/dashboard", middleware.Protected(), dashboardHandler.GetDashboard)

	// Wallets Group (/api/wallets)
	wallets := api.Group("/wallets", middleware.Protected())
	wallets.Post("/", walletHandler.CreateWallet)
	wallets.Get("/", walletHandler.GetWallets)
	wallets.Get("/:id", walletHandler.GetWallet)
	wallets.Put("/:id", walletHandler.UpdateWallet)
	wallets.Delete("/:id", walletHandler.DeleteWallet)

	// Categories Group (/api/categories)
	categories := api.Group("/categories", middleware.Protected())
	categories.Post("/", categoryHandler.CreateCategory)
	categories.Get("/", categoryHandler.GetCategories)
	categories.Get("/:id", categoryHandler.GetCategory)
	categories.Put("/:id", categoryHandler.UpdateCategory)
	categories.Delete("/:id", categoryHandler.DeleteCategory)

	// Transactions Group (/api/transactions)
	transactions := api.Group("/transactions", middleware.Protected())
	transactions.Get("/", h.GetTransactions)
	transactions.Post("/", h.CreateTransaction)
	transactions.Get("/calendar", h.GetCalendarData)
	transactions.Get("/report/export", h.ExportReport) // BEFORE /report so it doesn't match keys if any, but /report is exact match
	transactions.Get("/report", h.GetReport) 
	transactions.Get("/export", h.ExportTransactions)
	transactions.Post("/transfer", h.TransferTransaction)
	transactions.Get("/:id", h.GetTransaction)
	transactions.Put("/:id", h.UpdateTransaction)
	transactions.Delete("/:id", h.DeleteTransaction)

	// User Settings Routes (/api/user)
	userRoutes := api.Group("/user", middleware.Protected())
	userRoutes.Get("/profile", userHandler.GetProfile)
	userRoutes.Put("/profile", userHandler.UpdateProfile)
	userRoutes.Put("/password", userHandler.ChangePassword)

	// Debt Routes (/api/debts)
	debts := api.Group("/debts", middleware.Protected())
	debts.Post("/", debtHandler.CreateDebt)
	debts.Get("/", debtHandler.GetDebts)
	debts.Get("/:id", debtHandler.GetDebt)
	debts.Post("/:id/pay", debtHandler.PayDebt)
	debts.Put("/:id", debtHandler.UpdateDebt)
	debts.Delete("/:id", debtHandler.DeleteDebt)
	debts.Delete("/payments/:id", debtHandler.DeletePayment)

	// Wishlist Routes (/api/wishlist)
	wishlist := api.Group("/wishlist", middleware.Protected())
	wishlist.Post("/", wishlistHandler.Create)
	wishlist.Get("/", wishlistHandler.FindAll)
	wishlist.Get("/:id", wishlistHandler.FindByID)
	wishlist.Put("/:id", wishlistHandler.Update)
	wishlist.Delete("/:id", wishlistHandler.Delete)
	wishlist.Patch("/:id/bought", wishlistHandler.MarkAsBought)

	// Saving Goals Routes (/api/saving-goals)
	savingGoals := api.Group("/saving-goals", middleware.Protected())
	savingGoals.Get("/", savingGoalHandler.GetGoals)
	savingGoals.Post("/", savingGoalHandler.CreateGoal)
	savingGoals.Post("/:id/contributions", savingGoalHandler.AddContribution)

	// Swagger Route
	app.Get("/swagger/*", swagger.HandlerDefault) 

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(app.Listen(":" + port))
}