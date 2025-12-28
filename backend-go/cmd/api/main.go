package main

import (
	"log"

	"cuan-backend/internal/config"
	"cuan-backend/internal/handler"
	"cuan-backend/internal/repository"
	"cuan-backend/internal/service"

	_ "cuan-backend/docs" // UNCOMMENT THIS AFTER RUNNING swag init

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

	// 1. Init Config (DB)
	config.Connect()

	// 2. Init Layers
	userRepo := repository.NewUserRepository(config.DB)
	userSvc := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userSvc)

	repo := repository.NewTransactionRepository(config.DB)
	svc := service.NewTransactionService(repo)
	h := handler.NewTransactionHandler(svc)

	// 3. Init Fiber
	app := fiber.New()
	app.Use(cors.New())

	// 4. Routes
	api := app.Group("/api")
	api.Get("/transactions", h.GetTransactions)
	api.Post("/webhook", h.WebhookReceiver)

	// Auth Routes
	auth := app.Group("/auth")
	auth.Post("/register", userHandler.Register)
	auth.Post("/login", userHandler.Login)
	auth.Post("/logout", userHandler.Logout)

	// Swagger Route
	app.Get("/swagger/*", swagger.HandlerDefault) 

	log.Fatal(app.Listen(":8080"))
}
