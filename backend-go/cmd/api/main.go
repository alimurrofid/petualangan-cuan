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
	svc := service.NewTransactionService(repo, walletRepo, db)
	h := handler.NewTransactionHandler(svc)
	
	walletSvc := service.NewWalletService(walletRepo)
	walletHandler := handler.NewWalletHandler(walletSvc)

	categoryRepo := repository.NewCategoryRepository(db)
	categorySvc := service.NewCategoryService(categoryRepo)
	categoryHandler := handler.NewCategoryHandler(categorySvc)

	// Dashboard
	dashboardSvc := service.NewDashboardService(repo, walletRepo)
	dashboardHandler := handler.NewDashboardHandler(dashboardSvc)

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
	transactions.Get("/report", h.GetReport) 
	transactions.Post("/transfer", h.TransferTransaction)
	transactions.Get("/:id", h.GetTransaction)
	transactions.Put("/:id", h.UpdateTransaction)
	transactions.Delete("/:id", h.DeleteTransaction)

	// User Settings Routes (/api/user)
	userRoutes := api.Group("/user", middleware.Protected())
	userRoutes.Get("/profile", userHandler.GetProfile)
	userRoutes.Put("/profile", userHandler.UpdateProfile)
	userRoutes.Put("/password", userHandler.ChangePassword)

	// Swagger Route
	app.Get("/swagger/*", swagger.HandlerDefault) 

    // Static Uploads
    app.Static("/uploads", "./uploads")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(app.Listen(":" + port))
}