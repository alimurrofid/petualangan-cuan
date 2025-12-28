package main

import (
	"log"
	"os"

	"cuan-backend/internal/config"
	"cuan-backend/internal/handler"
	"cuan-backend/internal/repository"
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
	if err := godotenv.Load(); err != nil {
		log.Println("Info: No .env file found, relying on system env")
	}

	// Init Config (DB)
	config.Connect()

	// Init Layers
	userRepo := repository.NewUserRepository(config.DB)
	userSvc := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userSvc)

	repo := repository.NewTransactionRepository(config.DB)
	// TransactionService now needs WalletRepo and DB for transaction management
	// We need to initialize WalletRepo first or reuse it. 
	// To keep it clean, let's initialize WalletRepo before TransactionService
	walletRepo := repository.NewWalletRepository(config.DB)
	svc := service.NewTransactionService(repo, walletRepo, config.DB)
	h := handler.NewTransactionHandler(svc)
	
	walletSvc := service.NewWalletService(walletRepo)
	walletHandler := handler.NewWalletHandler(walletSvc)

	categoryRepo := repository.NewCategoryRepository(config.DB)
	categorySvc := service.NewCategoryService(categoryRepo)
	categoryHandler := handler.NewCategoryHandler(categorySvc)

	// Dashboard
	dashboardSvc := service.NewDashboardService(repo, walletRepo)
	dashboardHandler := handler.NewDashboardHandler(dashboardSvc)

	// Init Fiber
	app := fiber.New()
	app.Use(cors.New())

	// Routes
	api := app.Group("/api")
	api.Post("/webhook", h.WebhookReceiver)

	// Protected Routes
	protected := app.Group("/api", middleware.Protected())
	protected.Get("/dashboard", dashboardHandler.GetDashboard) // Dashboard
	protected.Post("/wallets", walletHandler.CreateWallet)
	protected.Get("/wallets", walletHandler.GetWallets)
	protected.Get("/wallets/:id", walletHandler.GetWallet)
	protected.Put("/wallets/:id", walletHandler.UpdateWallet)
	protected.Delete("/wallets/:id", walletHandler.DeleteWallet)

	protected.Post("/categories", categoryHandler.CreateCategory)
	protected.Get("/categories", categoryHandler.GetCategories)
	protected.Get("/categories/:id", categoryHandler.GetCategory)
	protected.Put("/categories/:id", categoryHandler.UpdateCategory)
	protected.Delete("/categories/:id", categoryHandler.DeleteCategory)

	protected.Get("/transactions", h.GetTransactions)
	protected.Post("/transactions", h.CreateTransaction)
	protected.Get("/transactions/calendar", h.GetCalendarData)
	protected.Get("/transactions/report", h.GetReport) // New route for reports
	protected.Post("/transactions/transfer", h.TransferTransaction)
	protected.Delete("/transactions/:id", h.DeleteTransaction)

	// Auth Routes
	auth := app.Group("/auth")
	auth.Post("/register", userHandler.Register)
	auth.Post("/login", userHandler.Login)
	auth.Post("/logout", userHandler.Logout)

	// User Settings Routes (Put under auth for now, or protected route)
	userRoutes := app.Group("/api/user", middleware.Protected())
	userRoutes.Put("/profile", userHandler.UpdateProfile)
	userRoutes.Put("/password", userHandler.ChangePassword)

	// Swagger Route
	app.Get("/swagger/*", swagger.HandlerDefault) 

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(app.Listen(":" + port))
}
