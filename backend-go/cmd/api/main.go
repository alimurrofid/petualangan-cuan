package main

import (
	"log"

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
	svc := service.NewTransactionService(repo)
	h := handler.NewTransactionHandler(svc)

	walletRepo := repository.NewWalletRepository(config.DB)
	walletSvc := service.NewWalletService(walletRepo)
	walletHandler := handler.NewWalletHandler(walletSvc)

	// Init Fiber
	app := fiber.New()
	app.Use(cors.New())

	// Routes
	api := app.Group("/api")
	api.Get("/transactions", h.GetTransactions)
	api.Post("/webhook", h.WebhookReceiver)

	// Protected Routes
	protected := app.Group("/api", middleware.Protected())
	protected.Post("/wallets", walletHandler.CreateWallet)
	protected.Get("/wallets", walletHandler.GetWallets)
	protected.Get("/wallets/:id", walletHandler.GetWallet)
	protected.Put("/wallets/:id", walletHandler.UpdateWallet)
	protected.Delete("/wallets/:id", walletHandler.DeleteWallet)

	// Auth Routes
	auth := app.Group("/auth")
	auth.Post("/register", userHandler.Register)
	auth.Post("/login", userHandler.Login)
	auth.Post("/logout", userHandler.Logout)

	// Swagger Route
	app.Get("/swagger/*", swagger.HandlerDefault) 

	log.Fatal(app.Listen(":8080"))
}
